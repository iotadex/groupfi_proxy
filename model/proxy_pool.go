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
func CreateProxyToPool() (string, error) {
	// create a ed25519 private key by random number
	pk := getEdPrivateKey()
	addr := iotago.Ed25519AddressFromPubKey([]byte(pk[32:]))
	address := addr.Bech32(iotago.PrefixShimmer)
	enpk := tools.Aes.GetEncryptString(hex.EncodeToString(pk), seeds)

	_, err := db.Exec("INSERT INTO `proxy_pool`(`address`,`enpk`) VALUES(?,?)", address, enpk)
	return address, err
}

// set the proxy's state to 1 after add balance to it
func InitPoolProxy(address string) error {
	_, err := db.Exec("update `proxy_pool` set `state`=1 where `address`=?", address)
	return err
}

// get the proxies which state are 0 in the proxy pool
func GetUninitializedProxiesFromPool() ([]string, error) {
	rows, err := db.Query("select `address` from `proxy_pool` where `state`=1")
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
func getEdPrivateKey() ed25519.PrivateKey {
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
	return ed25519.NewKeyFromSeed(seed)
}
