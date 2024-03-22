package service

import (
	"encoding/hex"
	"encoding/json"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/wallet"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func Start() {
	go RunProxyPool()
}

func RunMintNameNft() {
	addr, pk, err := model.GetIssuerByNftid(config.NameNftId)
	if err != nil {
		log.Panicf("model.GetIssuerByNftid error. %s, %v", config.NameNftId, err)
	}
	w := wallet.NewIotaSmrWallet(config.SmrRpc, addr, pk, config.NameNftId)
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
	addr, pk, err := model.GetIssuerByNftid(config.PkNftId)
	if err != nil {
		log.Panicf("model.GetIssuerByNftid error. %s, %v", config.PkNftId, err)
	}
	w := wallet.NewIotaSmrWallet(config.SmrRpc, addr, pk, config.PkNftId)
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
		w := wallet.NewIotaSmrWallet(config.SmrRpc, config.MainWallet, config.MainWalletPk, "0x0")
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
}
