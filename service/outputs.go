package service

import (
	"encoding/hex"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/wallet"
	"sync"
	"time"

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
		gl.OutLogger.Error("model.GetProxyPool error. %v, %d", err, model.CONFIRMED_SEND)
		return
	}
	if len(addrs) == 0 {
		return
	}

	for addr := range addrs {
		time.Sleep(time.Second)
		output, id, err := w.GetUnspentOutput(addr)
		if err != nil {
			gl.OutLogger.Error("w.GetUnspentOutput error. %s, %v", addr, err)
			continue
		}
		gl.OutLogger.Info("cache output %s : 0x%s", addr, hex.EncodeToString(id[:]))

		cacheOutputMu.Lock()
		cacheOutputs[addr] = CacheOutput{output: output, outputID: id}
		cacheOutputMu.Unlock()
	}
}

func updateMintNameNftCacheOutputs() {
	addr, _, err := model.GetIssuerByNftid(config.NameNftId)
	if err != nil {
		gl.OutLogger.Error("model.GetIssuerByNftid error. %s, %v", config.NameNftId, err)
		return
	}
	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, "", "", config.NameNftId)

	output, id, err := w.GetUnspentOutput(addr)
	if err != nil {
		gl.OutLogger.Error("w.GetUnspentOutput error. %s, %v", addr, err)
	} else {
		gl.OutLogger.Info("cache output %s : 0x%s", addr, hex.EncodeToString(id[:]))
		cacheOutputMu.Lock()
		cacheOutputs[addr] = CacheOutput{output: output, outputID: id}
		cacheOutputMu.Unlock()
	}

	nft, id, err := w.GetCollectionNFTOutput()
	if err != nil {
		gl.OutLogger.Error("w.GetCollectionNFTOutput error. %s, %v", config.NameNftId, err)
	} else {
		gl.OutLogger.Info("cache nft 0x%s", hex.EncodeToString(id[:]))
		cacheNFTMu.Lock()
		cacheNFT = nft
		cacheNFTID = id
		cacheNFTMu.Unlock()
	}
}
