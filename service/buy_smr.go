package service

import (
	"gproxy/model"
	"gproxy/tokens"
	"log/slog"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	iotago "github.com/iotaledger/iota.go/v3"
)

var evmBuyTasks sync.WaitGroup
var listenTokens map[uint64]*tokens.EvmToken // chainid->chains.Token

func StartListenBuySmrOrder() {
	listenTokens = make(map[uint64]*tokens.EvmToken)
	chains, err := model.GetChains()
	if err != nil {
		panic(err)
	}
	for _, c := range chains {
		t := tokens.NewEvmToken(c.Rpc, c.Wss, c.Contract, c.ChainID, c.ListenType)
		listenTokens[c.ChainID] = t
		go listen(c.ChainID, t)
	}
}

func listen(chainid uint64, t *tokens.EvmToken) {
	evmBuyTasks.Add(1)
	defer evmBuyTasks.Done()

	chLog, chOrder := t.StartListen()
	for {
		select {
		case log := <-chLog:
			switch log.Type {
			case 0:
				slog.Info(log.Info)
			case 1, 2:
				slog.Error(log.Info)
			case 3:
				slog.Info("Listen service is stoped!", "chainid", chainid)
				return
			}
		case order := <-chOrder:
			slog.Info("new buy order", "order", order)
			dealOrder(order)
		}
	}
}

func dealOrder(order tokens.Order) {
	// get the price from db
	p, err := model.GetSmrPrice(order.ChainId)
	if err != nil {
		slog.Error("model.GetSmrPrice", "order", order, "err", err)
		return
	}
	a, _ := new(big.Int).SetString(p.Price, 10)
	a = a.Mul(order.AmountOut, a)
	if a.Cmp(order.AmountIn) > 0 {
		slog.Error("amountIn is not satisfied", "amountIn", order.AmountIn.String(), "amountOut", order.AmountOut.String(), "price", p.Price)
		return
	}

	// store it to db
	var addr iotago.Ed25519Address
	copy(addr[:], order.EdAddr)
	smrAddr := addr.Bech32(iotago.PrefixShimmer)
	if err := model.StoreBuyOrder(order.ChainId, order.TxHash.Hex(), order.User.Hex(), hexutil.Encode(order.EdAddr), smrAddr, order.AmountIn.String(), order.AmountOut.Uint64()); err != nil {
		if !strings.HasPrefix(err.Error(), "Error 1062") {
			slog.Error("model.StoreBuyOrder", "order", order, "err", err)
		}
		return
	}
	buySmrSignal <- true
}

func StopListen() {
	for _, t := range listenTokens {
		t.StopListen()
	}
	evmBuyTasks.Wait()
}
