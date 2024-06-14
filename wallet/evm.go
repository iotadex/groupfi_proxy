package wallet

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"gproxy/tools"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EvmWallet struct {
	url     string
	pk      string
	chainID *big.Int
	client  *ethclient.Client
}

func NewEvmWallet(_url, _pk string, chainid int64) (*EvmWallet, error) {
	client, err := ethclient.Dial(_url)
	if err != nil {
		return nil, err
	}
	return &EvmWallet{
		url:     _url,
		pk:      _pk,
		chainID: big.NewInt(chainid),
		client:  client,
	}, nil
}

func (w *EvmWallet) SendERC20(erc20, to string, amount *big.Int) ([]byte, error) {
	if bytes.Equal(common.HexToAddress(to).Bytes(), common.HexToAddress("0x0000000000000000000000000000000000000000").Bytes()) {
		return nil, fmt.Errorf("to address error. %s", to)
	}
	paddedAddress := common.LeftPadBytes(common.HexToAddress(to).Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	methodid := []byte{0xa9, 0x05, 0x9c, 0xbb} // transfer method
	var data []byte
	data = append(data, methodid...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	value := big.NewInt(0)

	gasPrice, err := w.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	from, prv, err := w.getWalletAddress()
	if err != nil {
		return nil, err
	}

	nonce, err := w.client.PendingNonceAt(context.Background(), *from)
	if err != nil {
		return nil, err
	}
	gasLimit := uint64(3000000)
	tx := types.NewTransaction(nonce, common.HexToAddress(erc20), value, gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(w.chainID), prv)
	if err != nil {
		return nil, err
	}

	if err := w.client.SendTransaction(context.Background(), signedTx); err != nil {
		return nil, err
	}

	return signedTx.Hash().Bytes(), nil
}

func (w *EvmWallet) getWalletAddress() (*common.Address, *ecdsa.PrivateKey, error) {
	pk := tools.Aes.GetDecryptString(w.pk, seeds)
	if pk == nil {
		return nil, nil, fmt.Errorf("wallet pk error")
	}
	prv, err := crypto.HexToECDSA(string(pk))
	if err != nil {
		return nil, nil, fmt.Errorf("wallet bsv pk error. %v", err)
	}
	publicKeyECDSA, ok := prv.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, fmt.Errorf("error casting public key to ECDSA")
	}
	from := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &from, prv, nil
}
