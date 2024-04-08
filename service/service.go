package service

func Start() {
	RunKeepProxyPoolFull()
	RunSendSmr()
	StartListenBuySmrOrder()
}

func Stop() {

}
