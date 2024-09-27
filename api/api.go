package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gproxy/api/selfdata"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/profile"
	"gproxy/service"
	"gproxy/tokens"
	"gproxy/wallet"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	iotago "github.com/iotaledger/iota.go/v3"
)

// Send the test Token to user
func Faucet(c *gin.Context) {
	chainid, err := strconv.ParseUint(c.Query("chainid"), 10, 64)
	token := common.HexToAddress(c.Query("token"))
	to := common.HexToAddress(c.Query("to"))
	amount, b := new(big.Int).SetString(c.Query("amount"), 10)

	if err != nil || !b {
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  "params error.",
		})
		return
	}

	hashTx, err := service.FaucetSend(chainid, token.Hex(), to.Hex(), amount)
	if err != nil || !b {
		slog.Error("service.FaucetSend", "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": true,
		"data":   hexutil.Encode(hashTx),
	})
}

func GetChains(c *gin.Context) {
	update := c.DefaultQuery("update", "0")
	if update != "0" {
		if err := model.LoadEvmChains(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"result":      false,
				gl.ErrCodeStr: gl.SYSTEM_ERROR,
				gl.ErrMsgStr:  err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, model.EvmChains)
}

func GetRpcByChainId(c *gin.Context) {
	chainid, _ := strconv.ParseUint(c.Query("chainid"), 10, 64)
	if chainid != gl.LUKSO_CHAINID {
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  "Don't support this chainid",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": true,
		"rpc":    model.EvmChains[gl.LUKSO_CHAINID].Rpc,
	})
}

// Get the prices of smr on different evm chains
func SmrPrice(c *gin.Context) {
	sps, err := model.GetSmrPrices()
	if err != nil {
		slog.Error("model.GetSmrPrices", "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error",
		})
		return
	}
	for id, p := range sps {
		if c, exist := model.EvmChains[id]; exist {
			p.Contract = c.Contract
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
	Threshold string   `json:"threshold"`
	Erc       int      `json:"erc"`
	Ts        int64    `json:"ts"`
}

func FilterGroup(c *gin.Context) {
	f := Filter{}
	err := c.BindJSON(&f)
	threshold, b := new(big.Int).SetString(f.Threshold, 10)
	node, exist := model.EvmChains[f.Chain]
	if f.Chain == gl.SOLANA_CHAINID {
		exist = true
	}

	if err != nil || !exist || !b {
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  "params error.",
		})
		return
	}

	var indexes []uint16

	if f.Chain == gl.SOLANA_CHAINID {
		indexes, err = selfdata.FilterSolanaAddresses(f.Addresses, f.Contract, threshold.Uint64(), f.Erc)
	} else {
		addrs := make([]common.Address, 0, len(f.Addresses))
		for i := range f.Addresses {
			addrs = append(addrs, common.HexToAddress(f.Addresses[i]))
		}

		t := tokens.NewEvmToken(node.Rpc, node.Wss, node.Contract, f.Chain, 0)
		switch f.Erc {
		case gl.ERC20, gl.ERC404:
			indexes, err = t.FilterERC20Addresses(addrs, common.HexToAddress(f.Contract), threshold)
		case gl.ERC721:
			indexes, err = t.FilterERC721Addresses(addrs, common.HexToAddress(f.Contract))
		case gl.ERC_NATIVE:
			indexes, err = t.FilterEthAddresses(addrs, threshold)
		default:
			slog.Error("protocol error", "erc", f.Erc)
			c.JSON(http.StatusOK, gin.H{
				"result":      false,
				gl.ErrCodeStr: gl.SYSTEM_ERROR,
				gl.ErrMsgStr:  "system error",
			})
			return
		}
	}

	if err != nil {
		slog.Error("Filter addresses from group", "chain", f.Chain, "contract", f.Contract, "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  true,
		"indexes": indexes,
	})
}

type FilterV2 struct {
	Addresses []string `json:"addresses"`
	Chains    []struct {
		Chain     uint64 `json:"chain"`
		Contract  string `json:"contract"`
		Threshold string `json:"threshold"`
		Erc       int    `json:"erc"`
	} `json:"chains"`
	Ts int64 `json:"ts"`
}

