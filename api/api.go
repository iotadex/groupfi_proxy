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
	"github.com/ethereum/go-ethereum/common/hexutil"
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
			service.MintSignAccPkNft(signAcc, meta)
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
