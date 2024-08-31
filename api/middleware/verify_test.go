package middleware

import (
	"context"
	"encoding/hex"
	"fmt"
	"gproxy/tools"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func TestVerifyEth(t *testing.T) {
	// encryptedPrivateKey+evmAddress+pairXPublicKey+scenery+timestamp
	encryptedPrivateKey := "78ececefdabf64909d6b5817f9169e0beb9ec417f5b1cf626de52ad5ef14448b88fe8e49d67a04f2cefda9f4b663d14d241528a541bf5cdb035fe8c16746d5300af69b2844c5f6a98fb979088b3ca531c5c753ee1a282bd74278e45b70f5a53698f10f6d0f86fc049bdf420118ca5ed7"
	pairXPublicKey := "0x4b068c5d502f7d148a5cbc9018d8cf42c9f3ca05e8505d185bb52c7d44dd06b2"
	evmAddress := "0x0439ac5cbc8ae15d19340f398989c1b8b9e78525"
	timestamp := "1721111467"
	scenery := "2"
	data := encryptedPrivateKey + evmAddress + pairXPublicKey + scenery + timestamp
	data = fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	sign := common.FromHex("0x7cccd92dc30a491ada855136ace2d65389357cd5e594629506c1ae121f531e6e7944139863606237a4186f9d5fcf354001b3ae177176a7b29c97c7d9f3616c7d1c")
	err := verifyEthAddress(sign, crypto.Keccak256Hash([]byte(data)).Bytes(), common.HexToAddress(evmAddress))
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

func TestSolanaHello(t *testing.T) {
	client := rpc.New("http://localhost:8899")
	fmt.Println(client.GetBlockHeight(context.Background(), "recent"))
	account := solana.MustPublicKeyFromBase58("BC6ZoMVGS5BuUYZcUiHC1sBP7No4jmJPoJENsLXpsV9A")
	programId := solana.MustPublicKeyFromBase58("3t6momB2NJc4DdveCFoK2VSxaMXLh1ktEfvZa1qWyRhG")
	accMeta := solana.NewAccountMeta(account, true, true)
	instruction := solana.NewInstruction(programId, solana.AccountMetaSlice{accMeta}, nil)

	out, err := client.GetLatestBlockhash(context.Background(), "finalized")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out.Value.Blockhash.String())

	builder := solana.NewTransactionBuilder()
	tx, err := builder.AddInstruction(instruction).SetFeePayer(account).SetRecentBlockHash(out.Value.Blockhash).Build()
	if err != nil {
		t.Fatal(err)
	}

	fn := func(pubKey solana.PublicKey) *solana.PrivateKey {
		prvKey := solana.PrivateKey([]byte{169, 147, 72, 162, 246, 161, 44, 40, 106, 31, 39, 127, 244, 179, 252, 156, 133, 118, 71, 48, 40, 189, 206, 8, 241, 35, 195, 245, 165, 224, 119, 108, 151, 108, 141, 20, 194, 201, 187, 6, 62, 17, 157, 27, 57, 181, 221, 215, 193, 118, 193, 171, 13, 194, 70, 117, 93, 29, 242, 228, 204, 146, 251, 9})
		return &prvKey
	}

	tx.Sign(fn)

	sign, err := client.SendTransaction(context.Background(), tx)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(sign.String())
}

func TestSolanaAccountInfo(t *testing.T) {
	client := rpc.New("https://api.mainnet-beta.solana.com")
	account := solana.MustPublicKeyFromBase58("7AigsDtFL3D5JYMTAC9kh6mZMnnnNXkisREiFD9VqVmv")
	fmt.Println(hex.EncodeToString(account[:]))

	mint := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	conf := rpc.GetTokenAccountsConfig{
		ProgramId: &mint,
	}
	out, err := client.GetTokenAccountsByOwner(context.TODO(), account, &conf, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, ta := range out.Value {
		fmt.Println(ta.Pubkey.String())

		data := ta.Account.Data.GetBinary()
		fmt.Println(hex.EncodeToString(data))

		//var _a token.Account
		/*
			res, err := client.GetAccountInfo(context.Background(), ta.Pubkey)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(res.Value.Data)
		*/
	}
}

func TestCreateTokenAccount(t *testing.T) {
	account := solana.MustPublicKeyFromBase58("AJSyTPw8DsuYrBWXeZW7HNW5gY3fWLEzseJdujg3PryD")
	mint := solana.MustPublicKeyFromBase58("JBY8Ugso1Lge7VTERx1rpwx12hvv6jzDnHEZjMtRDwaR")
	pubkey, a, err := solana.FindAssociatedTokenAddress(account, mint)
	fmt.Println(pubkey.String(), a, err)
}

func TestFilterAddresses(t *testing.T) {
	url := fmt.Sprintf("%s/getTokenAccountBalance?spl=%d&account=%s", "http://solana.groupfi.ai", 1, "Gjmjory7TWKJXD2Jc6hKzAG991wWutFhtbXudzJqgx3p")
	data, err := tools.HttpGet(url)
	fmt.Println(string(data), err)
}

func TestLukso(t *testing.T) {
	data := "hello world"
	hashData := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	signature := common.FromHex("0x66297143ed66e22482bff6e7d42fad4ee9d535cae147cccfbe8a4b1a552a64f51ab14ea20a9dd9dfef5d723303553d103686766721a48c8e88e2f487de34b6841c")
	err := verifyLuksoAddress(signature, crypto.Keccak256Hash([]byte(hashData)).Bytes(), common.HexToAddress("0x8ffd1d75138fba044612549492aD25E9D9456F8E"))
	fmt.Println(err)
}
