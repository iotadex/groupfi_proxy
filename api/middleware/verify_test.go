package middleware

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gagliardetto/solana-go"
)

func TestVerifyEth(t *testing.T) {
	// encryptedPrivateKey+evmAddress+pairXPublicKey+scenery+timestamp
	encryptedPrivateKey := "0x7b2276657273696f6e223a227832353531392d7873616c736132302d706f6c7931333035222c226e6f6e6365223a222f2f546875654e70447859465861746556337232414e4635516b354a4d777044222c22657068656d5075626c69634b6579223a225431304f516a74313736695170394d524d56666f4e47677a67456e39524b33584d447133456b6a335143593d222c2263697068657274657874223a2275625548325137746a524b6d4d74315a63733153364653574542474d4839334d4c546b4f3636364f5a2b61513563744d6e4174624f4b66524e643157374a5a617a52497459464e6f467965475a5863635256654b6b6c744b76326a544a73696468633641343253387943513d227d"
	pairXPublicKey := "0xe9024a1bad751950c768417ad9de3709a67709a40331e79b4e80870b23aa17d6"
	evmAddress := "0x928100571464c900A2F53689353770455D78a200"
	timestamp := "1711445126"
	scenery := "1"
	data := encryptedPrivateKey + evmAddress + pairXPublicKey + scenery + timestamp
	data = fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	sign := common.FromHex("0xebce1073770bd60834071230d8f3faa9d7057a96974d338a355abb568e9fe2435cb85d5b490a9c59cb01fed0d7a7254a36dceb6d835d7b92e236cd82ee6a8db91b")
	_, err := verifyEthAddress(sign, crypto.Keccak256Hash([]byte(data)).Bytes())
	fmt.Println(err)
}

func TestVerifySolana(t *testing.T) {
	encryptedPrivateKey := "67bda84c39bd28da355116973cd6b19c3565d0b7dbced87b7340b2cb1cb69ed38b238eac03d38b644968f17ea72fda74cda8a3c95e225d4ea94274ca5cd486a5841c84baa992470527b363e8a5288e218191535be3e5f1325f51a0e8251cc5f67ff2f4cd5ced3a066879f2ae01eaeddf"
	pairXPublicKey := "0xe02d39be02ee6d40c8dee36198373801548cc55ecd8ddd1dc52482053104c8c2"
	evmAddress := "DUTUWgs4dhaeyvSQkeD5iLBcDZgAD7xDQf2v59iENhso"
	timestamp := "1720601578"
	scenery := "2"

	publicKey, _ := solana.PublicKeyFromBase58(evmAddress)
	sign := "4b90769b23a5d8d5b0da18da0d2a4f5d3da23292170a448401cacbe379c121422fa8f532a061f61f9b3bc9e0919cd731b441f345c55e20969af33344e9aff604"

	signature := solana.SignatureFromBytes(common.FromHex(sign))

	data := encryptedPrivateKey + evmAddress + pairXPublicKey + scenery + timestamp

	fmt.Println(publicKey.Verify([]byte(data), signature))
}
