package profile

import (
	"encoding/json"
	"gproxy/tools"
)

type LSP3ProfileResult struct {
	LSP3Profile struct {
		Name string `json:"name"`
	} `json:"LSP3Profile"`
	ProfileImageUrl string `json:"profileImageUrl"`
	LSPStandard     string `json:"LSPStandard"`
}

func LuksoProfile(address string, bUpdate bool) (*Did, error) {
	didcache.upMutex.Lock()
	defer didcache.upMutex.Unlock()

	if did, exist := didcache.up[address]; exist && !bUpdate {
		return &did, nil
	}

	baseUrl := "https://api.universalprofile.cloud/v1/42/address/"
	data, err := tools.HttpGet(baseUrl + address)
	if err != nil {
		return nil, err
	}

	result := LSP3ProfileResult{}
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	did := Did{result.LSP3Profile.Name, result.ProfileImageUrl}
	didcache.up[address] = did
	return &did, err
}
