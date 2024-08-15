package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/service"
	"gproxy/tokens"
	"gproxy/tools"
	"gproxy/wallet"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gagliardetto/solana-go"
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
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "params error.",
		})
		return
	}

	hashTx, err := service.FaucetSend(chainid, token.Hex(), to.Hex(), amount)
	if err != nil || !b {
		gl.OutLogger.Error("service.FaucetSend error. %v", err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": true,
		"data":   hexutil.Encode(hashTx),
	})
}

func GetChains(c *gin.Context) {
	chains := make(map[uint64][2]string)
	for id, chain := range config.EvmNodes {
		chains[id] = [2]string{chain.Rpc, chain.Contract}
	}

	c.JSON(http.StatusOK, gin.H{
		"chains": chains,
	})
}

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
	Threshold string   `json:"threshold"`
	Erc       int      `json:"erc"`
	Ts        int64    `json:"ts"`
}

func FilterGroup(c *gin.Context) {
	f := Filter{}
	err := c.BindJSON(&f)
	threshold, b := new(big.Int).SetString(f.Threshold, 10)
	node, exist := config.EvmNodes[f.Chain]
	if f.Chain == gl.SOLANA_CHAINID {
		exist = true
	}

	if err != nil || !exist || !b {
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "params error. ",
		})
		return
	}

	var indexes []uint16

	if f.Chain == gl.SOLANA_CHAINID {
		if f.Contract == gl.EVM_EMPTY_ADDRESS.Hex() || f.Contract == strings.ToUpper(gl.EVM_EMPTY_ADDRESS.Hex()) {
			f.Contract = gl.SOLANA_EMPTY_PUBKEY.String()
		}
		if f.Erc == 20 {
			err := associatedTokenAddresses(f.Addresses, f.Contract)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"result":   false,
					"err-code": gl.PARAMS_ERROR,
					"err-msg":  err.Error(),
				})
				return
			}
		}
		indexes, err = filterSolanaAddresses(f.Addresses, f.Contract, threshold.Uint64(), f.Erc)
	} else {
		addrs := make([]common.Address, 0, len(f.Addresses))
		for i := range f.Addresses {
			addrs = append(addrs, common.HexToAddress(f.Addresses[i]))
		}

		t := tokens.NewEvmToken(node.Rpc, node.Wss, node.Contract, f.Chain, 0)
		if f.Erc == 20 {
			indexes, err = t.FilterERC20Addresses(addrs, common.HexToAddress(f.Contract), threshold)
		} else if f.Erc == 721 {
			indexes, err = t.FilterERC721Addresses(addrs, common.HexToAddress(f.Contract))
		} else if f.Erc == 0 {
			indexes, err = t.FilterEthAddresses(addrs, threshold)
		} else {
			gl.OutLogger.Error("erc error. %d", f.Erc)
			c.JSON(http.StatusOK, gin.H{
				"result":   false,
				"err-code": gl.SYSTEM_ERROR,
				"err-msg":  "system error",
			})
			return
		}
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
	f := FilterV2{}
	c.BindJSON(&f)

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

	indexes, err := GetEvmBelowIndexes(evmAddresses, &f)
	if err != nil {
		gl.OutLogger.Error("Filter addresses from group error. %v", err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}
	indexes, err = GetSolanaAddresses(solAddresses, indexes, &f)
	if err != nil {
		gl.OutLogger.Error("Filter addresses from group error. %v", err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}

	inx := make([]int, 0)
	for i, b := range indexes {
		if !b {
			inx = append(inx, i)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  true,
		"indexes": inx,
	})
}

func GetEvmBelowIndexes(addrs []common.Address, f *FilterV2) ([]bool, error) {
	indexes := make([]bool, len(addrs))
	for _, c := range f.Chains {
		if c.Chain == gl.SOLANA_CHAINID {
			continue
		}
		node, exist := config.EvmNodes[c.Chain]
		threshold, b := new(big.Int).SetString(c.Threshold, 10)
		if !exist || !b {
			return nil, fmt.Errorf("chain not exist or threshold error. %d, %s", c.Chain, c.Threshold)
		}
		t := tokens.NewEvmToken(node.Rpc, "", node.Contract, c.Chain, 0)
		var inx []uint16
		var err error
		if c.Erc == 20 {
			inx, err = t.FilterERC20Addresses(addrs, common.HexToAddress(c.Contract), threshold)
		} else if c.Erc == 721 {
			inx, err = t.FilterERC721Addresses(addrs, common.HexToAddress(c.Contract))
		} else if c.Erc == 0 {
			inx, err = t.FilterEthAddresses(addrs, threshold)
		} else {
			err = fmt.Errorf("error erc. %d", c.Erc)
		}
		if err != nil {
			return nil, err
		}
		indexes = getIndexesFromInx(indexes, inx)
	}
	return indexes, nil
}

func GetSolanaAddresses(addrs []string, indexes []bool, f *FilterV2) ([]bool, error) {
	for _, c := range f.Chains {
		if c.Chain != gl.SOLANA_CHAINID {
			continue
		}
		if c.Contract == gl.EVM_EMPTY_ADDRESS.Hex() || c.Contract == strings.ToUpper(gl.EVM_EMPTY_ADDRESS.Hex()) {
			c.Contract = gl.SOLANA_EMPTY_PUBKEY.String()
		}
		if c.Erc == 20 {
			if err := associatedTokenAddresses(addrs, c.Contract); err != nil {
				return indexes, err
			}
		}
		threhold, _ := strconv.ParseUint(c.Threshold, 10, 64)
		inx, err := filterSolanaAddresses(addrs, c.Contract, threhold, c.Erc)
		if err != nil {
			return indexes, err
		}
		indexes = getIndexesFromInx(indexes, inx)
	}
	return indexes, nil
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
	node, exist := config.EvmNodes[f.Chain]
	if f.Chain == gl.SOLANA_CHAINID {
		exist = true
	}

	if err != nil || !exist || !b {
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  fmt.Sprintf("params error. %d, %s, %v", f.Chain, f.Threshold, err),
		})
		return
	}

	var res int8

	if f.Chain == gl.SOLANA_CHAINID {
		if f.Contract == gl.EVM_EMPTY_ADDRESS.Hex() || f.Contract == strings.ToUpper(gl.EVM_EMPTY_ADDRESS.Hex()) {
			f.Contract = gl.SOLANA_EMPTY_PUBKEY.String()
		}
		if f.Contract != gl.SOLANA_EMPTY_PUBKEY.String() {
			err1 := associatedTokenAddresses(f.Adds, f.Contract)
			err2 := associatedTokenAddresses(f.Subs, f.Contract)
			if err1 != nil || err2 != nil {
				c.JSON(http.StatusOK, gin.H{
					"result":   false,
					"err-code": gl.PARAMS_ERROR,
					"err-msg":  fmt.Sprintf("%v, %v", err1, err2),
				})
				return
			}
		}
		res, err = verifySolanaAddresses(f.Adds, f.Subs, f.Contract, threshold.Uint64(), f.Erc)
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
		gl.OutLogger.Error("check addresses for group error. %d, %s, %v", f.Chain, f.Contract, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}

	data, _ := json.Marshal(f)
	sign, err := wallet.SignEd25519Hash(data[:])
	if err != nil {
		gl.OutLogger.Error("service.SignEd25519Hash error. %v", err)
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
	to := c.Query("address")
	name := strings.ToLower(c.Query("name"))
	img := c.Query("image")
	if len(img) == 0 {
		img = config.DefaultImg
	}

	prefix, _, err := iotago.ParseBech32(to)
	if prefix != iotago.PrefixShimmer && err != nil {
		gl.OutLogger.Warn("User's address error. %s", to)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "smr address error",
		})
		return
	}

	if len(name) < 8 || len(name) > 20 || !isAlphaNumeric(name) {
		gl.OutLogger.Warn("User's name error. %s", name)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "name invalid",
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
		gl.OutLogger.Error("model.InsertNameNftRecord error. %s, %s, %v", to, name, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}
	if !b {
		gl.OutLogger.Warn("name used. %s, %s", to, name)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "name used",
		})
		return
	}

	service.MintNameNftSignal()

	c.JSON(http.StatusOK, gin.H{
		"result": true,
	})
}

