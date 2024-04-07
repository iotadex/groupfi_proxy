package api

import (
	"encoding/hex"
	"encoding/json"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/service"
	"gproxy/tokens"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	iotago "github.com/iotaledger/iota.go/v3"
	"golang.org/x/crypto/blake2b"
)

// Get the prices of smr on different evm chains
func SmrPrice(c *gin.Context) {
	sps, err := model.GetSmrPrices()
	if err != nil {
		gl.OutLogger.Error("model.GetSmrPrices error. %v", err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}
	for id, p := range sps {
		if sp, exist := config.EvmNodes[id]; exist {
			p.Contract = sp.Contract
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"result": true,
		"data":   sps,
	})
}

type Filter struct {
	Chain     uint64   `json:"chain"`
	Addresses []string `json:"addresses"`
	Contract  string   `json:"contract"`
	Threshold int64    `json:"threshold"`
	Erc       int      `json:"erc"`
}

func FilterGroup(c *gin.Context) {
	f := Filter{}
	err := c.BindJSON(&f)
	node, exist := config.EvmNodes[f.Chain]
	if err != nil || !exist {
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "params error. ",
		})
		return
	}
	addrs := make([]common.Address, 0, len(f.Addresses))
	for i := range f.Addresses {
		addrs = append(addrs, common.HexToAddress(f.Addresses[i]))
	}

	var indexes []uint16
	t := tokens.NewEvmToken(node.Rpc, node.Wss, node.Contract, f.Chain, 0)
	if f.Erc == 20 {
		indexes, err = t.FilterERC20Addresses(addrs, common.HexToAddress(f.Contract), big.NewInt(f.Threshold))
	} else if f.Erc == 721 {
		indexes, err = t.FilterERC721Addresses(addrs, common.HexToAddress(f.Contract))
	}

	if err != nil {
		gl.OutLogger.Error("Filter addresses from group error. %d, %s, %v", f.Chain, f.Contract, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  true,
		"indexes": indexes,
	})
}

type Verfiy struct {
	Chain     uint64   `json:"chain"`
	Adds      []string `json:"adds"`
	Subs      []string `json:"subs"`
	Contract  string   `json:"contract"`
	Threshold int64    `json:"threshold"`
	Erc       int      `json:"erc"`
}

func VerifyGroup(c *gin.Context) {
	f := Verfiy{}
	err := c.BindJSON(&f)
	node, exist := config.EvmNodes[f.Chain]
	if err != nil || !exist {
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "params error. ",
		})
		return
	}
	adds := make([]common.Address, 0, len(f.Adds))
	subs := make([]common.Address, 0, len(f.Subs))
	for i := range f.Adds {
		adds = append(adds, common.HexToAddress(f.Adds[i]))
	}
	for i := range f.Subs {
		subs = append(subs, common.HexToAddress(f.Subs[i]))
	}

	var res int8
	t := tokens.NewEvmToken(node.Rpc, node.Wss, node.Contract, f.Chain, 0)
	if f.Erc == 20 {
		res, err = t.CheckERC20Addresses(adds, subs, common.HexToAddress(f.Contract), big.NewInt(f.Threshold))
	} else if f.Erc == 721 {
		res, err = t.CheckERC721Addresses(adds, subs, common.HexToAddress(f.Contract))
	}
	if err != nil {
		gl.OutLogger.Error("check addresses for group error. %d, %s, %v", f.Chain, f.Contract, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}

	data, _ := json.Marshal(f)
	dataBytes := blake2b.Sum256(data)
	sign, err := service.SignEd25519Hash(dataBytes[:])
	if err != nil {
		gl.OutLogger.Error("service.SignEd25519Hash error. %d, %s, %v", f.Chain, f.Contract, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": true,
		"flag":   res,
		"sign":   hexutil.Encode(sign),
	})
}

func MintNFT(c *gin.Context) {
	address := c.Query("address")
	name := strings.ToLower(c.Query("name"))
	img := c.Query("image")
	if len(img) == 0 {
		img = config.DefaultImg
	}

	prefix, _, err := iotago.ParseBech32(address)
	if prefix != iotago.PrefixShimmer && err != nil {
		gl.OutLogger.Warn("User's address error. %s", address)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "smr address error",
		})
		return
	}

	if len(name) < 11 || len(name) > 20 || !isAlphaNumeric(name) {
		gl.OutLogger.Warn("User's name error. %s", name)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "name invalid",
		})
		return
	}

	b, err := model.VerifyAndInsertName(address, name, config.NameNftId)
	if err != nil {
		gl.OutLogger.Error("model.VerifyAndInsertName error. %s, %s, %v", address, name, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}
	if !b {
		gl.OutLogger.Warn("name used. %s, %s", address, name)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "name used",
		})
		return
	}

	data := make(map[string]string)
	data["standard"] = "IRC27"
	data["name"] = name + ".gf"
	data["type"] = "image/png"
	data["version"] = "v1.0"
	data["uri"] = img
	data["collectionId"] = config.NameNftId
	data["collectionName"] = "GroupFi OG Names"
	data["profile"] = "# GroupFi Name System"
	data["property"] = "groupfi-name"
	meta, _ := json.Marshal(data)

	service.MintNameNft(address, meta)

	c.JSON(http.StatusOK, gin.H{
		"result": true,
	})
}

func RegisterProxy(c *gin.Context) {
	account := c.GetString("account")
	data := strings.Split(c.GetString("data"), "_")
	signAcc := hexutil.Encode(common.FromHex(data[0]))
	if len(account) == 0 || len(signAcc) != 42 {
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "params error. " + account + " " + signAcc,
		})
		return
	}

	smr, err := model.RegisterProxyFromPool(account, signAcc)
	if err != nil {
		gl.OutLogger.Error("model.RegisterProxy error. %s, %s, %v", account, signAcc, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}

	if len(data) > 1 {
		meta := common.FromHex(data[1])
		if len(meta) > 0 {
			id, err := service.MintSignAccPkNft(signAcc, meta)
			gl.OutLogger.Info("pk nft mint. %s, %v", hexutil.Encode(id), err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"result":        true,
		"proxy_account": smr,
	})
}

func GetProxyAccount(c *gin.Context) {
	signAcc := c.GetString("publickey")
	proxy, err := model.GetProxyAccount(signAcc)
	if err != nil {
		gl.OutLogger.Error("model.GetProxyAccount error. %s, %v", signAcc, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}
	if proxy == nil {
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PROXY_NOT_EXIST,
			"err-msg":  "proxy account is not exist",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result":        true,
		"proxy_account": proxy.Smr,
	})
}

func SendTxEssence(c *gin.Context) {
	signAcc := c.GetString("publickey")
	txEssenceBytes := common.FromHex(c.GetString("data"))

	txid, err := service.SendTxEssence(signAcc, txEssenceBytes)
	if err != nil {
		gl.OutLogger.Error("service.SendTxEssence error. %s, %v", signAcc, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.MSG_OUTPUT_ILLEGAL,
			"err-msg":  "output illegal",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":        true,
		"transactionid": "0x" + hex.EncodeToString(txid),
	})
}

func isAlphaNumeric(str string) bool {
	for _, ch := range str {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", ch) {
			return false
		}
	}
	return true
}
