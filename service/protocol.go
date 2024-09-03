package service

import (
	"context"
	"gproxy/config"
	"log/slog"
	"sync"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
)

var nodeProtocol iotago.ProtocolParameters
var protocolMu sync.RWMutex

func UpateShimmerNodeProtocol() {
	nodeApi := nodeclient.New(config.ShimmerRpc)

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

func GetShimmerNodeProtocol() *iotago.ProtocolParameters {
	protocolMu.RLock()
	defer protocolMu.RUnlock()
	pp := nodeProtocol
	return &pp
}
