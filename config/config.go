package config

import (
	"encoding/json"
	"log"
	"math/big"
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

type faucetNode struct {
	Rpc          string `json:"rpc"`        // rpc url of evm node
	Wallet       string `json:"wallet"`     // sending wallet address
	MaxAmountStr string `json:"max_amount"` // max sending amount
	MaxAmount    *big.Int
}

const (
	BuySmr            = 1
	KeepProxyPool     = 2
	CheckProxyBalance = 3
	RecycleProxy      = 4
	SendSmr           = 5
	Faucet            = 6
)

var serviceType map[string]int = map[string]int{
	"buy_smr":             BuySmr,
	"keep_proxy_pool":     KeepProxyPool,
	"check_proxy_balance": CheckProxyBalance,
	"recycle_proxy":       RecycleProxy,
	"send_smr":            SendSmr,
	"faucet":              Faucet,
}

var (
	HttpPort               int                   // http service port
	Db                     db                    // database config
	HornetHealthyTime      int64                 // update hornet nodes, time as minutes
	SendIntervalTime       int64                 // the interval time of sending smr, seconds
	ProxyPoolCheckMinutes  int64                 // the interval time of checking proxy pool's count, minutes
	MinProxyPoolCount      int                   // the min proxy pool's count
	ProxyBalanceCheckHours int64                 // the interval time of checking proxy balance, hours
	ProxySendAmount        uint64                // the amount of sending smr per time
	ProxyWallet            string                // this is "1"
	ProxyPkNftTag          string                // pk nft's tag
	SplitLeftAmount        uint64                // divided left output's amount
	SplitLeftCount         uint64                // the output count of divided left output
	NameNftId              string                // name nft id
	NameNftDays            int                   // the expired time of the name nft, days
	DefaultImg             string                // default_image url of name nft
	NameNftTag             string                // name nft tag, string
	MaxMsgLockTime         int64                 // max lock time of msg output, seconds
	RecycleFilterTags      [][]byte              // recycle filter tags
	SignPrefix             string                // sign prefix "Creating account... "
	GroupfiDataUri         string                // groupfi data uri
	Services               map[int]bool          // service runs or not
	FaucetNodes            map[uint64]faucetNode // evm node for sending faucet

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
		HttpPort               int                   `json:"http_port"`
		Db                     db                    `json:"db"`
		EvmNodes               map[string]node       `json:"evm_node"`
		HornetHealthyTime      int64                 `json:"hornet_healthy_time"`
		SendIntervalTime       int64                 `json:"send_interval_time"`
		ProxyPoolCheckMinutes  int64                 `json:"proxy_pool_check_minutes"`
		MinProxyPoolCount      int                   `json:"min_proxy_pool_count"`
		ProxyBalanceCheckHours int64                 `json:"proxy_balance_check_hours"`
		ProxySendAmount        uint64                `json:"proxy_send_amount"`
		ProxyWallet            string                `json:"proxy_wallet"`
		ProxyPkNftTag          string                `json:"proxy_pk_nft_tag"`
		SplitLeftAmount        uint64                `json:"split_left_amount"`
		SplitLeftCount         uint64                `json:"split_left_count"`
		NameNftId              string                `json:"name_nftid"`
		NameNftDays            int                   `json:"name_nft_days"`
		DefaultImg             string                `json:"default_img"`
		NameNftTag             string                `json:"name_nft_tag"`
		MaxMsgLockDays         int64                 `json:"max_msg_locked_days"`
		RecycleFilterTags      []string              `json:"recycle_filter_tags"`
		SignPrefix             string                `json:"sign_prefix"`
		GroupfiDataUri         string                `json:"groupfi_data_uri"`
		Services               map[string]bool       `json:"services"`
		SignEdPk               string                `json:"sign_ed_pk"`
		FaucetNodes            map[string]faucetNode `json:"faucet_node"`
	}
	all := &Config{}
	if err = json.NewDecoder(file).Decode(all); err != nil {
		log.Panic(err)
	}
	HttpPort = all.HttpPort
	Db = all.Db
	HornetHealthyTime = all.HornetHealthyTime
	SendIntervalTime = all.SendIntervalTime
	ProxyPoolCheckMinutes = all.ProxyPoolCheckMinutes
	MinProxyPoolCount = all.MinProxyPoolCount
	ProxyBalanceCheckHours = all.ProxyBalanceCheckHours
	ProxySendAmount = all.ProxySendAmount
	ProxyWallet = all.ProxyWallet
	ProxyPkNftTag = all.ProxyPkNftTag
	SplitLeftAmount = all.SplitLeftAmount
	if SplitLeftAmount < 100000 {
		SplitLeftAmount = 100000
	}
	SplitLeftCount = all.SplitLeftCount
	if SplitLeftCount < 1 {
		SplitLeftCount = 1
	}
	NameNftId = all.NameNftId
	NameNftDays = all.NameNftDays
	DefaultImg = all.DefaultImg
	NameNftTag = all.NameNftTag
	MaxMsgLockTime = all.MaxMsgLockDays * 3600 * 24
	for _, tag := range all.RecycleFilterTags {
		RecycleFilterTags = append(RecycleFilterTags, []byte(tag))
	}
	SignPrefix = all.SignPrefix
	GroupfiDataUri = all.GroupfiDataUri
	Services = make(map[int]bool)
	for s, b := range all.Services {
		Services[serviceType[s]] = b
	}
	SignEdPk = all.SignEdPk

	FaucetNodes = make(map[uint64]faucetNode)
	for id, node := range all.FaucetNodes {
		chainid, _ := strconv.ParseUint(id, 10, 64)
		b := false
		if node.MaxAmount, b = new(big.Int).SetString(node.MaxAmountStr, 10); !b {
			log.Panic("faucet max amount error : " + node.MaxAmountStr)
		}
		FaucetNodes[chainid] = node
	}
}
