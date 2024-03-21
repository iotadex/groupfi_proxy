package model

import (
	"database/sql"
	"fmt"
	"gproxy/tools"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// db for all mysql database things
var db *sql.DB
var seeds [4]uint64

func SetSeeds(_seeds [4]uint64) {
	seeds = _seeds
}

func ConnectToMysql(host, port, database, usr, pwd string) {
	password := tools.Aes.GetDecryptString(pwd, seeds)
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", usr, string(password), host, port, database))
	if err != nil {
		log.Panic(err)
	}

	if err = db.Ping(); nil != err {
		log.Panic("Connect to Mysql error : " + err.Error())
	}

	LoadAllNames()
}

func Ping() error {
	if db == nil {
		return fmt.Errorf("mysql connection is nil")
	}
	if err := db.Ping(); nil != err {
		return fmt.Errorf("connect to Mysql error : %v", err)
	}
	return nil
}
