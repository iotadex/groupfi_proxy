package selfdata

import (
	"encoding/json"
	"fmt"
	"gproxy/config"
	"gproxy/tools"
)

func FilterSolanaAddresses(adds []string, programId string, threhold uint64, spl int) ([]uint16, error) {
	indexes := make([]uint16, 0)
	for i := range adds {
		if len(adds[i]) == 0 {
			continue
		}

		amount, err := getSolanaBalace(adds[i], programId, spl)
		if err != nil {
			return nil, err
		}

		if amount < threhold {
			indexes = append(indexes, uint16(i))
		}
	}
	return indexes, nil
}

func VerifySolanaAddresses(adds, subs []string, programId string, threhold uint64, spl int) (int8, error) {
	for i := range adds {
		amount, err := getSolanaBalace(adds[i], programId, spl)
		if err != nil {
			return 0, err
		}
		if amount < threhold {
			return 1, nil
		}
	}
	for i := range subs {
		amount, err := getSolanaBalace(adds[i], programId, spl)
		if err != nil {
			return 0, err
		}

		if amount >= threhold {
			return -1, nil
		}
	}
	return 0, nil
}

type SolanaBalace struct {
	Result  bool   `json:"result"`
	Amount  uint64 `json:"amount"` // The amount of tokens this account holds.
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

func getSolanaBalace(account, programId string, spl int) (uint64, error) {
	url := fmt.Sprintf("%s/getSolanaBalance?spl=%d&account=%s&programid=%s", config.GroupfiDataUri, spl, account, programId)
	data, err := tools.HttpGet(url)
	if err != nil {
		data, err = tools.HttpGet(url)
	}
	if err != nil {
		return 0, err
	}

	var sb SolanaBalace
	if err := json.Unmarshal(data, &sb); err != nil {
		return 0, fmt.Errorf("unmarshal solana rpc result error. %s", string(data))
	}

	if !sb.Result {
		return 0, fmt.Errorf("%v", sb.ErrMsg)
	}

	return sb.Amount, nil
}
