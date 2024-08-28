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
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

func FilterMangoAddresses(addresses []string) ([]int, error) {
	url := fmt.Sprintf("%s/filter", config.GroupfiDataUri)
	params := GroupfiDataFilterParam{
		Addresses: addresses,
		Threshold: "1",
		Erc:       gl.ERC_MANGO,
	}
	paramsJson, _ := json.Marshal(params)
	data, err := tools.HttpJsonPost(url, paramsJson)
	if err != nil {
		data, err = tools.HttpGet(url)
	}
	if err != nil {
		return nil, err
	}

	var res GroupfiDataResult
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("unmarshal solana rpc result error. %s", string(data))
	}

	if !res.Result {
		return nil, fmt.Errorf("%v", res.ErrMsg)
	}

	return res.Indexes, nil
}