// filterGroup for multichain
func FilterGroupV2(c *gin.Context) {
	f, done := filterGroupfiData(c)
	if done {
		return
	}

	// get out the solana addresses
	solAddresses := make([]string, 0, len(f.Addresses))
	evmAddresses := make([]common.Address, 0, len(f.Addresses))
	for _, addr := range f.Addresses {
		if strings.HasPrefix(addr, "0x") || strings.HasPrefix(addr, "0X") {
			solAddresses = append(solAddresses, "")
			evmAddresses = append(evmAddresses, common.HexToAddress(addr))
		} else {
			solAddresses = append(solAddresses, addr)
			evmAddresses = append(evmAddresses, gl.EVM_EMPTY_ADDRESS)
		}
	}

	indexes, err := getEvmBelowIndexes(evmAddresses, f)
	if err != nil {
		slog.Error("Filter addresses from group", "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error",
		})
		return
	}

	inx, err := getSolanaAddresses(solAddresses, indexes, f)
	if err != nil {
		slog.Error("Filter addresses from group", "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  true,
		"indexes": inx,
	})
}

func filterGroupfiData(c *gin.Context) (*FilterV2, bool) {
	f := FilterV2{}
	if err := c.BindJSON(&f); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  err.Error(),
		})
		return nil, true
	}

	done := false
	if len(f.Chains) == 1 {
		var err error
		var indexes []int
		switch f.Chains[0].Erc {
		case gl.ERC_MANGO:
			indexes, err = selfdata.FilterMangoAddresses(f.Addresses)
			done = true
		}

		if done {
			if err != nil {
				slog.Error("Filter addresses from group", "err", err)
				c.JSON(http.StatusOK, gin.H{
					"result":      false,
					gl.ErrCodeStr: gl.SYSTEM_ERROR,
					gl.ErrMsgStr:  "system error",
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"result":  true,
					"indexes": indexes,
				})
			}
		}
	}

	return &f, done
}

func getEvmBelowIndexes(addrs []common.Address, f *FilterV2) ([]bool, error) {
	indexes := make([]bool, len(addrs))
	for _, c := range f.Chains {
		if c.Chain == gl.SOLANA_CHAINID {
			continue
		}
		node, exist := model.EvmChains[c.Chain]
		threshold, b := new(big.Int).SetString(c.Threshold, 10)
		if !exist || !b {
			return nil, fmt.Errorf("chain not exist or threshold error. %d, %s", c.Chain, c.Threshold)
		}
		t := tokens.NewEvmToken(node.Rpc, "", node.Contract, c.Chain, 0)
		var inx []uint16
		var err error
		switch c.Erc {
		case gl.ERC20, gl.ERC404:
			inx, err = t.FilterERC20Addresses(addrs, common.HexToAddress(c.Contract), threshold)
		case gl.ERC721:
			inx, err = t.FilterERC721Addresses(addrs, common.HexToAddress(c.Contract))
		case gl.ERC_NATIVE:
			inx, err = t.FilterEthAddresses(addrs, threshold)
		default:
			err = fmt.Errorf("error erc. %d", c.Erc)
		}

		if err != nil {
			return nil, err
		}
		indexes = getIndexesFromInx(indexes, inx)
	}
	return indexes, nil
}

func getSolanaAddresses(addrs []string, indexes []bool, f *FilterV2) ([]int, error) {
	for _, c := range f.Chains {
		if c.Chain != gl.SOLANA_CHAINID {
			continue
		}
		threhold, _ := strconv.ParseUint(c.Threshold, 10, 64)
		inx, err := selfdata.FilterSolanaAddresses(addrs, c.Contract, threhold, c.Erc)
		if err != nil {
			return nil, err
		}
		indexes = getIndexesFromInx(indexes, inx)
		break
	}
	inx := make([]int, 0)
	for i, b := range indexes {
		if !b {
			inx = append(inx, i)
		}
	}
	return inx, nil
}

func getIndexesFromInx(indexes []bool, inx []uint16) []bool {
	allTrues := make([]bool, len(indexes))
	for i := 0; i < len(indexes); i++ {
		allTrues[i] = true
	}
	for _, i := range inx {
		allTrues[i] = false
	}
	for i := 0; i < len(indexes); i++ {
		indexes[i] = indexes[i] || allTrues[i]
	}
	return indexes
}

type Verfiy struct {
	Chain     uint64   `json:"chain"`
	Adds      []string `json:"adds"`
	Subs      []string `json:"subs"`
	Contract  string   `json:"contract"`
	Threshold string   `json:"threshold"`
	Erc       int      `json:"erc"`
	Ts        int64    `json:"ts"`
}

