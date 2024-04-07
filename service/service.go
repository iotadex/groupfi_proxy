package service

func Start() {
	RunKeepProxyPoolFull()
	RunSendSmr()
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
