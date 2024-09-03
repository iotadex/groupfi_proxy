package service

import (
	"gproxy/config"
	"gproxy/model"
	"gproxy/wallet"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var buySmrSignal chan bool

func RunSendSmr() {
	f := func(CurrentSentTs *int64) {
		// Get a record from database
		o, err := model.PopOnePendingSendSmrOrder()
		if err != nil {
			slog.Error("model.PopOnePendingSendSmrOrder", "err", err)
			return
		}
		if o == nil {
			return
		}

		// check the ts
		if o.Ts <= *CurrentSentTs {
			slog.Error("send_coin_pending id", "ts", CurrentSentTs, "order", *o)
			return
		}
		*CurrentSentTs = o.Ts

		// get the wallet
		w, err := getWallet(config.ProxyWallet)
		if err != nil {
			slog.Error("getWallet", "order", *o, "err", err)
			return
		}

		blockId, err := w.SendBasic(o.To, o.Amount)
		if err != nil {
			slog.Error("w.SendBasic", "order", *o, "err", err)
			// store back
			if err = model.StoreBackPendingSendSmrOrder(o.To, o.Amount, o.Type); err != nil {
				slog.Error("model.StoreBackPendingSendSmrOrder", "order", *o, "err", err)
			}
			return
		}

		// updata blockid and state
		if err = model.UpdatePendingOrderblockId(o.Id, hexutil.Encode(blockId)); err != nil {
			slog.Error("model.UpdatePendingOrderblockId", "order", *o, "err", err)
			return
		}

		time.Sleep(time.Second * 30)
	}
	buySmrSignal = make(chan bool, 10)
	go runCheckPendingOrders()
	CurrentSentTs := int64(-1)
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-buySmrSignal:
			f(&CurrentSentTs)
		case <-ticker.C:
			f(&CurrentSentTs)
		}
	}
}

func runCheckPendingOrders() {
	ticker := time.NewTicker(time.Second * 10)
	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, "", "", "")
	for range ticker.C {
		orders, err := model.GetPendingOrders()
		if err != nil {
			slog.Error("model.GetPendingOrders", "err", err)
			continue
		}

		for _, o := range orders {
			time.Sleep(time.Second)
			b, err := w.CheckTx(common.FromHex(o.BlockId))
			if b && err != nil {
				slog.Error("w.CheckTx", "blockid", o.BlockId, "err", err)
				continue
			}

			state := model.CONFIRMED_SEND
			if !b {
				state = model.FAILED_SEND
			}
			if err = model.UpdatePendingOrderState(o.Id, state); err != nil {
				slog.Error("model.UpdatePendingOrderState", "id", o.Id, "err", err)
			}
			if o.Type == model.SEND_POOL {
				if err = model.UpdateProxyPoolState(o.To, state); err != nil {
					slog.Error("model.UpdateProxyPoolState", "id", o.Id, "err", err)
				}
			}
		}
	}
}

func getWallet(nftid string) (*wallet.IotaSmrWallet, error) {
	// Get wallet
	addr, enpk, err := model.GetWallet(nftid)
	if err != nil {
		return nil, err
	}

	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, addr, enpk, "0x0")
	return w, nil
}
