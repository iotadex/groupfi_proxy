package main

import (
	"fmt"
	"gproxy/api"
	"gproxy/config"
	"gproxy/daemon"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/service"
	"gproxy/tokens"
	"gproxy/tools"
	"gproxy/wallet"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	if os.Args[len(os.Args)-1] != "-d" {
		tools.Aes.Input()
		os.Args = append(os.Args, "-d")
	}
	daemon.Background("./out.log", true)

	config.Load()
	setSeeds()

	gl.CreateLogFiles()

	model.ConnectToMysql(config.Db.Host, config.Db.Port, config.Db.DbName, config.Db.Usr, config.Db.Pwd)

	api.StartHttpServer(config.HttpPort)

	service.Start()

	daemon.WaitForKill()

	api.StopHttpServer()
}

func setSeeds() {
	seeds := tools.Aes.ReadRand()
	// set model's seeeds
	model.SetSeeds(seeds)
	// set service's seeds
	wallet.SetSeeds(seeds)
}

func TestFilter() {
	t := tokens.NewEvmToken("https://json-rpc.evm.shimmer.network", "", "0xAEaDcd57E4389678537d82891f095BBbE0ab9610", 148, 0)
	addrs := make([]common.Address, 0)
	addrs = append(addrs, common.HexToAddress("0x1CB7B54AAB4283782b8aF70d07F88AD795c952E9"))
	a, _ := new(big.Int).SetString("170000000000000000000", 10)
	indexes, err := t.FilterEthAddresses(addrs, a)
	fmt.Println(indexes, err)
}
