package service

import (
	"encoding/json"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/wallet"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

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
