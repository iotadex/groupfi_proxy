package main

import (
	"context"
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
	"time"

	"github.com/ethereum/go-ethereum/common"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
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

func TestGetBasicOutput(addr string) {
	nodeAPI := nodeclient.New("https://prerelease.api.iotacat.com")
	indexer, err := nodeAPI.Indexer(context.Background())
	if err != nil {
		panic(err)
	}

	notHas := false
	has := true
	nowTs := uint32(time.Now().Unix())
	query := nodeclient.BasicOutputsQuery{
		AddressBech32: addr,
		IndexerNativeTokenParas: nodeclient.IndexerNativeTokenParas{
			HasNativeTokens: &notHas,
		},
		IndexerTimelockParas: nodeclient.IndexerTimelockParas{
			HasTimelock:      &has,
			TimelockedBefore: nowTs,
		},
		IndexerExpirationParas: nodeclient.IndexerExpirationParas{
			HasExpiration: &notHas,
		},
		IndexerStorageDepositParas: nodeclient.IndexerStorageDepositParas{
			HasStorageDepositReturn: &notHas,
		},
	}
	res, err := indexer.Outputs(context.Background(), &query)
	if err != nil {
		panic(err)
	}
	extParas := iotago.ExternalUnlockParameters{
		ConfUnix: nowTs,
	}
	for res.Next() {
		outputs, _ := res.Outputs()
		for _, output := range outputs {
			if len(output.NativeTokenList()) > 0 {
				continue
			}
			if output.UnlockConditionSet().TimelocksExpired(&extParas) != nil {
				continue
			}
			tag := output.FeatureSet().TagFeature()
			if tag != nil {
				println(string(tag.Tag))
			}
		}
	}
}
