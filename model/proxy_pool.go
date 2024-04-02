package model

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"gproxy/tools"

	iotago "github.com/iotaledger/iota.go/v3"
)

// create a proxy and store it to db, the init state is 0 which cannot be used
func CreateProxyToPool(amount uint64, minCount int) error {
	// 1. begin a transaction of mysql
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return err
	}

	// 2. Get the count of proxy pool; check the count smaller than required or not
	row := tx.QueryRow("select count(`address`) from `proxy`")
	var count int
	if err := row.Scan(&count); err != nil {
		tx.Rollback()
		return err
	}
	if count < minCount {
		for i := minCount - count; i > -1; i-- {
			// create a ed25519 private key by random number
			bech32Addr, enpk := getEdPrivateKey()
			if _, err := tx.Exec("INSERT INTO `proxy_pool`(`address`,`enpk`) VALUES(?,?)", bech32Addr, enpk); err != nil {
				tx.Rollback()
				return err
			}

			// add it to pending send
			if err := InsertPendingSendSmrOrder(tx, bech32Addr, amount, SEND_POOL); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

// set the proxy's state to 1 after add balance to it
func InitPoolProxy(address string) error {
	_, err := db.Exec("update `proxy_pool` set `state`=1 where `address`=?", address)
	return err
}

// get the proxies which state are 0 in the proxy pool
func GetUninitializedProxiesFromPool() ([]string, error) {
	rows, err := db.Query("select `address` from `proxy_pool` where `state`=0")
	if err != nil {
		return nil, err
	}
	proxies := make([]string, 0)
	for rows.Next() {
		var proxy string
		if err := rows.Scan(&proxy); err != nil {
			return nil, err
		}
		proxies = append(proxies, proxy)
	}

	return proxies, nil
}

// count the total proxies in the  pool
func CountPoolProxies() (int, error) {
	row := db.QueryRow("select count(`address`) from `proxy`")
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// get private key of ed25519 type
func getEdPrivateKey() (string, string) {
	seed := make([]byte, 0)
	for i := 0; i < 8; i++ {
		var n uint32
		if err := binary.Read(rand.Reader, binary.LittleEndian, &n); err != nil {
			panic(err)
		}
		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, n)
		seed = append(seed, bytesBuffer.Bytes()...)
	}
	pk := ed25519.NewKeyFromSeed(seed)
	addr := iotago.Ed25519AddressFromPubKey([]byte(pk[32:]))
	bech32Addr := addr.Bech32(iotago.PrefixShimmer)
	enpk := tools.Aes.GetEncryptString(hex.EncodeToString(pk), seeds)
	return bech32Addr, hex.EncodeToString(enpk)
}
