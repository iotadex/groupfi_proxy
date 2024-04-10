package service

import (
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/tools"
	"gproxy/wallet"
	"time"

	"github.com/ethereum/go-ethereum/common"
	iotago "github.com/iotaledger/iota.go/v3"
)

func RunKeepProxyPoolFull() {
	f := func() {
		if err := model.CreateProxyToPool(config.ProxySendAmount, config.MinProxyPoolCount); err != nil {
			gl.OutLogger.Error("model.CreateProxyToPool error. %v", err)
		}
	}
	f()
	ticker := time.NewTicker(time.Hour * time.Duration(config.ProxyPoolCheckHours))
	for range ticker.C {
		f()
	}
}

func RunCheckProxyPoolBalance() {
	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, "", "", "")
	ticker := time.NewTicker(time.Hour * time.Duration(config.ProxyPoolCheckHours))
	for range ticker.C {
		addrs, err := model.GetUsedProxyPool()
		if err != nil {
			gl.OutLogger.Error("model.GetUsedProxyPool error. %v", err)
		}

		for bech32Addr, enpk := range addrs {
			time.Sleep(time.Second)
			pk := common.FromHex(string(tools.Aes.GetDecryptString(enpk, seeds)))
			addr := iotago.Ed25519AddressFromPubKey(pk[32:])
			if bech32Addr != addr.Bech32(iotago.PrefixShimmer) {
				continue
			}

			amount, err := w.Balance(bech32Addr)
			if err != nil {
				gl.OutLogger.Error("w.Balance error. %s : %v", bech32Addr, err)
				continue
			}

			if amount < (config.ProxySendAmount * 20 / 100) {
				if err := model.InsertSendSmrOrder(bech32Addr, config.ProxySendAmount, 3); err != nil {
					gl.OutLogger.Error("model.InsertSendSmrOrder error. %s : %v", bech32Addr, err)
				}
			}
		}
	}
}
