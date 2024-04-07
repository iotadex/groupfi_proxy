package service

import (
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/wallet"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func RunSendSmr() {
	go func() {
		CurrentSentTs := int64(-1)
		ticker := time.NewTicker(time.Second * 10)
		for range ticker.C {
			// Get a record from database
			o, err := model.PopOnePendingSendSmrOrder()
			if err != nil {
				gl.OutLogger.Error("model.PopOnePendingSendSmrOrder error. %v", err)
				continue
			}
			if o == nil {
				continue
			}

			// check the ts
			if o.Ts <= CurrentSentTs {
				gl.OutLogger.Error("send_coin_pending id error. %d : %v", CurrentSentTs, *o)
				continue
			}
			CurrentSentTs = o.Ts

			// get the wallet
			w, err := getWallet(config.ProxyWallet)
			if err != nil {
				gl.OutLogger.Error("getWallet error. %v, %v", *o, err)
				continue
			}

			blockId, err := w.SendBasic(o.To, o.Amount)
			if err != nil {
				gl.OutLogger.Error("w.SendBasic error. %v, %v", *o, err)
				// store back
				if err = model.StoreBackPendingSendSmrOrder(o.To, o.Amount, o.Type); err != nil {
					gl.OutLogger.Error("model.StoreBackPendingSendSmrOrder error. %v, %v", *o, err)
				}
				continue
			}

			// updata blockid and state
			if err = model.UpdatePendingOrderblockId(o.Id, hexutil.Encode(blockId)); err != nil {
				gl.OutLogger.Error("model.UpdatePendingOrderblockId error. %v, %v", *o, err)
				continue
			}

			time.Sleep(time.Second * 30)
		}
	}()

	go runCheckPendingOrders()
}

func runCheckPendingOrders() {
	ticker := time.NewTicker(time.Second * 10)
	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, "", "", "")
	for range ticker.C {
		orders, err := model.GetPendingOrders()
		if err != nil {
			gl.OutLogger.Error("model.GetPendingOrders error. %v", err)
			continue
		}

		for _, o := range orders {
			time.Sleep(time.Second)
			b, err := w.CheckTx(common.FromHex(o.BlockId))
			if b && err != nil {
				gl.OutLogger.Error("w.CheckTx error. %s, %v", o.BlockId, err)
				continue
			}

			state := 2
			if !b {
				state = 3
			}
			if err = model.UpdatePendingOrderState(o.Id, state); err != nil {
				gl.OutLogger.Error("model.UpdatePendingOrderState. %d, %v", o.Id, err)
			}
			if o.Type == model.SEND_POOL {
				if err = model.UpdateProxyPoolState(o.To, state); err != nil {
					gl.OutLogger.Error("model.UpdateProxyPoolState. %d, %v", o.Id, err)
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
