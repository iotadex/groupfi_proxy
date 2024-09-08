package selfdata

import (
	"encoding/json"
	"fmt"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/tools"
)

type GroupfiDataFilterParam struct {
	Addresses []string `json:"addresses"`
	Threshold string   `json:"threshold"`
	Erc       int      `json:"erc"`
}

type GroupfiDataResult struct {
	Result  bool   `json:"result"`
	Indexes []int  `json:"indexes"`
	ErrCode int    `json:"err-code"`
	ErrMsg  string `json:"err-msg"`
}

func FilterMangoAddresses(addresses []string) ([]int, error) {
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
