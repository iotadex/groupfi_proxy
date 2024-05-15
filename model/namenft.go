package model

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func InsertNameNftRecord(to, name, meta, collectionid string, expireDays int) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return false, err
	}

	tempNftid := to + strconv.FormatInt(time.Now().UnixNano(), 10)
	if _, err := tx.Exec("INSERT INTO `nft_name_record`(`nftid`,`name`,`to`,`meta`,`expire`,`collectionid`) VALUES(?,?,?,?,?,?)", tempNftid, name, to, meta, expireDays, collectionid); err != nil {
		tx.Rollback()
		if strings.HasPrefix(err.Error(), "Error 1062") {
			return false, nil
		}
		return false, err
	}
	return true, tx.Commit()
}

type NameNftRecord struct {
	Nftid  string
	To     string
	Meta   []byte
	Expire int
}

func PopOneNameNftRecord() (*NameNftRecord, error) {
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return nil, err
	}

	row := tx.QueryRow("select `nftid`,`to`,`meta`,`expire` from `nft_name_record` where `state`=? limit 1 for update", INIT_SEND)
	var nftid, to, meta string
	var expire int
	if err := row.Scan(&nftid, &to, &meta, &expire); err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if res, err := tx.Exec("update `nft_name_record` set `state`=? where `nftid`=?", POP_SEND, nftid); err != nil {
		tx.Rollback()
		return nil, err
	} else {
		if affected, err := res.RowsAffected(); (affected == 0) || (err != nil) {
			tx.Rollback()
			return nil, fmt.Errorf("there is racing when poping nft_name_record from db. %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("PopOneNftNameRecord tx commit error. %v", err)
	}

	return &NameNftRecord{Nftid: nftid, To: to, Meta: common.FromHex(meta), Expire: expire}, nil
}

func UpdateBlockIdToNameNftRecord(nftid, blockid string) error {
	_, err := db.Exec("update `nft_name_record` set `blockid`=?,`state`=? where `nftid`=?", blockid, PENDING_SEND, nftid)
	return err
}

func UpdateNameNft(id, nftid string) error {
	_, err := db.Exec("update `nft_name_record` set `nftid`=?,`meta`=?,`state`=? where `nftid`=?", nftid, "0", CONFIRMED_SEND, id)
	return err
}