func VerifyGroup(c *gin.Context) {
	f := Verfiy{}
	err := c.BindJSON(&f)
	threshold, b := new(big.Int).SetString(f.Threshold, 10)
	node, exist := model.EvmChains[f.Chain]
	if f.Chain == gl.SOLANA_CHAINID {
		exist = true
	}

	if err != nil || !exist || !b {
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  fmt.Sprintf("params error. %d, %s, %v", f.Chain, f.Threshold, err),
		})
		return
	}

	var res int8

	if f.Chain == gl.SOLANA_CHAINID {
		if f.Contract == gl.EVM_EMPTY_ADDRESS.Hex() || f.Contract == strings.ToUpper(gl.EVM_EMPTY_ADDRESS.Hex()) {
			f.Contract = gl.SOLANA_EMPTY_PUBKEY.String()
		}
		res, err = selfdata.VerifySolanaAddresses(f.Adds, f.Subs, f.Contract, threshold.Uint64(), f.Erc)
	} else {
		adds := make([]common.Address, 0, len(f.Adds))
		subs := make([]common.Address, 0, len(f.Subs))
		for i := range f.Adds {
			adds = append(adds, common.HexToAddress(f.Adds[i]))
		}
		for i := range f.Subs {
			subs = append(subs, common.HexToAddress(f.Subs[i]))
		}
		t := tokens.NewEvmToken(node.Rpc, node.Wss, node.Contract, f.Chain, 0)
		if f.Erc == 20 {
			res, err = t.CheckERC20Addresses(adds, subs, common.HexToAddress(f.Contract), threshold)
		} else if f.Erc == 721 {
			res, err = t.CheckERC721Addresses(adds, subs, common.HexToAddress(f.Contract))
		} else if f.Erc == 0 {
			res, err = t.CheckEthAddresses(adds, subs, threshold)
		}
	}

	if err != nil {
		slog.Error("check addresses for group", "chain", f.Chain, "contract", f.Contract, "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error",
		})
		return
	}

	data, _ := json.Marshal(f)
	sign, err := wallet.SignEd25519Hash(data[:])
	if err != nil {
		slog.Error("service.SignEd25519Hash", "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": true,
		"flag":   res,
		"sign":   hexutil.Encode(sign),
	})
}

type DidParams struct {
	Addresses []string `json:"addresses"`
	Updates   []bool   `json:"updates"`
}

func GetDids(c *gin.Context) {
	ps := DidParams{}
	if err := c.BindJSON(&ps); (err != nil) || (len(ps.Addresses) != len(ps.Updates)) {
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  err.Error(),
		})
		return
	}

	dids := make(map[string]map[uint64]profile.Did)
	for i, addr := range ps.Addresses {
		dids[addr] = profile.GetAllDids(addr, ps.Updates[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"result": true,
		"dids":   dids,
	})
}

func MintNFT(c *gin.Context) {
	to := c.Query("address")
	name := strings.ToLower(c.Query("name"))
	img := c.Query("image")
	if len(img) == 0 {
		img = config.DefaultImg
	}

	prefix, _, err := iotago.ParseBech32(to)
	if prefix != iotago.PrefixShimmer && err != nil {
		slog.Warn("User's address error", "to", to)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  "smr address error",
		})
		return
	}

	if len(name) < 8 || len(name) > 20 || !isAlphaNumeric(name) {
		slog.Warn("User's name", "name", name)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  "name invalid",
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

	b, err := model.InsertNameNftRecord(to, name, hexutil.Encode(meta), config.NameNftId, config.NameNftDays)
	if err != nil {
		slog.Error("model.InsertNameNftRecord", "to", to, "name", name, "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error",
		})
		return
	}
	if !b {
		slog.Warn("name used", "to", to, "name", name)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  "name used",
		})
		return
	}

	service.MintNameNftSignal()

	c.JSON(http.StatusOK, gin.H{
		"result": true,
	})
}

func CheckName(c *gin.Context) {
	valide, err := model.CheckName(c.Query("n"))
	if err != nil {
		slog.Error("model.CheckName.", "err", err)
	}
	if valide {
		c.JSON(http.StatusOK, gin.H{
			"result": true,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  "name used",
		})
	}
}

