package model

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"gproxy/config"
	"gproxy/tools"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func StoreBuyOrder(chain uint64, txHash, user, edAddr, bech32Addr, amountIn string, amountOut uint64) error {
	// 1. begin a transaction of mysql
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return err
	}

	// 2. insert into buy_order
	if _, err = tx.Exec("INSERT INTO `buy_order`(`chain`,`txhash`,`user`,`ed_addr`,`bech_addr`,`amount_in`,`amount_out`) VALUES(?,?,?,?,?,?,?)", chain, txHash, user, edAddr, bech32Addr, amountIn, amountOut); err != nil {
		tx.Rollback()
		return err
	}

	// 3. insert pending send
	if err := InsertPendingSendSmrOrder(tx, bech32Addr, amountOut, SEND_BUY); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

type SmrPrice struct {
	Contract string `json:"contract"`
	Token    string `json:"token"`
	Price    string `json:"price"`
	Deci     int    `json:"deci"`
}

func GetSmrPrices() (map[uint64]*SmrPrice, error) {
	rows, err := db.Query("SELECT `chain`,`token`,`price`,`deci`,`ts`,`sign` FROM `price`")
	if err != nil {
		return nil, err
	}

	sps := make(map[uint64]*SmrPrice)
	for rows.Next() {
		sp := &SmrPrice{}
		var sign string
		var chain, ts uint64

		if err := rows.Scan(&chain, &sp.Token, &sp.Price, &sp.Deci, &ts, &sign); err != nil {
			return nil, err
		}
		s, _ := tools.Aes.SignDataByECDSA(strconv.FormatUint(chain, 10)+sp.Token+sp.Price+strconv.Itoa(sp.Deci)+strconv.FormatUint(ts, 10), seeds)
		if !bytes.Equal(s, common.FromHex(sign)) {
			return nil, fmt.Errorf("sign error. %s, %s", sign, hexutil.Encode(s))
		}
		sps[chain] = sp
	}
	return sps, nil
}

func GetSmrPrice(chain uint64) (*SmrPrice, error) {
	row := db.QueryRow("SELECT `token`,`price`,`deci`,`ts`,`sign` FROM `price` where `chain`=?", chain)

	sp := &SmrPrice{}
	var ts int64
	var sign string
	if err := row.Scan(&sp.Token, &sp.Price, &sp.Deci, &ts, &sign); err != nil {
		return nil, err
	}
	s, _ := tools.Aes.SignDataByECDSA(strconv.FormatUint(chain, 10)+sp.Token+sp.Price+strconv.Itoa(sp.Deci)+strconv.FormatInt(ts, 10), seeds)
	if !bytes.Equal(s, common.FromHex(sign)) {
		return nil, fmt.Errorf("sign error. %s, %s", sign, hexutil.Encode(s))
	}
	if _, b := new(big.Int).SetString(sp.Price, 10); !b {
		return nil, fmt.Errorf("price error, %s", sp.Price)
	}
	return sp, nil
}

func SignEd25519Hash(msg []byte) ([]byte, error) {
	pk := common.FromHex(string(tools.Aes.GetDecryptString(config.SignEdPk, seeds)))
	if len(pk) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("error private key. length(%d)", len(pk))
	}
	signature := ed25519.Sign(pk, msg)
	return signature, nil
}
