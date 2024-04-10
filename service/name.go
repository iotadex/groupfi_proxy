package service

import (
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
	ticker := time.NewTicker(time.Second * time.Duration(config.SendIntervalTime))
	for range ticker.C {
		r, err := model.PopOneNameNftRecord()
		if err != nil {
			gl.OutLogger.Error("model.PopOneNftNameRecord error. %v", err)
			continue
		}

		if id, err := w.MintNameNFT(r.To, r.Expire, r.Meta, []byte(config.NameNftTag)); err != nil {
			gl.OutLogger.Error("sq.w.MintNameNFT error. %s, %v", r.To, err)
		} else {
			if err = model.UpdateBlockIdToNameNftRecord(r.Nftid, hexutil.Encode(id)); err != nil {
				gl.OutLogger.Error("model.UpdateBlockIdToNameNftRecord error.%s : %s : %v", r.Nftid, hexutil.Encode(id), err)
			}
			go checkNameNft(w, r.Nftid, r.To, id)
			time.Sleep(time.Second * 30)
		}
	}
}

func checkNameNft(w *wallet.IotaSmrWallet, id, addr string, blockId []byte) {
	time.Sleep(time.Minute)
	nftid, err := w.GetNftOutputFromBlockID(blockId)
	if err != nil {
		gl.OutLogger.Error("w.GetNftOutputFromBlockID error. %s, %v", hexutil.Encode(blockId), err)
		return
	}
	gl.OutLogger.Info("Mint name nft %s, %s", nftid, addr)

	if err := model.UpdateNameNft(id, nftid); err != nil {
		gl.OutLogger.Error("model.StoreNameNft error. %s, %s, %v", nftid, hexutil.Encode(blockId), err)
	}
}
