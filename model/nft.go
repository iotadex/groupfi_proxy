package model

import (
	"log"
	"sync"
)

type NameCache struct {
	names map[string]bool
	sync.RWMutex
}

func (nc *NameCache) insert(name string) bool {
	nc.Lock()
	defer nc.Unlock()
	if _, exist := nc.names[name]; exist {
		return false
	}
	nc.names[name] = true
	return true
}

func (nc *NameCache) delete(name string) {
	nc.Lock()
	defer nc.Unlock()
	delete(nc.names, name)
}

var names NameCache = NameCache{
	names: make(map[string]bool),
}

func LoadAllNames() {
	rows, err := db.Query("select `name` from `nft_record`")
	if err != nil {
		log.Panic(err)
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Panic(err)
		}
		names.insert(name)
	}
}

func VerifyAndInsertName(name string) bool {
	return names.insert(name)
}

func DeleteName(name string) {
	names.delete(name)
}

func StoreNft(nftid, name, user, blockid, collectionid string) error {
	names.insert(name)
	_, err := db.Exec("insert into `nft_record`(`nftid`,`name`,`address`,`blockid`,`collectionid`) VALUES(?,?,?,?,?)", nftid, name, user, blockid, collectionid)
	return err
}
