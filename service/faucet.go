package service

import (
	"fmt"
	"gproxy/config"
	"gproxy/wallet"
	"log"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

var faucet map[uint64]*wallet.EvmWallet
var faucetUsed map[string]int64
var faucetMu sync.RWMutex

func StartFaucet() {
	faucet = make(map[uint64]*wallet.EvmWallet)
	faucetUsed = make(map[string]int64)
	for chainid, node := range config.FaucetNodes {
		var err error
		faucet[chainid], err = wallet.NewEvmWallet(node.Rpc, node.Wallet, int64(chainid))
		if err != nil {
			log.Panic(err)
		}
	}
}

func FaucetSend(chainid uint64, erc20 string, to string, amount *big.Int) ([]byte, error) {
	if node := config.FaucetNodes[chainid]; amount.Cmp(node.MaxAmount) > 0 {
		return nil, fmt.Errorf("faucet amount over max. %s", amount.String())
	}

	faucetMu.Lock()
	defer faucetMu.Unlock()

	if t, exist := faucetUsed[to]; exist && ((t + 86400) > time.Now().Unix()) {
		return nil, fmt.Errorf("test token time is not over 24 hours")
	}

	w, exist := faucet[chainid]
	if !exist {
		return nil, fmt.Errorf("evm network %d not supposed", chainid)
	}

	hashTx, err := w.SendERC20(erc20, to, amount)
	if err == nil {
		faucetUsed[to] = time.Now().Unix()
		slog.Info("faucet test token", "to", to, "hash", hexutil.Encode(hashTx))
	}
	return hashTx, err
}
