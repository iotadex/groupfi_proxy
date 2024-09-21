package profile

import (
	"gproxy/gl"
	"gproxy/model"
	"log/slog"
	"sync"
)

type Did struct {
	Name     string `json:"name"`
	ImageUrl string `json:"image_url"`
}

type didCache struct {
	mint      map[string]Did
	mintMutex sync.Mutex
	up        map[string]Did
	upMutex   sync.Mutex
}

var didcache = didCache{
	mint: make(map[string]Did),
	up:   make(map[string]Did),
}

func GetAllDids(address string, bUpdate bool) map[uint64]Did {
	dids := make(map[uint64]Did)
	// 1. Get did from lukso up
	if did, err := LuksoProfile(address, bUpdate); err != nil {
		slog.Error("Get profile from lukso", "err", err)
	} else {
		dids[gl.LUKSO_CHAINID] = *did
	}

	if did, err := MintName(address, bUpdate); err != nil {
		slog.Error("Get name from mint", "err", err)
	} else {
		dids[gl.MINT_CHAINID] = *did
	}

	if name, err := model.GetNameByEvmAddress(address); err != nil {
		slog.Error("Get name from db", "err", err)
	} else {
		dids[gl.SHIMMER_CHAINID] = Did{name, ""}
	}
	return dids
}
