package model

import "sync"

var signAccounts *SignAccountToSmrProxy = NewSignAccountToSmrProxy()

type ShimmerAccount struct {
	Account string // evm address
	Chain   string // chain name
	Smr     string // shimmer bech32 address
	EnPk    string // encrypt private key
	State   int    // proxy account' state
	TempTs  int64  // temp account update timestamp
}

type SignAccountToSmrProxy struct {
	users map[string]*ShimmerAccount // temp account -> proxy account
	sync.RWMutex
}

func NewSignAccountToSmrProxy() *SignAccountToSmrProxy {
	return &SignAccountToSmrProxy{
		users: make(map[string]*ShimmerAccount),
	}
}

func (tasp *SignAccountToSmrProxy) Add(signAcc string, sa *ShimmerAccount) {
	tasp.Lock()
	defer tasp.Unlock()
	tasp.users[signAcc] = sa
}

func (tasp *SignAccountToSmrProxy) Get(signAcc string) *ShimmerAccount {
	tasp.RLock()
	defer tasp.RUnlock()
	return tasp.users[signAcc]
}
