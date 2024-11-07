package service

import (
	"context"
	"gproxy/config"
	"gproxy/model"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/iotaledger/iota.go/v3/nodeclient"
)

var hornetNodes []model.HornetNode
var hornetNodesMux sync.RWMutex

func StartHornetNodes() {
	f := func() bool {
		// Get hornetNodes from db
		nodes, err := model.GetHornetNodes()
		if err != nil || len(nodes) == 0 {
			slog.Error("get hornet nodes", "Err", err, "HornetNodeCount", len(nodes))
			return false
		}

		healthyNodes := make([]model.HornetNode, 0, len(nodes))
		totalWeight := 0
		for i := range nodes {
			// Get the node info to judge it if healthy
			hornetApi := nodeclient.New(nodes[i].Url)
			info, err := hornetApi.Info(context.Background())
			if err != nil {
				slog.Error("nodeApi.Info", "err", err)
				continue
			}
			if !info.Status.IsHealthy {
				slog.Error("hornet node is not healthy", "url", nodes[i].Url)
				continue
			}
			totalWeight += nodes[i].Weight
			nodes[i].Weight = totalWeight
			nodes[i].Info = &info.Protocol
			healthyNodes = append(healthyNodes, nodes[i])
		}

		hornetNodesMux.Lock()
		hornetNodes = healthyNodes
		hornetNodesMux.Unlock()

		if len(healthyNodes) == 0 {
			slog.Error("There is no healthy hornet node")
			return false
		}
		return true
	}

	if !f() {
		panic("There is no healthy hornet node!")
	}

	go func() {
		ticker := time.NewTicker(time.Minute * time.Duration(config.HornetHealthyTime))
		for range ticker.C {
			f()
		}
	}()
}

func GetEnableHornetNode() *model.HornetNode {
	hornetNodesMux.RLock()
	defer hornetNodesMux.RUnlock()

	if len(hornetNodes) == 0 {
		return nil
	}

	l := len(hornetNodes)
	if hornetNodes[l-1].Weight <= 0 {
		return nil
	}

	r := rand.IntN(hornetNodes[l-1].Weight) + 1
	var node model.HornetNode
	for i := l - 1; i >= 0; i-- {
		if r > hornetNodes[i].Weight {
			break
		}
		node = hornetNodes[i]
	}
	return &node
}
