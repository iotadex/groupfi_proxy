package service

import (
	"fmt"
	"gproxy/config"
	"gproxy/model"
)

func SendSignAccToShimmer(signAcc string, metadata []byte) error {
	// 1. Get the proxy account
	proxy, err := model.GetProxyAccount(signAcc)
	if err != nil {
		return err
	}
	if proxy == nil {
		return fmt.Errorf("proxy account is not exist")
	}
	if !sendingMan.Push(proxy.Account, proxy.Smr, proxy.EnPk, &MetaMsg{metadata, config.MetaSignAccLockTime}) {
		return fmt.Errorf("send meta to shimmer error. %s, %s", proxy.Account, proxy.Smr)
	}
	return nil
}

func SendMsgToShimmer(signAcc string, metadata []byte) error {
	// 1. Get the proxy account
	proxy, err := model.GetProxyAccount(signAcc)
	if err != nil {
		return err
	}
	if proxy == nil {
		return fmt.Errorf("proxy account is not exist")
	}

	if !sendingMan.Push(proxy.Account, proxy.Smr, proxy.EnPk, &MetaMsg{metadata, config.MetaMsgLockTime}) {
		return fmt.Errorf("send meta to shimmer error. %s, %s", proxy.Account, proxy.Smr)
	}
	return nil
}
