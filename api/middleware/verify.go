package middleware

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gproxy/config"
	"gproxy/gl"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
)

type MetaData struct {
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
	PairXPublicKey      string `json:"pairXPublicKey"`
	EvmAddress          string `json:"evmAddress"`
	Timestamp           int64  `json:"timestamp"`
	Scenery             int    `json:"scenery"`
	Extra               string `json:"extra"`
	Signature           string `json:"signature"`
}

func VerifyEvmSign(c *gin.Context) {
	md := MetaData{}
	if err := c.BindJSON(&md); err != nil {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "not json format",
		})
		return
	}

	if md.Timestamp+600 < time.Now().Unix() {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.TIMEOUT_ERROR,
			"err-msg":  "sign expired",
		})
		return
	}

	signature := common.FromHex(md.Signature)
	data := config.SignPrefix + md.EncryptedPrivateKey + md.EvmAddress + md.PairXPublicKey + strconv.Itoa(md.Scenery) + strconv.FormatInt(md.Timestamp, 10) + md.Extra

	account := ""
	if len(md.EvmAddress) == 42 && (strings.HasPrefix(md.EvmAddress, "0x") || strings.HasPrefix(md.EvmAddress, "0X")) {
		hashData := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
		var err error
		if len(md.Extra) > 4 {
			err = verifyWithExtra(common.FromHex(md.Extra), signature, crypto.Keccak256Hash([]byte(hashData)).Bytes(), common.HexToAddress(md.EvmAddress))
		} else {
			err = verifyEthAddress(signature, crypto.Keccak256Hash([]byte(hashData)).Bytes(), common.HexToAddress(md.EvmAddress))
		}
		if err != nil {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"result":   false,
				"err-code": gl.SIGN_ERROR,
				"err-msg":  "sign error. " + err.Error(),
			})
			slog.Error("User's sign", "postParams", md, "err", err)
			return
		}
		account = md.EvmAddress
	} else {
		publicKey, _ := solana.PublicKeyFromBase58(md.EvmAddress)
		signature := solana.SignatureFromBytes(signature)
		if len(publicKey) != ed25519.PublicKeySize || signature.IsZero() {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"result":   false,
				"err-code": gl.PARAMS_ERROR,
				"err-msg":  "public key error",
			})
			return
		}

		if !publicKey.Verify([]byte(data), signature) {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"result":   false,
				"err-code": gl.SIGN_ERROR,
				"err-msg":  "sign error",
			})
			slog.Error("User's sign error", "postParams", md)
			return
		}
		account = publicKey.String()
	}

	meta, _ := json.Marshal(md)

	c.Set("account", account)
	c.Set("sign_acc", md.PairXPublicKey)
	c.Set("meta", hex.EncodeToString(meta))
	c.Next()
}

func VerifySolSign(c *gin.Context) {
	md := MetaData{}
	if err := c.BindJSON(&md); err != nil {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "not json format",
		})
		return
	}

	if md.Timestamp+600 < time.Now().Unix() {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.TIMEOUT_ERROR,
			"err-msg":  "sign expired",
		})
		return
	}

	publicKey, _ := solana.PublicKeyFromBase58(md.EvmAddress)
	signature, _ := solana.SignatureFromBase58(md.Signature)
	if len(publicKey) != ed25519.PublicKeySize || signature.IsZero() {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "public key error",
		})
		return
	}

	data := md.EncryptedPrivateKey + md.EvmAddress + md.PairXPublicKey + strconv.Itoa(md.Scenery) + strconv.FormatInt(md.Timestamp, 10)

	if !publicKey.Verify([]byte(data), signature) {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SIGN_ERROR,
			"err-msg":  "sign error",
		})
		slog.Error("User's sign error", "postParams", md)
		return
	}

	meta, _ := json.Marshal(md)
	c.Set("account", publicKey.String())
	c.Set("sign_acc", md.PairXPublicKey)
	c.Set("meta", hex.EncodeToString(meta))
	c.Next()
}

type SignData struct {
	PublicKey string `json:"publickey"`
	Data      string `json:"data"`
	Ts        int64  `json:"ts"`
	Sign      string `json:"sign"`
}

func VerifyEd25519Sign(c *gin.Context) {
	sd := SignData{}
	if err := c.BindJSON(&sd); err != nil {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "not json format",
		})
		return
	}

	signature := common.FromHex(sd.Sign)

	if sd.Ts+600 < time.Now().Unix() {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.TIMEOUT_ERROR,
			"err-msg":  "sign expired",
		})
		return
	}

	publicKey := common.FromHex(sd.PublicKey)
	if len(publicKey) != ed25519.PublicKeySize {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "public key error",
		})
		return
	}

	if !ed25519.Verify(publicKey, []byte(sd.Data+strconv.FormatInt(sd.Ts, 10)), signature) {
		c.Abort()
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SIGN_ERROR,
			"err-msg":  "sign error",
		})
		slog.Error("User's sign error", "postParams", sd)
		return
	}

	c.Set("publickey", sd.PublicKey)
	c.Set("data", sd.Data)
	c.Next()
}

func verifyEthAddress(signature, hashData []byte, addr common.Address) error {
	if len(signature) < 65 {
		return fmt.Errorf("signature length is too short")
	}
	if signature[64] < 27 {
		if signature[64] != 0 && signature[64] != 1 {
			return fmt.Errorf("signature error")
		}
	} else {
		signature[64] -= 27
	}
	pubkey, err := crypto.SigToPub(hashData, signature)
	if err != nil {
		return err
	}
	if addr.Cmp(crypto.PubkeyToAddress(*pubkey)) != 0 {
		return fmt.Errorf("signature not match, %s", crypto.PubkeyToAddress(*pubkey).Hex())
	}
	return nil
}

func verifyLuksoAddress(signature, hashData []byte, addr common.Address) error {
	client, err := ethclient.Dial(EvmChains[42].Rpc)
	if err != nil {
		return err
	}
	lukso, err := NewILukso(gl.LUKSO_UP_HELP, client)
	if err != nil {
		return err
	}

	controllerAddresses, err := lukso.GetControllerAddresses(&bind.CallOpts{}, addr)

	for _, cAddr := range controllerAddresses {
		if err = verifyEthAddress(signature, hashData, common.BytesToAddress(cAddr)); err == nil {
			return nil
		}
	}

	return err
}

func verifyWithExtra(extra, signature, hashData []byte, addr common.Address) error {
	kv := make(map[string]interface{})
	if err := json.Unmarshal(extra, &kv); err != nil {
		return err
	}
	if _, exist := kv["lsp"]; exist {
		return verifyLuksoAddress(signature, hashData, addr)
	}
	return fmt.Errorf("error extra. %s", hex.EncodeToString(extra))
}
