package api

import (
	"encoding/hex"
	"encoding/json"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/service"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	iotago "github.com/iotaledger/iota.go/v3"
)

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
			"err-code": 1,
			"err-msg":  "smr address error",
		})
		return
	}

	if len(name) < 11 || len(name) > 20 || !isAlphaNumeric(name) {
		gl.OutLogger.Warn("User's name error. %s", name)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": 1,
			"err-msg":  "name invalid",
		})
		return
	}
	if !model.VerifyAndInsertName(name) {
		gl.OutLogger.Warn("verify name used. %s", name)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": 2,
			"err-msg":  "name has been used",
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
	chain := c.GetString("chain")
	account := c.GetString("account")
	data := strings.Split(c.GetString("data"), "_")
	signAcc := common.HexToAddress(data[0]).Hex()
	if len(account) == 0 || len(chain) == 0 || len(signAcc) != 42 {
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": 2,
			"err-msg":  "params error. " + account + " " + chain + " " + signAcc,
		})
		return
	}

	smr, err := model.RegisterProxy(chain, account, signAcc)
	if err != nil {
		gl.OutLogger.Error("model.RegisterProxy error. %s, %s, %s, %v", account, chain, signAcc, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": 10,
			"err-msg":  "system error",
		})
		return
	}

	if len(data) > 1 {
		meta := common.FromHex(data[1])
		if len(meta) > 0 {
			service.MintSignAccPkNft(signAcc, meta)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"result":        true,
		"proxy_account": smr,
	})
}

func GetProxyAccount(c *gin.Context) {
	signAcc := c.GetString("account")
	proxy, err := model.GetProxyAccount(signAcc)
	if err != nil {
		gl.OutLogger.Error("model.GetProxyAccount error. %s, %v", signAcc, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": 10,
			"err-msg":  "system error",
		})
		return
	}
	if proxy == nil {
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": 9,
			"err-msg":  "proxy account is not exist",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result":        true,
		"proxy_account": proxy.Smr,
	})
}

func SignMetaMsg(c *gin.Context) {
	signAcc := c.GetString("account")
	meta := common.FromHex(c.GetString("data"))

	signHash, err := service.SignMetaMsg(signAcc, meta)
	if err != nil {
		gl.OutLogger.Error("model.RegisterProxy error. %s, %v", signAcc, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": 10,
			"err-msg":  "system error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":    true,
		"sign_hash": "0x" + hex.EncodeToString(signHash),
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
