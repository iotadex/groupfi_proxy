package service

import (
	"encoding/hex"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/tokens"
	"gproxy/wallet"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	iotago "github.com/iotaledger/iota.go/v3"
)

var evmBuyTasks sync.WaitGroup
var listenTokens map[string]*tokens.EvmToken // symbol->chains.Token

func StartListenSell() {
	listenTokens = make(map[string]*tokens.EvmToken)
	for chainid, node := range config.EvmNodes {
		t := tokens.NewEvmToken(node.Rpc, node.Wss, chainid, node.Contract, node.ListenType)
		listenTokens[chainid] = t
		go listen(chainid, t)
	}
}

// key = chainid + symbol
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
	// store it to db
	var addr iotago.Ed25519Address
	copy(addr[:], order.PubKey)
	smrAddr := addr.Bech32(iotago.PrefixShimmer)
	if err := model.StoreBuyOrder(order.ChainId, order.TxHash.Hex(), order.User.Hex(), hexutil.Encode(order.PubKey), smrAddr, order.Amount.String()); err != nil {
		if !strings.HasPrefix(err.Error(), "Error 1062") {
			gl.OutLogger.Error("model.StoreBuyOrder error. %v, %v", err, order)
		}
		return
	}

	// add a sending order to
	w := wallet.NewIotaSmrWallet(config.SmrRpc, config.MainWallet, config.MainWalletPk, "0x0")
	id, err := w.SendBasic(smrAddr, config.ProxySendAmount)
	if err != nil {
		gl.OutLogger.Error("w.SendBasic error. %s, %v", smrAddr, err)
		return
	}
	gl.OutLogger.Info("Buy smr, send to %s. block_id : 0x%s", smrAddr, hex.EncodeToString(id))
}

func StopListen() {
	for chainid := range config.EvmNodes {
		t := listenTokens[chainid]
		t.StopListen()
	}
	evmBuyTasks.Wait()
}
