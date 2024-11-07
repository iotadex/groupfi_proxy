package model

import iotago "github.com/iotaledger/iota.go/v3"

type HornetNode struct {
	Id     int    `json:"id"`
	Url    string `json:"url"`
	Weight int    `json:"weight"`
	Info   *iotago.ProtocolParameters
}

func GetHornetNodes() ([]HornetNode, error) {
	rows, err := db.Query("select `id`,`url`,`weight` from `hornet` where `state`=1")
	if err != nil {
		return nil, err
	}
	nodes := make([]HornetNode, 0)
	for rows.Next() {
		n := HornetNode{}
		if err = rows.Scan(&n.Id, &n.Url, &n.Weight); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}
