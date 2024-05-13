package config

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
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

const (
	BuySmr            = 1
	KeepProxyPool     = 2
	CheckProxyBalance = 3
	SendSmr           = 4
)

var serviceType map[string]int = map[string]int{
	"buy_smr":             BuySmr,
	"keep_proxy_pool":     KeepProxyPool,
	"check_proxy_balance": CheckProxyBalance,
	"send_smr":            SendSmr,
}

var (
	HttpPort            int             // http service port
	Db                  db              // database config
	EvmNodes            map[uint64]node // evm node config of groupfi
	ShimmerRpc          string          // shimmer L1 network rpc url
	UpdateProtocolTime  int64           // update protocol parameters of shimmer node, time as seconds
	SendIntervalTime    int64           // the interval time of sending smr, seconds
	ProxyPoolCheckHours int64           // the interval time of checking proxy pool's count, hours
	MinProxyPoolCount   int             // the min proxy pool's count
	ProxySendAmount     uint64          // the amount of sending smr per time
	ProxyWallet         string          // this is "1"
	ProxyPkNftTag       string          // pk nft's tag
	NameNftId           string          // name nft id
	NameNftDays         int             // the expired time of the name nft, days
	DefaultImg          string          // default_image url of name nft
	NameNftTag          string          // name nft tag, string
	MaxMsgLockTime      int64           // max lock time of msg output, seconds
	Services            map[int]bool    // service runs or not

	SignEdPk string // use it to sign the group addresses data, send back to user
)

// Load load config file
func Load() {
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
		UpdateProtocolTime  int64           `json:"update_protocol_time"`
		ProxyPoolCheckHours int64           `json:"proxy_pool_check_hours"`
		MinProxyPoolCount   int             `json:"min_proxy_pool_count"`
		ProxySendAmount     uint64          `json:"proxy_send_amount"`
		ProxyWallet         string          `json:"proxy_wallet"`
		ProxyPkNftTag       string          `json:"proxy_pk_nft_tag"`
		NameNftId           string          `json:"name_nftid"`
		NameNftDays         int             `json:"name_nft_days"`
		DefaultImg          string          `json:"default_img"`
		NameNftTag          string          `json:"name_nft_tag"`
		MaxMsgLockDays      int64           `json:"max_msg_locked_days"`
		Services            map[string]bool `json:"services"`
		SignEdPk            string          `json:"sign_ed_pk"`
	}
	all := &Config{}
	if err = json.NewDecoder(file).Decode(all); err != nil {
		log.Panic(err)
	}
	HttpPort = all.HttpPort
	Db = all.Db
	EvmNodes = make(map[uint64]node)
	for id, node := range all.EvmNodes {
		chainid, _ := strconv.ParseUint(id, 10, 64)
		EvmNodes[chainid] = node
	}
	ShimmerRpc = all.ShimmerRpc
	UpdateProtocolTime = all.UpdateProtocolTime
	SendIntervalTime = all.SendIntervalTime
	ProxyPoolCheckHours = all.ProxyPoolCheckHours
	MinProxyPoolCount = all.MinProxyPoolCount
	ProxySendAmount = all.ProxySendAmount
	ProxyWallet = all.ProxyWallet
	ProxyPkNftTag = all.ProxyPkNftTag
	NameNftId = all.NameNftId
	NameNftDays = all.NameNftDays
	DefaultImg = all.DefaultImg
	NameNftTag = all.NameNftTag
	MaxMsgLockTime = all.MaxMsgLockDays * 3600 * 24
	Services = make(map[int]bool)
	for s, b := range all.Services {
		Services[serviceType[s]] = b
	}
	SignEdPk = all.SignEdPk
}
