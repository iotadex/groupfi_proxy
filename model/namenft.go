package model

import (
	"strconv"
	"strings"
	"time"
)

func VerifyAndInsertName(user, name, collectionid string) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		return false, err
	}

	tempNftid := user + strconv.FormatInt(time.Now().UnixNano(), 10)
	if _, err := tx.Exec("insert into `nft_record`(`nftid`,`name`,`address`,`collectionid`) VALUES(?,?,?,?)", tempNftid, name, user, collectionid); err != nil {
		tx.Rollback()
		if strings.HasPrefix(err.Error(), "Error 1062") {
			return false, nil
		}
		return false, err
	}
	return true, tx.Commit()
}

func UpdateNameNft(nftid, blockid, name string) error {
	_, err := db.Exec("update `nft_record` set `nftid`=?, `blockid`=? where `name`=?", nftid, blockid, name)
	return err
}
