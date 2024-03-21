package config

import (
	"encoding/json"
	"log"
	"os"
)

type db struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	DbName string `json:"dbname"`
	Usr    string `json:"usr"`
	Pwd    string `json:"pwd"`
}

var (
	Db                  db
	HttpPort            int
	SmrRpc              string
	SendIntervalTime    int64
	MetaSignAccLockTime int64
	MetaMsgLockTime     int64
	QueueExpiredHours   int64
	ProxyPoolCheckHours int64 // hours
	MinProxyPoolCount   int
	ProxySendAmount     uint64
	MainWallet          string
	MainWalletPk        string
	NameNftId           string
	PkNftId             string
	DefaultImg          string
	Days                int
	MaxMsgLockTime      int64
)

// Load load config file
func init() {
	file, err := os.Open("config/config.json")
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	type Config struct {
		HttpPort int `json:"http_port"`
		Db       db  `json:"db"`
	}
	all := &Config{}
	if err = json.NewDecoder(file).Decode(all); err != nil {
		log.Panic(err)
	}
	Db = all.Db
	HttpPort = all.HttpPort
}
