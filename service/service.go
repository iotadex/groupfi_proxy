package service

import "gproxy/config"

func Start() {
	go UpateShimmerNodeProtocol()

	go StartUpdateCacheOutputs()

	if config.Services[config.KeepProxyPool] {
		go RunKeepProxyPoolFull()
	}

	if config.Services[config.CheckProxyBalance] {
		go RunCheckProxyPoolBalance()
	}

	if config.RecycleMsgTime > 0 {
		go RunRecycleMsgOutputs()
	}

	if config.Services[config.SendSmr] {
		go RunSendSmr()
	}

	if config.Services[config.BuySmr] {
		go StartListenBuySmrOrder()
	}

	if len(config.NameNftId) > 0 {
		go RunMintNameNft()
	}

	if config.Services[config.Faucet] {
		StartFaucet()
	}
}

func Stop() {

}
