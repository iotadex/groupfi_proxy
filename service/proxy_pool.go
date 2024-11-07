package service

import (
	"encoding/hex"
	"gproxy/config"
	"gproxy/model"
	"gproxy/wallet"
	"log/slog"
	"time"
)

func RunKeepProxyPoolFull() {
	f := func() {
		if count, err := model.CreateProxyToPool(config.ProxySendAmount, config.MinProxyPoolCount); err != nil {
			slog.Error("model.CreateProxyToPool", "err", err)
		} else if count > 0 {
			CreateProxyPoolSignal <- true
		}
	}
	f()
	ticker := time.NewTicker(time.Minute * time.Duration(config.ProxyPoolCheckMinutes))
	for range ticker.C {
		f()
	}
}

func RunCheckProxyPoolBalance() {
	ticker := time.NewTicker(time.Hour * time.Duration(config.ProxyBalanceCheckHours))
	for range ticker.C {
		addrs, err := model.GetProxyPool(model.USED_ADDRESS)
		if err != nil {
			slog.Error("model.GetUsedProxyPool", "err", err)
		}

		node := GetEnableHornetNode()
		if node == nil || node.Info == nil {
			slog.Error("RunCheckProxyPoolBalance. There is no healthy hornet node")
			continue
		}
		w := wallet.NewIotaSmrWallet(node.Url, "", "", "")

		for bech32Addr, enpk := range addrs {
			time.Sleep(time.Second * 5)
			if !wallet.ChecKEd25519Addr(enpk, bech32Addr) {
				continue
			}

			amount, err := w.Balance(bech32Addr)
			if err != nil {
				slog.Error("w.Balance", "addr", bech32Addr, "err", err)
				continue
			}

			if amount < (config.ProxySendAmount * 20 / 100) {
				if err := model.InsertSendSmrOrder(bech32Addr, config.ProxySendAmount, 3); err != nil {
					slog.Error("model.InsertSendSmrOrder", "addr", bech32Addr, "err", err)
				}
			}
		}
	}
}

func RunRecycleMsgOutputs() {
	f := func() {
		addrs, err := model.GetProxyPool(model.USED_ADDRESS)
		if err != nil {
			slog.Error("model.GetUsedProxyPool", "err", err)
		}

		for bech32Addr, enpk := range addrs {
			time.Sleep(time.Second * 5)
			if !wallet.ChecKEd25519Addr(enpk, bech32Addr) {
				continue
			}

			node := GetEnableHornetNode()
			if node == nil || node.Info == nil {
				slog.Error("RunRecycleMsgOutputs. There is no healthy hornet node")
				return
			}
			w := wallet.NewIotaSmrWallet(node.Url, bech32Addr, enpk, "")

			id, err := w.Recycle(config.RecycleFilterTags)
			if err != nil {
				slog.Error("w.Recycle", "addr", bech32Addr, "err", err)
				continue
			}
			if len(id) > 0 {
				slog.Info(bech32Addr+" recycle msg", "blockid", hex.EncodeToString(id))
			}
		}
	}
	f()
	ticker := time.NewTicker(time.Hour * 24)
	for range ticker.C {
		f()
	}
}
