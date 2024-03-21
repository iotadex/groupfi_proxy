package api

import (
	"encoding/json"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/service"
	"gproxy/wallet"
	"net/http"
	"strings"
	"time"

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

	w := wallet.NewIotaSmrWallet(config.SmrRpc, config.MainWallet, config.MainWalletPk, config.NFTID)
	data := make(map[string]string)
	data["standard"] = "IRC27"
	data["name"] = name + ".gf"
	data["type"] = "image/png"
	data["version"] = "v1.0"
	data["uri"] = img
	data["collectionId"] = config.NFTID
	data["collectionName"] = "GroupFi OG Names"
	data["profile"] = "# GroupFi Name System"
	data["property"] = "groupfi-name"
	meta, _ := json.Marshal(data)
	hash, err := w.MintNFT(address, config.Days, meta, []byte("group-id"))
	if err != nil {
		gl.OutLogger.Warn("w.MintNFT error. %s : %v", address, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": 3,
			"err-msg":  "network error when mint nft",
		})
		model.DeleteName(name)
		return
	}

	// store it to db
	go func() {
		time.Sleep(time.Minute * 2)
		nftid, err := w.GetNftOutputFromBlockID(hash)
		if err != nil {
			gl.OutLogger.Error("w.GetNftOutputFromBlockID error. %s, %v", hexutil.Encode(hash), err)
			return
		}
		gl.OutLogger.Info("Mint nft %s, %s, %s", nftid, name, address)

		if err := model.StoreNft(nftid, name, address, hexutil.Encode(hash), config.NFTID); err != nil {
			gl.OutLogger.Error("model.StoreNft error. %s, %s, %s, %v", nftid, hexutil.Encode(hash), name, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"result":   true,
		"block_id": hexutil.Encode(hash),
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
			service.SendSignAccToShimmer(signAcc, meta)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"result":        true,
		"proxy_account": smr,
	})
}

func SendMsgToShimmer(c *gin.Context) {
	signAcc := c.GetString("account")
	meta := common.FromHex(c.GetString("data"))

	if err := service.SendMsgToShimmer(signAcc, meta); err != nil {
		gl.OutLogger.Error("model.RegisterProxy error. %s, %v", signAcc, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": 10,
			"err-msg":  "system error",
		})
		return
	}
}

func isAlphaNumeric(str string) bool {
	for _, ch := range str {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", ch) {
			return false
		}
	}
	return true
}
