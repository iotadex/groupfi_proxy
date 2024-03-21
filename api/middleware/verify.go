package middleware

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"gproxy/gl"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

func VerifySignature(c *gin.Context) {
	//get user's public key
	chain := c.Query("chain")
	data := c.Query("data")
	ts := c.Query("ts")
	sign := c.Query("sign")

	signature := common.FromHex(sign)

	timeStamp, _ := strconv.ParseInt(ts, 10, 64)
	if timeStamp+600 < time.Now().Unix() {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": 0,
			"err-msg":  "sign expired",
		})
		return
	}

	hashData := chain + data + ts
	hashData = fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len([]byte(hashData)), hashData)
	publickey, err := verifyEthAddress(signature, crypto.Keccak256Hash([]byte(hashData)).Bytes())
	addr := crypto.PubkeyToAddress(*publickey).Hex()

	if err != nil {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": 0,
			"err-msg":  "sign error. " + err.Error(),
		})
		gl.OutLogger.Error("User's sign error. %s : %s : %v", addr, data, err)
		return
	}

	c.Set("chain", chain)
	c.Set("account", addr)
	c.Set("publickey", hex.EncodeToString(crypto.FromECDSAPub(publickey)))
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
