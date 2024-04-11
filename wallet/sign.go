package wallet

import (
	"crypto/ed25519"
	"fmt"
	"gproxy/config"
	"gproxy/tools"

	"github.com/ethereum/go-ethereum/common"
	iotago "github.com/iotaledger/iota.go/v3"
)

func SignEd25519Hash(msg []byte) ([]byte, error) {
	pk := common.FromHex(string(tools.Aes.GetDecryptString(config.SignEdPk, seeds)))
	if len(pk) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("error private key. length(%d)", len(pk))
	}
	signature := ed25519.Sign(pk, msg)
	return signature, nil
}

func SignIotaSmrHashWithPK(essence *iotago.TransactionEssence, enpk string) (*iotago.Ed25519Signature, error) {
	pk := tools.Aes.GetDecryptString(enpk, seeds)
	if len(pk) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("pk len error. %d", len(pk))
	}
	hash, err := essence.SigningMessage()
	if err != nil {
		return nil, err
	}
	addr := iotago.Ed25519AddressFromPubKey(pk[32:])
	addrKeys := iotago.NewAddressKeysForEd25519Address(&addr, pk)
	signer := iotago.NewInMemoryAddressSigner(addrKeys)
	signature, err := signer.Sign(&addr, hash)
	if err != nil {
		return nil, err
	}
	return signature.(*iotago.Ed25519Signature), nil
}

func ChecKEd25519Addr(enpk, bech32Addr string) bool {
	pk := common.FromHex(string(tools.Aes.GetDecryptString(enpk, seeds)))
	if len(pk) != ed25519.PrivateKeySize {
		return false
	}
	addr := iotago.Ed25519AddressFromPubKey(pk[32:])
	return bech32Addr == addr.Bech32(iotago.PrefixShimmer)
}