func MintNameNftForMM(c *gin.Context) {
	signAcc := c.GetString("publickey")
	name := strings.ToLower(c.GetString("data"))

	proxy, err := model.GetProxyAccount(signAcc)
	if err != nil {
		slog.Error("model.GetProxyAccount", "signAcc", signAcc, "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error",
		})
		return
	}
	if proxy == nil {
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PROXY_NOT_EXIST,
			gl.ErrMsgStr:  "proxy account is not exist",
		})
		return
	}

	if len(name) < 8 || len(name) > 20 || !isAlphaNumeric(name) {
		slog.Warn("User's name error", "name", name)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  "name invalid",
		})
		return
	}

	data := make(map[string]string)
	data["standard"] = "IRC27"
	data["name"] = name + ".gf"
	data["type"] = "image/png"
	data["version"] = "v1.0"
	data["uri"] = config.DefaultImg
	data["collectionId"] = config.NameNftId
	data["collectionName"] = "GroupFi OG Names"
	data["profile"] = "# GroupFi Name System"
	data["property"] = "groupfi-name"
	meta, _ := json.Marshal(data)

	b, err := model.InsertNameNftRecord(proxy.Smr, name, hexutil.Encode(meta), config.NameNftId, 0)
	if err != nil {
		slog.Error("model.VerifyAndInsertName", "smr", proxy.Smr, "name", name, "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error",
		})
		return
	}
	if !b {
		slog.Warn("name used", "smr", proxy.Smr, "name", name)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  "name used",
		})
		return
	}

	service.MintNameNftSignal()

	c.JSON(http.StatusOK, gin.H{
		"result": true,
	})
}

func RegisterProxy(c *gin.Context) {
	account := c.GetString("account")
	signAcc := hexutil.Encode(common.FromHex(c.GetString("sign_acc")))
	meta := c.GetString("meta")
	if len(account) == 0 || len(signAcc) != 66 {
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PARAMS_ERROR,
			gl.ErrMsgStr:  "params error. " + account + " " + signAcc,
		})
		return
	}

	smr, err := model.RegisterProxyFromPool(account, signAcc)
	if err != nil {
		slog.Error("model.RegisterProxy", "account", account, "signAcc", signAcc, "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error",
		})
		return
	}

	txid, outputs, err := service.MintSignAccPkNft(signAcc, common.FromHex(meta))
	if err != nil {
		slog.Error("service.MintSignAccPkNft", "smr", smr, "signAcc", signAcc, "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "mint pk nft error",
		})
		return
	} else {
		slog.Info("mint pk nft", "TransactionId", hexutil.Encode(txid))
	}

	count := len(outputs)
	outputids := make([]string, 0, count)
	for i := 0; i < count; i++ {
		outputid := append(txid[:], byte(i+1))
		outputid = append(outputid, 0)
		outputids = append(outputids, hexutil.Encode(outputid))
	}

	c.JSON(http.StatusOK, gin.H{
		"result":        true,
		"proxy_account": smr,
		"outputids":     outputids,
		"outputs":       outputs,
	})
}

func GetProxyAccount(c *gin.Context) {
	signAcc := c.Query("publickey")
	proxy, err := model.GetProxyAccount(signAcc)
	if err != nil {
		slog.Error("model.GetProxyAccount", "signAcc", signAcc, "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.SYSTEM_ERROR,
			gl.ErrMsgStr:  "system error",
		})
		return
	}
	if proxy == nil {
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.PROXY_NOT_EXIST,
			gl.ErrMsgStr:  "proxy account is not exist",
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

	txid, bid, err := service.SendTxEssence(signAcc, txEssenceBytes, false)
	if err != nil {
		slog.Error("service.SendTxEssence", "signAcc", signAcc, "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.MSG_OUTPUT_ILLEGAL,
			gl.ErrMsgStr:  "output illegal or proxy not exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":        true,
		"transactionid": "0x" + hex.EncodeToString(txid),
		"blockid":       "0x" + hex.EncodeToString(bid),
	})
}

func SendTxEssenceAsyn(c *gin.Context) {
	signAcc := c.GetString("publickey")
	txEssenceBytes := common.FromHex(c.GetString("data"))

	txid, bid, err := service.SendTxEssence(signAcc, txEssenceBytes, true)
	if err != nil {
		slog.Error("service.SendTxEssence", "signAcc", signAcc, "err", err)
		c.JSON(http.StatusOK, gin.H{
			"result":      false,
			gl.ErrCodeStr: gl.MSG_OUTPUT_ILLEGAL,
			gl.ErrMsgStr:  "output illegal or proxy not exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":        true,
		"transactionid": "0x" + hex.EncodeToString(txid),
		"blockid":       "0x" + hex.EncodeToString(bid),
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
