package service

import (
	"encoding/hex"
	"gproxy/config"
	"gproxy/model"
	"gproxy/wallet"
	"log/slog"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	iotago "github.com/iotaledger/iota.go/v3"
)

type CacheOutput struct {
	output   iotago.Output
	outputID iotago.OutputID
}

var CreateProxyPoolSignal chan bool
var MintNameNFTSignal chan bool
var cacheOutputs map[string]CacheOutput
var cacheOutputMu sync.RWMutex
var cacheNFT *iotago.NFTOutput
var cacheNFTID iotago.OutputID
var cacheNFTMu sync.RWMutex

func init() {
	cacheOutputs = make(map[string]CacheOutput)
	CreateProxyPoolSignal = make(chan bool, 10)
	MintNameNFTSignal = make(chan bool)
}

func GetCacheOutput(addr string) (iotago.Output, iotago.OutputID) {
	cacheOutputMu.RLock()
	defer cacheOutputMu.RUnlock()
	if c, exist := cacheOutputs[addr]; exist {
		delete(cacheOutputs, addr)
		return c.output, c.outputID
	}
	return nil, iotago.OutputID{}
}

func GetCacheNFT() (*iotago.NFTOutput, iotago.OutputID) {
	cacheNFTMu.RLock()
	defer cacheNFTMu.RUnlock()
	nft := cacheNFT
	cacheNFT = nil
	return nft, cacheNFTID
}

func StartUpdateCacheOutputs() {
	updateProxyPoolCacheOutputs()
	if len(config.NameNftId) > 0 {
		updateMintNameNftCacheOutputs()
	}
	for {
		select {
		case <-CreateProxyPoolSignal:
			updateProxyPoolCacheOutputs()
		case <-MintNameNFTSignal:
			updateMintNameNftCacheOutputs()
		}
	}
}

func updateProxyPoolCacheOutputs() {
	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, "", "", "")

	// Get addresses which are confirmed
	addrs, err := model.GetProxyPool(model.CONFIRMED_SEND)
	if err != nil {
		slog.Error("model.GetProxyPool", "type", model.CONFIRMED_SEND, "err", err)
		return
	}
	if len(addrs) == 0 {
		return
	}

	for addr := range addrs {
		time.Sleep(time.Second)
		output, id, err := w.GetUnspentOutput(addr)
		if err != nil {
			slog.Error("w.GetUnspentOutput error. %s, %v", addr, err)
			continue
		}
		slog.Info("cache output", "addr", addr, "id", hex.EncodeToString(id[:]))

		cacheOutputMu.Lock()
		cacheOutputs[addr] = CacheOutput{output: output, outputID: id}
		cacheOutputMu.Unlock()
	}
}

func updateMintNameNftCacheOutputs() {
	addr, _, err := model.GetIssuerByNftid(config.NameNftId)
	if err != nil {
		slog.Error("model.GetIssuerByNftid", "namenftid", config.NameNftId, "err", err)
		return
	}
	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, "", "", config.NameNftId)

	output, id, err := w.GetUnspentOutput(addr)
	if err != nil {
		slog.Error("w.GetUnspentOutput", "addr", addr, "err", err)
	} else {
		slog.Info("cache output", "addr", addr, "id", hexutil.Encode(id[:]))
		cacheOutputMu.Lock()
		cacheOutputs[addr] = CacheOutput{output: output, outputID: id}
		cacheOutputMu.Unlock()
	}

	nft, id, err := w.GetCollectionNFTOutput()
	if err != nil {
		slog.Error("w.GetCollectionNFTOutput", "namenftid", config.NameNftId, "err", err)
	} else {
		slog.Info("cache nft", "id", hexutil.Encode(id[:]))
		cacheNFTMu.Lock()
		cacheNFT = nft
		cacheNFTID = id
		cacheNFTMu.Unlock()
	}
}
