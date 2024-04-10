package service

import "gproxy/config"

func Start() {
	RunKeepProxyPoolFull()
	RunSendSmr()
	StartListenBuySmrOrder()
	if len(config.NameNftId) > 0 {
		go RunMintNameNft()
	}
}

func Stop() {

}
