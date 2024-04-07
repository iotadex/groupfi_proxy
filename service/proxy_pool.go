package service

import (
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"time"
)

func RunKeepProxyPoolFull() {
	go func() {
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
	}()
}
