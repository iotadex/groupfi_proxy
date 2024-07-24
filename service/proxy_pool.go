package service

import (
	"encoding/hex"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/wallet"
	"time"
)

func RunKeepProxyPoolFull() {
	f := func() {
		if count, err := model.CreateProxyToPool(config.ProxySendAmount, config.MinProxyPoolCount); err != nil {
			gl.OutLogger.Error("model.CreateProxyToPool error. %v", err)
		} else if count > 0 {
			CreateProxyPoolSignal <- true
		}
	}
	f()
	ticker := time.NewTicker(time.Hour * time.Duration(config.ProxyPoolCheckMinutes))
	for range ticker.C {
		f()
	}
}

func RunCheckProxyPoolBalance() {
	ticker := time.NewTicker(time.Hour * time.Duration(config.ProxyBalanceCheckHours))
	for range ticker.C {
		addrs, err := model.GetProxyPool(model.USED_ADDRESS)
		if err != nil {
			gl.OutLogger.Error("model.GetUsedProxyPool error. %v", err)
		}

		for bech32Addr, enpk := range addrs {
			time.Sleep(time.Second * 5)
			if !wallet.ChecKEd25519Addr(enpk, bech32Addr) {
				continue
			}

			w := wallet.NewIotaSmrWallet(config.ShimmerRpc, bech32Addr, enpk, "")

			id, err := w.Recycle(time.Now().Unix() - config.RecycleMsgTime)
			if err != nil {
				gl.OutLogger.Error("w.Recycle error. %s : %v", bech32Addr, err)
				continue
			}
			gl.OutLogger.Error("%s recycle msg, blockid : %s", bech32Addr, hex.EncodeToString(id))
		}
	}
}

func RunRecycleMsgOutputs() {
	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, "", "", "")
	ticker := time.NewTicker(time.Hour * time.Duration(config.ProxyBalanceCheckHours))
	for range ticker.C {
		addrs, err := model.GetProxyPool(model.USED_ADDRESS)
		if err != nil {
			gl.OutLogger.Error("model.GetUsedProxyPool error. %v", err)
		}

		for bech32Addr, enpk := range addrs {
			time.Sleep(time.Second * 5)
			if !wallet.ChecKEd25519Addr(enpk, bech32Addr) {
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
