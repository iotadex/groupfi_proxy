package service

import (
	"gproxy/config"
	"gproxy/model"
	"gproxy/wallet"
	"log"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

var mintNameSignal chan bool

func RunMintNameNft() {
	mintNameSignal = make(chan bool, 1)
	f := func(addr, pk string, preMintTs *int64) {
		if (time.Now().Unix() - *preMintTs) < config.SendIntervalTime {
			return
		}
		r, err := model.PopOneNameNftRecord()
		if err != nil {
			slog.Error("model.PopOneNftNameRecord", "err", err)
			return
		}
		if r == nil {
			return
		}

		basicOutput, basicId := GetCacheOutput(addr)
		nftOutput, nftOutputId := GetCacheNFT()

		node := GetEnableHornetNode()
		if node == nil || node.Info == nil {
			slog.Error("mint name nft. There is no healthy hornet node")
			return
		}
		w := wallet.NewIotaSmrWallet(node.Url, addr, pk, config.NameNftId)

		if id, err := w.MintNameNFT(r.To, r.Expire, r.Meta, []byte(config.NameNftTag), basicOutput, basicId, nftOutput, nftOutputId, node.Info); err != nil {
			slog.Error("sq.w.MintNameNFT", "to", r.To, "err", err)
		} else {
			if err = model.UpdateBlockIdToNameNftRecord(r.Nftid, hexutil.Encode(id)); err != nil {
				slog.Error("model.UpdateBlockIdToNameNftRecord", "nftid", r.Nftid, "id", hexutil.Encode(id), "err", err)
			}
			*preMintTs = time.Now().Unix()
			go checkNameNft(w, r.Nftid, r.To, id)
		}
	}

	addr, pk, err := model.GetIssuerByNftid(config.NameNftId)
	if err != nil {
		log.Panicf("model.GetIssuerByNftid error. %s, %v", config.NameNftId, err)
	}
	ticker := time.NewTicker(time.Second * time.Duration(config.SendIntervalTime))
	preMintTs := int64(0)
	for {
		select {
		case <-mintNameSignal:
			f(addr, pk, &preMintTs)
		case <-ticker.C:
			f(addr, pk, &preMintTs)
		}
	}
}

func MintNameNftSignal() {
	mintNameSignal <- true
}

func checkNameNft(w *wallet.IotaSmrWallet, id, addr string, blockId []byte) {
	time.Sleep(time.Minute)
	nftid, err := w.GetNftOutputFromBlockID(blockId)
	if err != nil {
		slog.Error("w.GetNftOutputFromBlockID", "blockid", hexutil.Encode(blockId), "err", err)
		return
	}
	slog.Info("Mint name nft", "nftid", nftid, "addr", addr)
	MintNameNFTSignal <- true

	if err := model.UpdateNameNft(id, nftid); err != nil {
		slog.Error("model.StoreNameNft", "nftid", nftid, "blockid", hexutil.Encode(blockId), "err", err)
	}
}
