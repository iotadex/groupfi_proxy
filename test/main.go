package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"gproxy/api"
	"gproxy/api/middleware"
	"gproxy/tools"
	"log"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

var URL string = "https://testapi.groupfi.ai"

func main() {
	fmt.Println("1. mint name nft")
	fmt.Println("2. get smr price")
	fmt.Println("3. filter group")
	fmt.Println("4. verify group")
	fmt.Println("5. accept owner")
	fmt.Println("6. SetReward")
	fmt.Println("7. Get proxy account")
	for {
		fmt.Print("Key your choice (0 quit): ")
		s := int(0)
		fmt.Scanf("%d", &s)
		switch s {
		case 1:
			MintNameNft()
		case 2:
			GetSmrPrice()
		case 3:
			FilterGroup()
		case 4:
			VerifyGroup()
		case 5:
			Register()
		case 6:
			MintNameNftForMM()
		case 7:
			GetProxyAccount()
		case 0:
			return
		}
	}

	// Register()
}

func MintNameNft() {
	to := "smr1qrgk26r86pntdsav5rpz638vyl3dzmjl7qqmxdf34rdsl8pjqn0qzf8f7ms"
	name := "wangyi123456"
	url := URL + "/mint_nicknft?address=" + to + "&name=" + name
	res, err := tools.HttpGet(url)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(res))
}

func GetSmrPrice() {
	url := URL + "/smr_price"
	res, err := tools.HttpGet(url)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(res))
}

func Register() {
	//0xef87a8bd0430990a943ee8f6eac40e1529eff40a7f0f3bf25e901a0eced63c455bcae1495b358f1b0968588745c5f92afa2ea40d0a3951d9a6d133d6550c1e27
	//5bcae1495b358f1b0968588745c5f92afa2ea40d0a3951d9a6d133d6550c1e27
	//main-evm-account: 0xaa6d9f1cb05c7285fab30eb1fa74c7839e8cb758d2d1be728ac5412b73d6b441
	//0xe2243FfFd353b15F9c74A4D3359F02dB78072758
	privateKey, err := crypto.HexToECDSA("aa6d9f1cb05c7285fab30eb1fa74c7839e8cb758d2d1be728ac5412b73d6b441")
	if err != nil {
		log.Fatal(err)
	}

	md := middleware.MetaData{
		EncryptedPrivateKey: "0x04b8e701fdd0617634243d5cdcad1c2c157f0843f61fba5e7b603b46ee53eff875a7f8cdf85bdb315f5bb935c68c95c0b074a99b7aa6a07f49738ecdcd8b07f0e3f2dbe1a0c99a66086504fc65837538ff587cc84bc4623b98bf492bdec368976c82411861ec05bf809f8487735dfcff2aeaf0926a91b381cb0f5a379432f015825b9603c3b21597bfc09a85795b5ba139",
		PairXPublicKey:      "5bcae1495b358f1b0968588745c5f92afa2ea40d0a3951d9a6d133d6550c1e27",
		EvmAddress:          "0xe2243FfFd353b15F9c74A4D3359F02dB78072758",
		Timestamp:           time.Now().Unix(),
		Scenery:             1,
	}
	data := md.EncryptedPrivateKey + md.EvmAddress + md.PairXPublicKey + strconv.Itoa(md.Scenery) + strconv.FormatInt(md.Timestamp, 10)
	hashData := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	sign, err := crypto.Sign(crypto.Keccak256Hash([]byte(hashData)).Bytes(), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	md.Signature = hexutil.Encode(sign)

	url := URL + "/proxy/register"
	postParams, _ := json.Marshal(md)
	header := make(map[string]string)
	header["Content-Type"] = "application/json; charset=UTF-8"
	res, err := tools.HttpRequest(url, "POST", postParams, header)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(res))
}

func GetProxyAccount() {
	sd := middleware.SignData{
		PublicKey: "0x5bcae1495b358f1b0968588745c5f92afa2ea40d0a3951d9a6d133d6550c1e27",
		Ts:        time.Now().Unix(),
	}

	pk := common.FromHex("0xef87a8bd0430990a943ee8f6eac40e1529eff40a7f0f3bf25e901a0eced63c455bcae1495b358f1b0968588745c5f92afa2ea40d0a3951d9a6d133d6550c1e27")
	sign := ed25519.Sign(pk, []byte(sd.Data+strconv.FormatInt(sd.Ts, 10)))
	sd.Sign = hexutil.Encode(sign)

	url := URL + "/proxy/account"
	postParams, _ := json.Marshal(sd)
	header := make(map[string]string)
	header["Content-Type"] = "application/json; charset=UTF-8"
	res, err := tools.HttpRequest(url, "POST", postParams, header)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(res))
}

func FilterGroup() {
	f := api.Filter{
		Chain: 148,
		Addresses: []string{
			"0x1CB7B54AAB4283782b8aF70d07F88AD795c952E9",
			"0x928100571464c900A2F53689353770455D78a200",
			"0x504dF97f0e5425Eae1D32ACBE5B2E7Dc1f1Dd9cf",
			"0xFf803bc4f2D0516101eB7Aa643299e1BAF5d78F2",
			"0xd2Bae936E942115f1759631f6Ae5642D10B4824e",
		},
		Contract:  "0x544F353C02363D848dBAC8Dc3a818B36B7f9355e",
		Threshold: 1,
		Erc:       721,
		Ts:        time.Now().Unix(),
	}
	url := URL + "/group/filter"
	postParams, _ := json.Marshal(f)
	header := make(map[string]string)
	header["Content-Type"] = "application/json; charset=UTF-8"
	res, err := tools.HttpRequest(url, "POST", postParams, header)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(res))
}

func VerifyGroup() {
	f := api.Verfiy{
		Chain: 148,
		Adds: []string{
			"0x1CB7B54AAB4283782b8aF70d07F88AD795c952E9",
			"0x504dF97f0e5425Eae1D32ACBE5B2E7Dc1f1Dd9cf",
			"0xFf803bc4f2D0516101eB7Aa643299e1BAF5d78F2",
			"0xd2Bae936E942115f1759631f6Ae5642D10B4824e",
		},
		Subs:      []string{},
		Contract:  "0x544F353C02363D848dBAC8Dc3a818B36B7f9355e",
		Threshold: 1,
		Erc:       721,
		Ts:        time.Now().Unix(),
	}
	url := URL + "/group/verify"
	postParams, _ := json.Marshal(f)
	header := make(map[string]string)
	header["Content-Type"] = "application/json; charset=UTF-8"
	res, err := tools.HttpRequest(url, "POST", postParams, header)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(res))
}

func MintNameNftForMM() {
	name := "wangyi123457"
	sd := middleware.SignData{
		PublicKey: "0x5bcae1495b358f1b0968588745c5f92afa2ea40d0a3951d9a6d133d6550c1e27",
		Data:      name,
		Ts:        time.Now().Unix(),
	}

	pk := common.FromHex("0xef87a8bd0430990a943ee8f6eac40e1529eff40a7f0f3bf25e901a0eced63c455bcae1495b358f1b0968588745c5f92afa2ea40d0a3951d9a6d133d6550c1e27")
	sign := ed25519.Sign(pk, []byte(sd.Data+strconv.FormatInt(sd.Ts, 10)))
	sd.Sign = hexutil.Encode(sign)

	url := URL + "/proxy/mint_nicknft"
	postParams, _ := json.Marshal(sd)
	header := make(map[string]string)
	header["Content-Type"] = "application/json; charset=UTF-8"
	res, err := tools.HttpRequest(url, "POST", postParams, header)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(res))
}
