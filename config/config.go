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

type node struct {
	Rpc        string `json:"rpc"`         // rpc url of evm node
	Wss        string `json:"wss"`         // wss url of evm node
	Contract   string `json:"contract"`    // contract address of groupfi
	ListenType int    `json:"listen_type"` // 0: listen event log, 1: scan event log
}

var (
	HttpPort            int             // http service port
	Db                  db              // database config
	EvmNodes            map[string]node // evm node config of groupfi
	ShimmerRpc          string          // shimmer L1 network rpc url
	SendIntervalTime    int64           // the interval time of sending smr, seconds
	ProxyPoolCheckHours int64           // the interval time of checking proxy pool's count, hours
	MinProxyPoolCount   int             // the min proxy pool's count
	ProxySendAmount     uint64          // the amount of sending smr per time
	ProxyWallet         string          // this is "1"
	ProxyPkNftTag       string          // pk nft's tag
	NameNftId           string          // name nft id
	NameNftDays         int             // the expired time of the name nft, days
	DefaultImg          string          // default_image url of name nft
	MaxMsgLockTime      int64           // max lock time of msg output, seconds

	SignEdPk string // use it to sign the group addresses data, send back to user
)

// Load load config file
func init() {
	file, err := os.Open("config/config.json")
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	type Config struct {
		HttpPort            int             `json:"http_port"`
		Db                  db              `json:"db"`
		EvmNodes            map[string]node `json:"evm_node"`
		ShimmerRpc          string          `json:"shimmer_rpc"`
		SendIntervalTime    int64           `json:"send_interval_time"`
		ProxyPoolCheckHours int64           `json:"proxy_pool_check_hours"`
		MinProxyPoolCount   int             `json:"min_proxy_pool_count"`
		ProxySendAmount     uint64          `json:"proxy_send_amount"`
		ProxyWallet         string          `json:"proxy_wallet"`
		ProxyPkNftTag       string          `json:"proxy_pk_nft_tag"`
		NameNftId           string          `json:"name_nftid"`
		NameNftDays         int             `json:"name_nft_days"`
		DefaultImg          string          `json:"default_img"`
		MaxMsgLockDays      int64           `json:"max_msg_locked_days"`
	}
	all := &Config{}
	if err = json.NewDecoder(file).Decode(all); err != nil {
		log.Panic(err)
	}
	HttpPort = all.HttpPort
	Db = all.Db
	EvmNodes = all.EvmNodes
	ShimmerRpc = all.ShimmerRpc
	SendIntervalTime = all.SendIntervalTime
	ProxyPoolCheckHours = all.ProxyPoolCheckHours
	MinProxyPoolCount = all.MinProxyPoolCount
	ProxySendAmount = all.ProxySendAmount
	ProxyWallet = all.ProxyWallet
	ProxyPkNftTag = all.ProxyPkNftTag
	NameNftId = all.NameNftId
	NameNftDays = all.NameNftDays
	DefaultImg = all.DefaultImg
	MaxMsgLockTime = all.MaxMsgLockDays * 3600 * 24
}
