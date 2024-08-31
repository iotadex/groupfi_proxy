package middleware

import "gproxy/model"

var EvmChains map[uint64]model.Chain

func LoadEvmChains() error {
	cs, err := model.GetChains()
	if err != nil {
		return err
	}

	EvmChains = make(map[uint64]model.Chain)

	for _, c := range cs {
		EvmChains[c.ChainID] = c
	}
	return nil
}
