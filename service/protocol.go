package service

/*
var nodeProtocol iotago.ProtocolParameters
var protocolMu sync.RWMutex

func UpateHornetNodeProtocol() {
	nodeApi := nodeclient.New(config.HornetRpcDefault)

	f := func() {
		info, err := nodeApi.Info(context.Background())
		if err != nil {
			slog.Error("nodeApi.Info", "err", err)
			return
		}

		protocolMu.Lock()
		nodeProtocol = info.Protocol
		protocolMu.Unlock()
	}

	f()
	ticker := time.NewTicker(time.Second * time.Duration(config.UpdateProtocolTime))
	for range ticker.C {
		f()
	}
}

func GetHornetNodeProtocol() *iotago.ProtocolParameters {
	protocolMu.RLock()
	defer protocolMu.RUnlock()
	pp := nodeProtocol
	return &pp
}
*/