func MintNameNftForMM(c *gin.Context) {
	signAcc := c.GetString("publickey")
	name := strings.ToLower(c.GetString("data"))

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

	if len(name) < 8 || len(name) > 20 || !isAlphaNumeric(name) {
		gl.OutLogger.Warn("User's name error. %s", name)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "name invalid",
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
		gl.OutLogger.Error("model.VerifyAndInsertName error. %s, %s, %v", proxy.Smr, name, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "system error",
		})
		return
	}
	if !b {
		gl.OutLogger.Warn("name used. %s, %s", proxy.Smr, name)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.PARAMS_ERROR,
			"err-msg":  "name used",
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

	if id, err := service.MintSignAccPkNft(signAcc, common.FromHex(meta)); err != nil {
		gl.OutLogger.Error("service.MintSignAccPkNft error. %s, %s, %v", smr, signAcc, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.SYSTEM_ERROR,
			"err-msg":  "mint pk nft error",
		})
		return
	} else {
		gl.OutLogger.Info("mint pk nft. 0x%s", hex.EncodeToString(id))
	}

	c.JSON(http.StatusOK, gin.H{
		"result":        true,
		"proxy_account": smr,
	})
}

func GetProxyAccount(c *gin.Context) {
	signAcc := c.Query("publickey")
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

	txid, bid, err := service.SendTxEssence(signAcc, txEssenceBytes, false)
	if err != nil {
		gl.OutLogger.Error("service.SendTxEssence error. %s, %v", signAcc, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.MSG_OUTPUT_ILLEGAL,
			"err-msg":  "output illegal or proxy not exist",
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
		gl.OutLogger.Error("service.SendTxEssence error. %s, %v", signAcc, err)
		c.JSON(http.StatusOK, gin.H{
			"result":   false,
			"err-code": gl.MSG_OUTPUT_ILLEGAL,
			"err-msg":  "output illegal or proxy not exist",
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

type Account struct {
	Result    bool   `json:"result"`
	ProgramId string `json:"programid"` // base58
	Owner     string `json:"owner"`     // The owner of this account. base58
	Amount    uint64 `json:"amount"`    // The amount of tokens this account holds.
}

func filterSolanaAddresses(adds []string, programId string, threhold uint64, spl int) ([]uint16, error) {
	indexes := make([]uint16, 0)
	for i := range adds {
		if len(adds[i]) == 0 {
			continue
		}
		url := fmt.Sprintf("%s/getTokenAccountBalance?spl=%d&account=%s&collection=%s", config.SolanaRpc, spl, adds[i], programId)
		data, err := tools.HttpGet(url)
		if err != nil {
			data, err = tools.HttpGet(url)
		}
		if err != nil {
			return nil, err
		}
		var acc Account
		if err := json.Unmarshal(data, &acc); err != nil {
			return nil, fmt.Errorf("unmarshal solana rpc result error. %s", string(data))
		}

		if (acc.ProgramId != programId) || (acc.Amount < threhold) {
			indexes = append(indexes, uint16(i))
		}
	}
	return indexes, nil
}

func verifySolanaAddresses(adds, subs []string, programId string, threhold uint64, spl int) (int8, error) {
	for i := range adds {
		url := fmt.Sprintf("%s/getTokenAccountBalance?spl=%d&account=%s", config.SolanaRpc, spl, adds[i])
		data, err := tools.HttpGet(url)
		if err != nil {
			data, err = tools.HttpGet(url)
		}
		if err != nil {
			return 0, err
		}
		var acc Account
		if err := json.Unmarshal(data, &acc); err != nil {
			return 0, fmt.Errorf("unmarshal solana rpc result error. %s", string(data))
		}
		if (acc.ProgramId != programId) || (acc.Amount < threhold) {
			return 1, nil
		}
	}
	for i := range subs {
		url := config.SolanaRpc + "/getTokenAccountBalance?account=" + subs[i]
		data, err := tools.HttpGet(url)
		if err != nil {
			data, err = tools.HttpGet(url)
		}
		if err != nil {
			return 0, err
		}
		var acc Account
		if err := json.Unmarshal(data, &acc); err != nil {
			return 0, fmt.Errorf("unmarshal solana rpc result error. %s", string(data))
		}

		if (acc.ProgramId != programId) || (acc.Amount >= threhold) {
			return -1, nil
		}
	}
	return 0, nil
}

func associatedTokenAddresses(addrs []string, contract string) error {
	for i, addr := range addrs {
		owner, err1 := solana.PublicKeyFromBase58(addr)
		mint, err2 := solana.PublicKeyFromBase58(contract)
		pubkey, _, err3 := solana.FindAssociatedTokenAddress(owner, mint)
		if err1 != nil || err2 != nil || err3 != nil {

			return fmt.Errorf("addresses params error. %v, %v, %v", err1, err2, err3)
		}
		addrs[i] = pubkey.String()
	}
	return nil
}
