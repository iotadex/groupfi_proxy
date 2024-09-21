package model

type Chain struct {
	ChainID    uint64 `json:"chainid"`
	Name       string `json:"name"`
	Symbol     string `josn:"symbol"`
	Decimal    int    `json:"decimal"`
	Contract   string `json:"contract"`
	PicUri     string `json:"pic_uri"`
	Rpc        string `json:"-"`
	Wss        string `json:"-"`
	ListenType int    `json:"-"`
	MaxBlock   uint64 `json:"-"`
}

func GetChains() ([]Chain, error) {
	rows, err := db.Query("select `chainid`,`name`,`symbol`,`deci`,`contract`,`pic_uri`,`rpc`,`wss`,`listen`,`max_block` from `chain` where `state`=1")
	if err != nil {
		return nil, err
	}
	chains := make([]Chain, 0)
	for rows.Next() {
		c := Chain{}
		if err = rows.Scan(&c.ChainID, &c.Name, &c.Symbol, &c.Decimal, &c.Contract, &c.PicUri, &c.Rpc, &c.Wss, &c.ListenType, &c.MaxBlock); err != nil {
			return nil, err
		}
		if len(c.PicUri) < 10 {
			c.PicUri = ""
		}
		chains = append(chains, c)
	}
	return chains, nil
}

var EvmChains map[uint64]Chain

func LoadEvmChains() error {
	cs, err := GetChains()
	if err != nil {
		return err
	}

	EvmChains = make(map[uint64]Chain)

	for _, c := range cs {
		EvmChains[c.ChainID] = c
	}
	return nil
}
