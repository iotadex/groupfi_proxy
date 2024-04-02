package service

import (
	"encoding/json"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/wallet"
	"log"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func Start() {
}

func RunMintNameNft() {
	addr, pk, err := model.GetIssuerByNftid(config.NameNftId)
	if err != nil {
		log.Panicf("model.GetIssuerByNftid error. %s, %v", config.NameNftId, err)
	}
	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, addr, pk, config.NameNftId)
	mintNameNftQueue = NewSendQueue()
	q := mintNameNftQueue
	ticker := time.NewTicker(time.Second * time.Duration(config.SendIntervalTime))
	q.status.Store(true)
	for range ticker.C {
		if !q.status.Load() {
			return
		}
		msg := q.pop()
		if msg == nil {
			continue
		}
		if id, err := w.MintNameNFT(msg.Addr, msg.ExpireDays, msg.NftMeta, msg.NftTag); err != nil {
			gl.OutLogger.Error("sq.w.MintNameNFT error. %s, %v", msg.Addr, err)
			//push the meta to the queue's front position
			q.pushFront(msg)
		} else {
			data := make(map[string]string)
			json.Unmarshal(msg.NftMeta, &data)
			go checkNameNft(w, msg.Addr, data["name"], id)
			time.Sleep(time.Second * 30)
		}
	}
}

func checkNameNft(w *wallet.IotaSmrWallet, addr, name string, blockId []byte) {
	time.Sleep(time.Minute)
	nftid, err := w.GetNftOutputFromBlockID(blockId)
	if err != nil {
		gl.OutLogger.Error("w.GetNftOutputFromBlockID error. %s, %v", hexutil.Encode(blockId), err)
		return
	}
	gl.OutLogger.Info("Mint name nft %s, %s, %s", nftid, addr, name)

	if err := model.UpdateNameNft(nftid, hexutil.Encode(blockId), name); err != nil {
		gl.OutLogger.Error("model.StoreNameNft error. %s, %s, %s, %v", nftid, hexutil.Encode(blockId), name, err)
	}
}

func RunMintPkNft() {
	addr, pk, err := model.GetIssuerByNftid(config.ProxyWalletNftId)
	if err != nil {
		log.Panicf("model.GetIssuerByNftid error. %s, %v", config.ProxyWalletNftId, err)
	}
	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, addr, pk, config.ProxyWalletNftId)
	mintPkNftQueue = NewSendQueue()
	q := mintPkNftQueue
	ticker := time.NewTicker(time.Second * time.Duration(config.SendIntervalTime))
	q.status.Store(true)
	for range ticker.C {
		if !q.status.Load() {
			return
		}
		msg := q.pop()
		if msg == nil {
			continue
		}
		if id, err := w.MintPkNFT(msg.Addr, msg.NftMeta, msg.NftTag); err != nil {
			gl.OutLogger.Error("sq.w.MintPkNFT error. %s, %v", msg.Addr, err)
			//push the meta to the queue's front position
			q.pushFront(msg)
		} else {
			go checkPkNft(w, msg.Addr, id)
			time.Sleep(time.Second * 30)
		}
	}
}

func checkPkNft(w *wallet.IotaSmrWallet, addr string, blockId []byte) {
	time.Sleep(time.Minute)
	nftid, err := w.GetNftOutputFromBlockID(blockId)
	if err != nil {
		gl.OutLogger.Error("w.GetNftOutputFromBlockID error. %s, %v", hexutil.Encode(blockId), err)
		return
	}
	gl.OutLogger.Info("Mint pk nft %s, %s", nftid, addr)
}

// run a thread to keep the proxy pool full
/*
func RunProxyPool() {
	f := func() {
		// 1. Get the count of proxy pool; check the count smaller than required or not
		count, err := model.CountPoolProxies()
		if err != nil {
			gl.OutLogger.Error("model.CountPoolProxies error. %v", err)
			return
		}
		if count < config.MinProxyPoolCount {
			for i := config.MinProxyPoolCount - count; i > -1; i-- {
				addr, err := model.CreateProxyToPool()
				if err != nil {
					gl.OutLogger.Error("model.CreateProxyToPool error. %v", err)
					return
				}
				gl.OutLogger.Info("Create a proxy to pool. %s", addr)
			}
		}

		// 2. Get the uninitialized proxies from pool, and send a number to them
		addresses, err := model.GetUninitializedProxiesFromPool()
		if err != nil {
			gl.OutLogger.Error("model.GetUninitializedProxiesFromPool error. %v", err)
			return
		}
		w := wallet.NewIotaSmrWallet(config.ShimmerRpc, config.MainWallet, config.MainWalletPk, "0x0")
		ids := make([][]byte, len(addresses))
		for i, addr := range addresses {
			id, err := w.SendBasic(addr, config.ProxySendAmount)
			if err != nil {
				gl.OutLogger.Error("w.SendBasic error. %s, %v", addr, err)
				continue
			}
			ids[i] = id
			gl.OutLogger.Info("Send smr to %s. block_id : 0x%s", addr, hex.EncodeToString(id))
			time.Sleep(30 * time.Second)
		}
		time.Sleep(time.Minute)

		// 3. check the proxies of pool and update their state
		for i, id := range ids {
			b, err := w.CheckTx(id)
			if err != nil {
				gl.OutLogger.Error("w.CheckTx error. 0x%s, %v", hex.EncodeToString(id), err)
				continue
			}
			if b {
				if err = model.InitPoolProxy(addresses[i]); err != nil {
					gl.OutLogger.Error("model.InitPoolProxy error. %s, %v", addresses[i], err)
				}
			}
			time.Sleep(time.Second)
		}
	}
	f()
	ticker := time.NewTicker(time.Hour * time.Duration(config.ProxyPoolCheckHours))
	for range ticker.C {
		f()
	}
}*/

func RunSendSmr() {
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
		w, err := getWallet(strconv.Itoa(int(o.Type)))
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
}

func RunCheckPendingOrders() {
	ticker := time.NewTicker(time.Second * 10)
	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, "", "", "")
	for range ticker.C {
		orders, err := model.GetPendingOrders()
		if err != nil {
			gl.OutLogger.Error("model.GetPendingOrders error. %v", err)
			continue
		}

		for id, blockid := range orders {
			time.Sleep(time.Second)
			b, err := w.CheckTx(common.FromHex(blockid))
			if b && err != nil {
				gl.OutLogger.Error("w.CheckTx error. %s, %v", blockid, err)
				continue
			}

			state := 2
			if !b {
				state = 3
			}
			if err = model.UpdatePendingOrderState(id, state); err != nil {
				gl.OutLogger.Error("model.UpdatePendingOrderState. %d, %v", id, err)
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
