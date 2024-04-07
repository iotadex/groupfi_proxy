package service

import (
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/tokens"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	iotago "github.com/iotaledger/iota.go/v3"
)

var evmBuyTasks sync.WaitGroup
var listenTokens map[string]*tokens.EvmToken // symbol->chains.Token

func StartListenBuySmrOrder() {
	listenTokens = make(map[string]*tokens.EvmToken)
	for chainid, node := range config.EvmNodes {
		t := tokens.NewEvmToken(node.Rpc, node.Wss, chainid, node.Contract, node.ListenType)
		listenTokens[chainid] = t
		go listen(chainid, t)
	}
}

func listen(chainid string, t *tokens.EvmToken) {
	evmBuyTasks.Add(1)
	defer evmBuyTasks.Done()

	chLog, chOrder := t.StartListen()
	for {
		select {
		case log := <-chLog:
			switch log.Type {
			case 0:
				gl.OutLogger.Info(log.Info)
			case 1, 2:
				gl.OutLogger.Error(log.Info)
			case 3:
				gl.OutLogger.Info("Listen service %s is stoped!", chainid)
				return
			}
		case order := <-chOrder:
			gl.OutLogger.Info("%s : %s : %s : %s", order.ChainId, order.TxHash, order.User.Hex(), order.Amount.String())
			dealOrder(order)
		}
	}
}

func dealOrder(order tokens.Order) {
	// get the price from db
	p, err := model.GetSmrPrice(order.ChainId)
	if err != nil {
		gl.OutLogger.Error("model.GetSmrPrice error. %v, %v", err, order)
		return
	}
	a, _ := new(big.Int).SetString(p.Amount, 10)
	if a == nil || order.Amount == nil || a.Cmp(order.Amount) < 0 {
		gl.OutLogger.Error("smr price amount is not satisfied. %s, %v", p.Amount, order.Amount)
		return
	}

	// store it to db
	var addr iotago.Ed25519Address
	copy(addr[:], order.EdAddr)
	smrAddr := addr.Bech32(iotago.PrefixShimmer)
	if err := model.StoreBuyOrder(order.ChainId, order.TxHash.Hex(), order.User.Hex(), hexutil.Encode(order.EdAddr), smrAddr, order.Amount.String(), config.ProxySendAmount); err != nil {
		if !strings.HasPrefix(err.Error(), "Error 1062") {
			gl.OutLogger.Error("model.StoreBuyOrder error. %v, %v", err, order)
		}
		return
	}
}

func StopListen() {
	for chainid := range config.EvmNodes {
		t := listenTokens[chainid]
		t.StopListen()
	}
	evmBuyTasks.Wait()
}
