package profile

import (
	"encoding/hex"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/tokens"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const mintProfileContract = "0x6D3B3F99177FB2A5de7F9E928a9BD807bF7b5BAD"

func MintName(address string, bUpdate bool) (*Did, error) {
	didcache.mintMutex.Lock()
	defer didcache.mintMutex.Unlock()

	if did, exist := didcache.mint[address]; exist && !bUpdate {
		return &did, nil
	}

	client, err := ethclient.Dial(model.EvmChains[gl.MINT_CHAINID].Rpc)
	if err != nil {
		return nil, err
	}
	mn, err := tokens.NewMintName(common.HexToAddress(mintProfileContract), client)
	if err != nil {
		return nil, err
	}
	name, err := mn.Name(&bind.CallOpts{}, nameHash(address))
	if err != nil {
		return nil, err
	}

	did := Did{name, ""}
	didcache.mint[address] = did

	return &did, err
}

/**
 * @description Hashes ENS name
 *
 * - Since ENS names prohibit certain forbidden characters (e.g. underscore) and have other validation rules, you likely want to [normalize ENS names](https://docs.ens.domains/contract-api-reference/name-processing#normalising-names) with [UTS-46 normalization](https://unicode.org/reports/tr46) before passing them to `namehash`. You can use the built-in [`normalize`](https://viem.sh/docs/ens/utilities/normalize) function for this.
 *
 * @example
 * namehash('wevm.eth')
 * '0xf246651c1b9a6b141d19c2604e9a58f567973833990f830d882534a747801359'
 *
 * @link https://eips.ethereum.org/EIPS/eip-137
 */
func nameHash(address string) [32]byte {
	// address.addr.reverse -> hashName
	result := common.FromHex("0x91d1777781884d03a6757a803996e38de2a42967fb37eeaca72729271025a9e2")
	hashed := crypto.Keccak256Hash([]byte(hex.EncodeToString(common.FromHex(address)))).Bytes()
	result = crypto.Keccak256Hash(append(result, hashed...)).Bytes()

	var res [32]byte
	copy(res[:], result)
	return res
}
