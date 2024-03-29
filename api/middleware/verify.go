package middleware

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"gproxy/gl"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

func VerifyEvmSign(c *gin.Context) {
	metaBytes := common.FromHex(c.Query("data"))
	meta := make(map[string]interface{})
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "not json format",
		})
		return
	}

	timeStamp, b := meta["timestamp"].(int64)
	if !b || timeStamp+600 < time.Now().Unix() {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.TIMEOUT_ERROR,
			"err-msg":  "sign expired",
		})
		return
	}

	signature := common.FromHex(meta["signature"].(string))
	encryptedPrivateKey, _ := meta["encryptedPrivateKey"].(string)
	pairXPublicKey, _ := meta["pairXPublicKey"].(string)
	evmAddress, _ := meta["evmAddress"].(string)
	scenery := meta["scenery"].(int64)
	timestamp := meta["timestamp"].(int64)
	data := encryptedPrivateKey + evmAddress + pairXPublicKey + strconv.FormatInt(scenery, 10) + strconv.FormatInt(timestamp, 10)
	data = fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)

	hashData := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	publickey, err := verifyEthAddress(signature, crypto.Keccak256Hash([]byte(hashData)).Bytes())
	addr := crypto.PubkeyToAddress(*publickey).Hex()

	if err != nil {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SIGN_ERROR,
			"err-msg":  "sign error. " + err.Error(),
		})
		gl.OutLogger.Error("User's sign error. %s : %s : %v", addr, data, err)
		return
	}

	c.Set("account", addr)
	c.Set("data", data)
	c.Next()
}

func VerifyEd25519Sign(c *gin.Context) {
	publickey := c.Query("publickey")
	data := c.Query("data")
	ts := c.Query("ts")
	sign := c.Query("sign")

	signature := common.FromHex(sign)

	timeStamp, _ := strconv.ParseInt(ts, 10, 64)
	if timeStamp+600 < time.Now().Unix() {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.TIMEOUT_ERROR,
			"err-msg":  "sign expired",
		})
		return
	}

	if !ed25519.Verify(common.FromHex(publickey), signature, []byte(data+ts)) {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SIGN_ERROR,
			"err-msg":  "sign error",
		})
		gl.OutLogger.Error("User's sign error. %s : %s : %s : %s", publickey, data, ts, sign)
		return
	}

	c.Set("publickey", publickey)
	c.Set("data", data)
	c.Next()
}

func verifyEthAddress(signature, hashData []byte) (*ecdsa.PublicKey, error) {
	if len(signature) < 65 {
		return nil, fmt.Errorf("signature length is too short")
	}
	if signature[64] < 27 {
		if signature[64] != 0 && signature[64] != 1 {
			return nil, fmt.Errorf("signature error")
		}
	} else {
		signature[64] -= 27
	}
	return crypto.SigToPub(hashData, signature)
}
