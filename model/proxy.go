package model

import (
	"crypto/rand"
	"database/sql"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

// Register a shimmer address as the proxy of evm account, if the proxy is exist, update the sign_acc
// @chain		: the evm network chain symbol
// @account		: the evm address
// @signAcc 	: the sign account, a evm address
// @return, the proxy account, a shimmer address
func RegisterProxyFromPool(account string, signAcc string) (string, error) {
	// 1. Check db that the shimmer proxy address is exist or not
	row := db.QueryRow("select `smr`,`sign_acc` from `proxy` where `account`=?", account)
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
	row = tx.QueryRow("select `id`,`address`,`enpk` from `proxy_pool` where `state`=? limit 1", CONFIRMED_SEND)
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

	// 4. delete the proxy from proxy_pool
	if res, err := tx.Exec("delete from `proxy` where `id`=?", id); err != nil {
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
	row := db.QueryRow("select `account`,`chain`,`smr`,`pk`,`state` from `proxy` where `sign_acc`=?", signAcc)
	var acc, chain, smr, pk string
	var state int
	if err := row.Scan(&acc, &chain, &smr, &pk, &state); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, nil
	}
	return &ShimmerAccount{
		Account: acc,
		Chain:   chain,
		Smr:     smr,
		EnPk:    pk,
		State:   state,
	}, nil
}

func EncryptByPublicKey(srcData, publicKeyBytes []byte) (string, error) {
	publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return "", err
	}
	encryptBytes, err := ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(publicKey), srcData, nil, nil)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(encryptBytes), nil
}
