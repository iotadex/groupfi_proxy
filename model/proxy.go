package model

import (
	"database/sql"
	"fmt"
	"sync"
)

// Register a shimmer address as the proxy of evm account, if the proxy is exist, update the sign_acc
// @chain		: the evm network chain symbol
// @account		: the evm address
// @signAcc 	: the sign account, a evm address
// @return, the proxy account, a shimmer address
func RegisterProxyFromPool(account string, signAcc string) (string, error) {
	// 1. Check db that the shimmer proxy address is exist or not
	row := db.QueryRow("select `smr` from `proxy` where `account`=?", account)
	var smr string
	if err := row.Scan(&smr); err == nil {
		// update the sign_acc
		_, err = db.Exec("update `proxy` set `sign_acc`=? where `account`=?", signAcc, account)
		return smr, err
	} else if err != sql.ErrNoRows {
		return "", err
	}

	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return "", err
	}
	// 2. get a proxy from proxy_pool
	row = tx.QueryRow("select `id`,`address`,`enpk` from `proxy_pool` where `state`=? limit 1 for update", CONFIRMED_SEND)
	var id int64
	var enpk string
	if err := row.Scan(&id, &smr, &enpk); err != nil {
		tx.Rollback()
		return "", err
	}

	// 3. store the proxy to db
	if _, err := tx.Exec("insert into `proxy`(`account`,`sign_acc`,`smr`,`pk`) VALUES(?,?,?,?)", account, signAcc, smr, enpk); err != nil {
		tx.Rollback()
		return "", err
	}

	// 4. change the proxy from proxy_pool
	if res, err := tx.Exec("update `proxy_pool` set `state`=? where `id`=?", USED_ADDRESS, id); err != nil {
		tx.Rollback()
		return "", err
	} else {
		if affected, err := res.RowsAffected(); (affected == 0) || (err != nil) {
			tx.Rollback()
			return "", fmt.Errorf("there is racing when moving proxy from pool. %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("tx commit error. %v", err)
	}

	return smr, nil
}

func GetProxyAccount(signAcc string) (*ShimmerAccount, error) {
	proxy := signAccounts.Get(signAcc)
	if proxy != nil {
		return proxy, nil
	}

	proxy, err := getProxyAccount(signAcc)
	if err != nil {
		return nil, err
	}
	if proxy != nil {
		signAccounts.Add(signAcc, proxy)
	}
	return proxy, nil
}

func getProxyAccount(signAcc string) (*ShimmerAccount, error) {
	row := db.QueryRow("select `account`,`smr`,`pk` from `proxy` where `sign_acc`=?", signAcc)
	var acc, smr, pk string
	if err := row.Scan(&acc, &smr, &pk); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, nil
	}
	return &ShimmerAccount{
		Account: acc,
		Smr:     smr,
		EnPk:    pk,
	}, nil
}

var signAccounts *SignAccountToSmrProxy = NewSignAccountToSmrProxy()

type ShimmerAccount struct {
	Account string // evm address
	Smr     string // shimmer bech32 address
	EnPk    string // encrypt private key
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
