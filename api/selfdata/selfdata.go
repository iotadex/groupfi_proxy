package selfdata

import (
	"encoding/json"
	"fmt"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/tools"

	"github.com/ethereum/go-ethereum/common"
)

type GroupfiDataFilterParam struct {
	Chain     uint64   `json:"chain"`
	Contract  string   `json:"contract"`
	Addresses []string `json:"addresses"`
	Threshold string   `json:"threshold"`
	Erc       int      `json:"erc"`
}

type GroupfiDataResult struct {
	Result  bool     `json:"result"`
	Indexes []uint16 `json:"indexes"`
	ErrCode int      `json:"err-code"`
	ErrMsg  string   `json:"err-msg"`
}

func FilterMangoAddresses(addresses []string) ([]uint16, error) {
	url := fmt.Sprintf("%s/filter", config.GroupfiDataUri)
	params := GroupfiDataFilterParam{
		Addresses: addresses,
		Threshold: "1",
		Erc:       gl.ERC_MANGO,
	}
	data, err := tools.HttpJsonPost(url, params)
	if err != nil {
		data, err = tools.HttpJsonPost(url, params)
	}
	if err != nil {
		return nil, err
	}

	var res GroupfiDataResult
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("unmarshal solana rpc result error. %s", string(data))
	}

	if !res.Result {
		return nil, fmt.Errorf("%s", string(data))
	}

	return res.Indexes, nil
}

func FilterNftAddresses(chainid uint64, contract, threshhold string, erc int, addresses []common.Address) ([]uint16, error) {
	url := fmt.Sprintf("%s/filter", config.GroupfiDataUri)
	addrs := make([]string, 0, len(addresses))
	for _, addr := range addresses {
		addrs = append(addrs, addr.Hex())
	}
	params := GroupfiDataFilterParam{
		Chain:     chainid,
		Contract:  contract,
		Addresses: addrs,
		Threshold: threshhold,
		Erc:       erc,
	}
	data, err := tools.HttpJsonPost(url, params)
	if err != nil {
		data, err = tools.HttpJsonPost(url, params)
	}
	if err != nil {
		return nil, err
	}

	var res GroupfiDataResult
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("unmarshal solana rpc result error. %s", string(data))
	}

	if !res.Result {
		return nil, fmt.Errorf("%s", string(data))
	}

	return res.Indexes, nil
}
