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
func CreateProxyToPool(amount uint64, minCount int) (int, error) {
	// 1. begin a transaction of mysql
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return 0, err
	}

	// 2. Get the count of proxy pool; check the count smaller than required or not
	row := tx.QueryRow("select count(`address`) from `proxy_pool` where `state`=? for update", CONFIRMED_SEND)
	var count int
	if err := row.Scan(&count); err != nil {
		tx.Rollback()
		return 0, err
	}
	if count < minCount {
		for i := minCount - count; i > 0; i-- {
			// create a ed25519 private key by random number
			bech32Addr, enpk := getEdPrivateKey()
			if _, err := tx.Exec("INSERT INTO `proxy_pool`(`address`,`enpk`) VALUES(?,?)", bech32Addr, enpk); err != nil {
				tx.Rollback()
				return 0, err
			}

			// add it to pending send
			if err := InsertPendingSendSmrOrder(tx, bech32Addr, amount, SEND_POOL); err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	return minCount - count, tx.Commit()
}

// set the proxy's state to 1 after add balance to it
func UpdateProxyPoolState(address string, state int) error {
	_, err := db.Exec("update `proxy_pool` set `state`=? where `address`=? and state!=?", state, address, PENDING_SEND)
	return err
}

func GetProxyPool(state int) (map[string]string, error) {
	rows, err := db.Query("SELECT `address`,`enpk` FROM `proxy_pool` where `state`=?", state)
	if err != nil {
		return nil, err
	}

	addrs := make(map[string]string)
	for rows.Next() {
		var addr, pk string
		if err := rows.Scan(&addr, &pk); err != nil {
			return nil, err
		}
		addrs[addr] = pk
	}
	return addrs, nil
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
