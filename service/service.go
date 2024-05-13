package service

import "gproxy/config"

func Start() {
	go UpateShimmerNodeProtocol()

	if config.Services[config.KeepProxyPool] {
		go RunKeepProxyPoolFull()
	}

	if config.Services[config.CheckProxyBalance] {
		go RunCheckProxyPoolBalance()
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
}

func Stop() {

}
