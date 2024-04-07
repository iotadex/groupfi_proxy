package model

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"gproxy/tools"
	"strconv"
	"time"
)

type SendingType int

const (
	SEND_POOL = 1
	SEND_BUY  = 2
)

type PendingSendSmrOrder struct {
	Id      int64
	BlockId string
	To      string
	Amount  uint64
	Type    SendingType
	Ts      int64
}

func GetWallet(nftid string) (string, string, error) {
	row := db.QueryRow("select `address`,`pk` from `issuer` where `nftid`=?", nftid)
	var addr, pk string
	err := row.Scan(&addr, &pk)
	return addr, pk, err
}

func InsertPendingSendSmrOrder(tx *sql.Tx, to string, amount uint64, _t SendingType) error {
	ts := time.Now().UnixMilli()
	data := to + strconv.FormatUint(amount, 10) + strconv.Itoa(int(_t)) + "0" + strconv.FormatInt(ts, 10)
	sign, _ := tools.Aes.SignDataByECDSA(data, seeds)
	_, err := tx.Exec("INSERT INTO `send_smr`(`to`,`amount`,`type`,`state`,`ts`,`sign`) VALUES (?,?,?,0,?,?)", to, amount, _t, ts, hex.EncodeToString(sign))
	return err
}

func StoreBackPendingSendSmrOrder(to string, amount uint64, _t SendingType) error {
	ts := time.Now().UnixMilli()
	data := to + strconv.FormatUint(amount, 10) + strconv.Itoa(int(_t)) + "0" + strconv.FormatInt(ts, 10)
	sign, _ := tools.Aes.SignDataByECDSA(data, seeds)
	_, err := db.Exec("INSERT INTO `send_smr`(`to`,`amount`,`type`,`state`,`ts`,`sign`) VALUES (?,?,?,0,?,?)", to, amount, _t, ts, hex.EncodeToString(sign))
	return err
}

func PopOnePendingSendSmrOrder() (*PendingSendSmrOrder, error) {
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return nil, err
	}

	row := tx.QueryRow("select `id`,`to`,`amount`,`type`,`ts`,`sign` from `send_smr` where `state`=0 limit 1")
	var to, sign string
	var _t int
	var id, ts int64
	var amount uint64
	if err := row.Scan(&id, &to, &amount, &_t, &ts, &sign); err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	data := to + strconv.FormatUint(amount, 10) + strconv.Itoa(_t) + "0" + strconv.FormatInt(ts, 10)
	_sign, _ := tools.Aes.SignDataByECDSA(data, seeds)
	if sign != hex.EncodeToString(_sign) {
		tx.Rollback()
		return nil, fmt.Errorf("sign error. %d", id)
	}

	if res, err := tx.Exec("update `send_smr` set `state`=1 where `id`=?", id); err != nil {
		tx.Rollback()
		return nil, err
	} else {
		if affected, err := res.RowsAffected(); (affected == 0) || (err != nil) {
			tx.Rollback()
			return nil, fmt.Errorf("there is racing when poping send_smr from db. %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("tx commit error. %v", err)
	}

	return &PendingSendSmrOrder{Id: id, To: to, Amount: amount, Type: SendingType(_t), Ts: ts}, nil
}

func UpdatePendingOrderblockId(id int64, blockid string) error {
	_, err := db.Exec("update `send_smr` set `blockid`=?,`state`=1 where `id`=?", blockid, id)
	return err
}

func GetPendingOrders() ([]*PendingSendSmrOrder, error) {
	rows, err := db.Query("select `id`,`blockid`,`to`,`type` from `send_smr` where `state`=1")
	if err != nil {
		return nil, err
	}
	orders := make([]*PendingSendSmrOrder, 0)
	for rows.Next() {
		psso := PendingSendSmrOrder{}
		if err := rows.Scan(&psso.Id, &psso.BlockId, &psso.To, &psso.Type); err != nil {
			return nil, err
		}
		orders = append(orders, &psso)
	}
	return orders, nil
}

func UpdatePendingOrderState(id int64, state int) error {
	_, err := db.Exec("update `send_smr` set `state`=? where `id`=?", state, id)
	return err
}
