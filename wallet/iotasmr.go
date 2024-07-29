package wallet

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"gproxy/tools"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"golang.org/x/crypto/blake2b"
)

var seeds [4]uint64

func SetSeeds(_seeds [4]uint64) {
	seeds = _seeds
}

type IotaSmrWallet struct {
	nodeAPI  *nodeclient.Client
	outputid []byte
	addr     string
	pk       string
	nftID    iotago.NFTID
}

func NewIotaSmrWallet(_url, _addr, _pk, _nftID string) *IotaSmrWallet {
	w := &IotaSmrWallet{
		nodeAPI: nodeclient.New(_url),
		addr:    _addr,
		pk:      _pk,
	}
	nftID := common.FromHex(_nftID)
	copy(w.nftID[:], nftID)
	return w
}

func (w *IotaSmrWallet) SendMetaOnly(meta []byte, lockedTime int64) ([]byte, error) {
	prefix, toAddr, err := iotago.ParseBech32(w.addr)
	if err != nil {
		return nil, fmt.Errorf("toAddress error. %s, %v", w.addr, err)
	}

	info, err := w.nodeAPI.Info(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get iotasmr node info error. %v", err)
	}

	addr, signer, err := w.getWalletAddress()
	if err != nil {
		return nil, err
	}

	txBuilder := builder.NewTransactionBuilder(info.Protocol.NetworkID())

	output := iotago.BasicOutput{
		Amount: 0,
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{
				Address: toAddr,
			},
			&iotago.TimelockUnlockCondition{
				UnixTime: uint32(time.Now().Unix() + lockedTime),
			},
		},
		Features: iotago.Features{&iotago.MetadataFeature{
			Data: meta,
		}},
	}
	output.Amount = uint64(info.Protocol.RentStructure.VByteCost) * uint64(output.VBytes(&info.Protocol.RentStructure, nil))

	left, newOutput, err := w.getBasiceUnSpentOutputsWithOutputId(txBuilder, output.Amount, prefix, addr)
	if err != nil {
		return nil, err
	}
	txBuilder.AddOutput(&output)
	if left > 0 {
		txBuilder.AddOutput(&iotago.BasicOutput{
			Amount: left,
			Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{
				Address: addr,
			}},
		})
	}

	blockBuilder := txBuilder.BuildAndSwapToBlockBuilder(&info.Protocol, signer, nil)

	block, err := blockBuilder.Tips(context.Background(), w.nodeAPI).
		ProofOfWork(context.Background(), &info.Protocol, float64(info.Protocol.MinPoWScore)).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build block error. %v", err)
	}
	id, err := w.nodeAPI.SubmitBlock(context.Background(), block, &info.Protocol)
	if err != nil {
		return nil, fmt.Errorf("send block to node error. %v", err)
	}

	w.outputid = newOutput
	return id[:], nil
}

func (w *IotaSmrWallet) SendBasic(bech32To string, amount uint64) ([]byte, error) {
	prefix, toAddr, err := iotago.ParseBech32(bech32To)
	if err != nil {
		return nil, fmt.Errorf("toAddress error. %s, %v", bech32To, err)
	}

	info, err := w.nodeAPI.Info(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get iotasmr node info error. %v", err)
	}

	addr, signer, err := w.getWalletAddress()
	if err != nil {
		return nil, err
	}

	txBuilder := builder.NewTransactionBuilder(info.Protocol.NetworkID())

	output := iotago.BasicOutput{
		Amount: amount,
		Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{
			Address: toAddr,
		}},
	}
	left, err := w.getBasiceUnSpentOutputs(txBuilder, amount, prefix, addr)
	if err != nil {
		return nil, err
	}
	txBuilder.AddOutput(&output)
	if left > 0 {
		txBuilder.AddOutput(&iotago.BasicOutput{
			Amount: left,
			Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{
				Address: addr,
			}},
		})
	}
	blockBuilder := txBuilder.BuildAndSwapToBlockBuilder(&info.Protocol, signer, nil)

	block, err := blockBuilder.Tips(context.Background(), w.nodeAPI).
		ProofOfWork(context.Background(), &info.Protocol, float64(info.Protocol.MinPoWScore)).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build block error. %v", err)
	}
	id, err := w.nodeAPI.SubmitBlock(context.Background(), block, &info.Protocol)
	if err != nil {
		return nil, fmt.Errorf("send block to node error. %v", err)
	}

	return id[:], nil
}

func (w *IotaSmrWallet) Recycle(filterTag []byte) ([]byte, error) {
	info, err := w.nodeAPI.Info(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get iotasmr node info error. %v", err)
	}

	addr, signer, err := w.getWalletAddress()
	if err != nil {
		return nil, err
	}

	txBuilder := builder.NewTransactionBuilder(info.Protocol.NetworkID())

	left, err := w.getBasiceTimeoutUnSpentOutputs(txBuilder, iotago.PrefixShimmer, addr, filterTag)
	if err != nil {
		return nil, err
	}
	if left > 0 {
		txBuilder.AddOutput(&iotago.BasicOutput{
			Amount: left,
			Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{
				Address: addr,
			}},
		})
	} else {
		return nil, nil
	}
	blockBuilder := txBuilder.BuildAndSwapToBlockBuilder(&info.Protocol, signer, nil)

	block, err := blockBuilder.Tips(context.Background(), w.nodeAPI).
		ProofOfWork(context.Background(), &info.Protocol, float64(info.Protocol.MinPoWScore)).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build block error. %v", err)
	}
	id, err := w.nodeAPI.SubmitBlock(context.Background(), block, &info.Protocol)
	if err != nil {
		return nil, fmt.Errorf("send block to node error. %v", err)
	}

	return id[:], nil
}

func (w *IotaSmrWallet) MintNameNFT(bech32To string, days int, meta, tag []byte, basicOutput iotago.Output, basicOutputId iotago.OutputID, nftOutput *iotago.NFTOutput, nftOutputId iotago.OutputID, protocol *iotago.ProtocolParameters) ([]byte, error) {
	prefix, toAddr, err := iotago.ParseBech32(bech32To)
	if err != nil {
		return nil, fmt.Errorf("toAddress error. %s, %v", bech32To, err)
	}

	addr, signer, err := w.getWalletAddress()
	if err != nil {
		return nil, err
	}

	txBuilder := builder.NewTransactionBuilder(protocol.NetworkID())
	var collectionInput *iotago.NFTOutput
	if nftOutput != nil {
		collectionInput = nftOutput
		txBuilder.AddInput(&builder.TxInput{UnlockTarget: addr, Input: nftOutput, InputID: nftOutputId})
	} else {
		collectionInput, err = w.getNFTOutput(txBuilder, addr)
		if err != nil {
			return nil, fmt.Errorf("getNFTOutput error. %v", err)
		}
	}
	if collectionInput == nil {
		return nil, fmt.Errorf("collectionInput nil. %s", w.nftID.String())
	}

	var mintOutput iotago.NFTOutput
	if days == 0 {
		mintOutput = w.createBasicNFTOutput(toAddr, meta, tag, protocol)
	} else {
		mintOutput = w.createStorageDepositReturnNameNFTOutput(toAddr, addr, days, meta, tag, protocol)
	}

	collectionOutput := iotago.NFTOutput{
		NFTID: w.nftID,
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: addr},
		},
		ImmutableFeatures: iotago.Features{
			&iotago.MetadataFeature{Data: collectionInput.ImmutableFeatureSet().MetadataFeature().Data},
		},
	}
	collectionOutput.Amount = uint64(protocol.RentStructure.VByteCost) * uint64(collectionOutput.VBytes(&protocol.RentStructure, nil))
	txBuilder.AddOutput(&mintOutput).AddOutput(&collectionOutput)

	needSmrAmount := mintOutput.Amount + collectionOutput.Amount
	if needSmrAmount != collectionInput.Amount {
		var left uint64
		if basicOutput != nil {
			txBuilder.AddInput(&builder.TxInput{UnlockTarget: addr, Input: basicOutput, InputID: basicOutputId})
			left = basicOutput.Deposit()
		} else {
			left, err = w.getBasiceUnSpentOutputs(txBuilder, 0, prefix, addr)
			if err != nil {
				return nil, fmt.Errorf("get basic shimmer outputs error. %s, %v", addr.Bech32(prefix), err)
			}
		}

		left += collectionInput.Amount
		smrOutput := &iotago.BasicOutput{
			Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{
				Address: addr,
			}},
		}
		smrOutput.Amount = uint64(protocol.RentStructure.VByteCost) * uint64(smrOutput.VBytes(&protocol.RentStructure, nil))
		if left < (needSmrAmount + smrOutput.Amount) {
			return nil, fmt.Errorf("balance amount is not enough. %d : %d", needSmrAmount+smrOutput.Amount, left)
		}
		smrOutput.Amount = left - needSmrAmount
		txBuilder.AddOutput(smrOutput)
	}

	blockBuilder := txBuilder.BuildAndSwapToBlockBuilder(protocol, signer, nil)

	block, err := blockBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("build block error. %v", err)
	}
	id, err := w.nodeAPI.SubmitBlock(context.Background(), block, protocol)
	if err != nil {
		return nil, fmt.Errorf("send block to node error. %v", err)
	}

	return id[:], nil
}

func (w *IotaSmrWallet) createStorageDepositReturnNameNFTOutput(toAddr, returnAddr iotago.Address, days int, meta, tag []byte, protocol *iotago.ProtocolParameters) iotago.NFTOutput {
	ts := time.Now().AddDate(0, 0, days).Unix()
	mintOutput := iotago.NFTOutput{
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: toAddr},
			&iotago.StorageDepositReturnUnlockCondition{
				ReturnAddress: returnAddr,
				Amount:        0,
			},
			&iotago.ExpirationUnlockCondition{
				ReturnAddress: returnAddr,
				UnixTime:      uint32(ts),
			},
		},
		ImmutableFeatures: iotago.Features{
			&iotago.IssuerFeature{Address: w.nftID.ToAddress()},
			&iotago.MetadataFeature{Data: meta},
		},
	}
	if len(tag) > 0 {
		mintOutput.Features = iotago.Features{
			&iotago.TagFeature{Tag: tag},
		}
	}
	mintOutput.Amount = uint64(protocol.RentStructure.VByteCost) * uint64(mintOutput.VBytes(&protocol.RentStructure, nil))
	mintOutput.Conditions[1].(*iotago.StorageDepositReturnUnlockCondition).Amount = mintOutput.Amount
	return mintOutput
}

func (w *IotaSmrWallet) createBasicNFTOutput(toAddr iotago.Address, meta, tag []byte, protocol *iotago.ProtocolParameters) iotago.NFTOutput {
	mintOutput := iotago.NFTOutput{
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: toAddr},
		},
		ImmutableFeatures: iotago.Features{
			&iotago.IssuerFeature{Address: w.nftID.ToAddress()},
			&iotago.MetadataFeature{Data: meta},
		},
	}
	if len(tag) > 0 {
		mintOutput.Features = iotago.Features{
			&iotago.TagFeature{Tag: tag},
		}
	}
	mintOutput.Amount = uint64(protocol.RentStructure.VByteCost) * uint64(mintOutput.VBytes(&protocol.RentStructure, nil))
	return mintOutput
}

func (w *IotaSmrWallet) MinPkCollectionNft(bech32To string, meta, tag []byte, bOutput iotago.Output, bOutputID iotago.OutputID, protocol *iotago.ProtocolParameters) ([]byte, error) {
	prefix, toAddr, err := iotago.ParseBech32(bech32To)
	if err != nil {
		return nil, fmt.Errorf("toAddress error. %s, %v", bech32To, err)
	}

	addr, signer, err := w.getWalletAddress()
	if err != nil {
		return nil, err
	}

	collectionOutput := iotago.NFTOutput{
		Conditions: iotago.UnlockConditions{
			&iotago.AddressUnlockCondition{Address: toAddr},
		},
		ImmutableFeatures: iotago.Features{
			&iotago.MetadataFeature{Data: meta},
		},
		Features: iotago.Features{&iotago.TagFeature{Tag: tag}},
	}
	collectionOutput.Amount = uint64(protocol.RentStructure.VByteCost) * uint64(collectionOutput.VBytes(&protocol.RentStructure, nil))

	txBuilder := builder.NewTransactionBuilder(protocol.NetworkID())
	txBuilder.AddOutput(&collectionOutput)

	var left uint64
	if bOutput != nil {
		txBuilder.AddInput(&builder.TxInput{UnlockTarget: addr, Input: bOutput, InputID: bOutputID})
		left = bOutput.Deposit() - collectionOutput.Amount
	} else {
		left, err = w.getBasiceUnSpentOutputs(txBuilder, collectionOutput.Amount, prefix, addr)
		if err != nil {
			return nil, fmt.Errorf("get basic shimmer outputs error. %s, %v", addr.Bech32(prefix), err)
		}
	}

	if left > 0 {
		smrOutput := &iotago.BasicOutput{
			Amount: left,
			Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{
				Address: addr,
			}},
		}
		txBuilder.AddOutput(smrOutput)
	}

	blockBuilder := txBuilder.BuildAndSwapToBlockBuilder(protocol, signer, nil)

	block, err := blockBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("build block error. %v", err)
	}
	id, err := w.nodeAPI.SubmitBlock(context.Background(), block, protocol)
	if err != nil {
		return nil, fmt.Errorf("send block to node error. %v", err)
	}

	return id[:], nil
}

func (w *IotaSmrWallet) CollectBackTimeLocked(lockedTime int64) (bool, error) {
	_, toAddr, err := iotago.ParseBech32(w.addr)
	if err != nil {
		return false, fmt.Errorf("toAddress error. %s, %v", w.addr, err)
	}

	info, err := w.nodeAPI.Info(context.Background())
	if err != nil {
		return false, fmt.Errorf("get iotasmr node info error. %v", err)
	}

	addr, signer, err := w.getWalletAddress()
	if err != nil {
		return false, err
	}

	txBuilder := builder.NewTransactionBuilder(info.Protocol.NetworkID())
	left, err := w.getTimelockOutputs(txBuilder, lockedTime, addr)
	if err != nil {
		return false, err
	}
	if left == 0 {
		return true, nil
	}

	output := iotago.BasicOutput{
		Amount:     left,
		Conditions: iotago.UnlockConditions{&iotago.AddressUnlockCondition{Address: toAddr}},
	}
	txBuilder.AddOutput(&output)

	blockBuilder := txBuilder.BuildAndSwapToBlockBuilder(&info.Protocol, signer, nil)

	block, err := blockBuilder.Tips(context.Background(), w.nodeAPI).
		ProofOfWork(context.Background(), &info.Protocol, float64(info.Protocol.MinPoWScore)).
		Build()
	if err != nil {
		return false, fmt.Errorf("build block error. %v", err)
	}
	if _, err = w.nodeAPI.SubmitBlock(context.Background(), block, &info.Protocol); err != nil {
		return false, fmt.Errorf("send block to node error. %v", err)
	}
	return false, nil
}

func (w *IotaSmrWallet) CheckTx(blockId []byte) (bool, error) {
	bid := iotago.EmptyBlockID()
	if len(blockId) != 32 {
		return true, fmt.Errorf("txid error. 0x%s", hex.EncodeToString(blockId))
	}
	copy(bid[:], blockId)

	res, err := w.nodeAPI.BlockMetadataByBlockID(context.Background(), bid)
	if err != nil {
		return true, err
	}
	if !res.Solid {
		return true, fmt.Errorf("txid has not solid. 0x%s", hex.EncodeToString(blockId))
	}
	if res.ConflictReason != 0 {
		return false, fmt.Errorf("%d : %s", res.ConflictReason, res.LedgerInclusionState)
	}
	return true, nil
}

func (w *IotaSmrWallet) GetUnspentOutput(addr string) (iotago.Output, iotago.OutputID, error) {
	indexer, err := w.nodeAPI.Indexer(context.Background())
	if err != nil {
		return nil, iotago.OutputID{}, err
	}

	notHas := false
	query := nodeclient.BasicOutputsQuery{
		AddressBech32: addr,
		IndexerNativeTokenParas: nodeclient.IndexerNativeTokenParas{
			HasNativeTokens: &notHas,
		},
		IndexerTimelockParas: nodeclient.IndexerTimelockParas{
			HasTimelock: &notHas,
		},
		IndexerExpirationParas: nodeclient.IndexerExpirationParas{
			HasExpiration: &notHas,
		},
		IndexerStorageDepositParas: nodeclient.IndexerStorageDepositParas{
			HasStorageDepositReturn: &notHas,
		},
	}
	res, err := indexer.Outputs(context.Background(), &query)
	if err != nil {
		return nil, iotago.OutputID{}, err
	}
	for res.Next() {
		ids, err := res.Response.Items.OutputIDs()
		if err != nil {
			return nil, iotago.OutputID{}, err
		}

		outputs, _ := res.Outputs()
		var o iotago.Output
		var id iotago.OutputID
		var max uint64
		for i, output := range outputs {
			if max < output.Deposit() {
				max = output.Deposit()
				o, id = output, ids[i]
			}
		}
		if max > 0 {
			return o, id, nil
		}
	}
	return nil, iotago.OutputID{}, nil
}

func (w *IotaSmrWallet) GetCollectionNFTOutput() (*iotago.NFTOutput, iotago.OutputID, error) {
	indexer, err := w.nodeAPI.Indexer(context.Background())
	if err != nil {
		return nil, iotago.OutputID{}, err
	}

	outputID, nftOutput, _, err := indexer.NFT(context.Background(), w.nftID)
	if err != nil {
		return nil, iotago.OutputID{}, err
	}

	return nftOutput, *outputID, nil
}

func (w *IotaSmrWallet) SendSignedTxData(tx *iotago.Transaction) ([]byte, error) {
	info, err := w.nodeAPI.Info(context.Background())
	if err != nil {
		return nil, err
	}

	blockBuilder, err := NewBlockBuilder(&info.Protocol, tx)
	if err != nil {
		return nil, err
	}

	block, err := blockBuilder.Tips(context.Background(), w.nodeAPI).
		ProofOfWork(context.Background(), &info.Protocol, float64(info.Protocol.MinPoWScore)).
		Build()
	if err != nil {
		return nil, err
	}
	id, err := w.nodeAPI.SubmitBlock(context.Background(), block, &info.Protocol)
	if err != nil {
		return nil, err
	}
	return id[:], err
}

func (w *IotaSmrWallet) SendSignedTxDataWithoutPow(tx *iotago.Transaction, protocol *iotago.ProtocolParameters) ([]byte, error) {
	blockBuilder, err := NewBlockBuilder(protocol, tx)
	if err != nil {
		return nil, err
	}

	block, err := blockBuilder.Build()
	if err != nil {
		return nil, err
	}
	id, err := w.nodeAPI.SubmitBlock(context.Background(), block, protocol)
	if err != nil {
		return nil, err
	}
	return id[:], err
}

func (w *IotaSmrWallet) Balance(bech32TAddr string) (uint64, error) {
	indexer, err := w.nodeAPI.Indexer(context.Background())
	if err != nil {
		return 0, err
	}

	notHas := false
	query := nodeclient.BasicOutputsQuery{
		AddressBech32: bech32TAddr,
		IndexerNativeTokenParas: nodeclient.IndexerNativeTokenParas{
			HasNativeTokens: &notHas,
		},
		IndexerTimelockParas: nodeclient.IndexerTimelockParas{
			HasTimelock: &notHas,
		},
		IndexerExpirationParas: nodeclient.IndexerExpirationParas{
			HasExpiration: &notHas,
		},
		IndexerStorageDepositParas: nodeclient.IndexerStorageDepositParas{
			HasStorageDepositReturn: &notHas,
		},
	}
	res, err := indexer.Outputs(context.Background(), &query)
	if err != nil {
		return 0, err
	}
	sum := uint64(0)
	for res.Next() {
		outputs, _ := res.Outputs()
		for _, output := range outputs {
			sum += output.Deposit()
		}
	}
	return sum, nil
}

func (w *IotaSmrWallet) getWalletAddress() (iotago.Address, iotago.AddressSigner, error) {
	pk, err := hex.DecodeString(string(tools.Aes.GetDecryptString(w.pk, seeds)))
	if len(pk) != 64 || err != nil {
		return nil, nil, fmt.Errorf("wallet iotasmr pk error")
	}
	addr := iotago.Ed25519AddressFromPubKey(pk[32:])
	addrKeys := iotago.NewAddressKeysForEd25519Address(&addr, pk)
	signer := iotago.NewInMemoryAddressSigner(addrKeys)
	return &addr, signer, nil
}

func (w *IotaSmrWallet) getBasiceUnSpentOutputs(b *builder.TransactionBuilder, amount uint64, prefix iotago.NetworkPrefix, addr iotago.Address) (uint64, error) {
	indexer, err := w.nodeAPI.Indexer(context.Background())
	if err != nil {
		return 0, err
	}

	notHas := false
	query := nodeclient.BasicOutputsQuery{
		AddressBech32: addr.Bech32(prefix),
		IndexerNativeTokenParas: nodeclient.IndexerNativeTokenParas{
			HasNativeTokens: &notHas,
		},
		IndexerTimelockParas: nodeclient.IndexerTimelockParas{
			HasTimelock: &notHas,
		},
		IndexerExpirationParas: nodeclient.IndexerExpirationParas{
			HasExpiration: &notHas,
		},
		IndexerStorageDepositParas: nodeclient.IndexerStorageDepositParas{
			HasStorageDepositReturn: &notHas,
		},
	}
	res, err := indexer.Outputs(context.Background(), &query)
	if err != nil {
		return 0, err
	}
	sum := uint64(0)
	for res.Next() {
		ids, err := res.Response.Items.OutputIDs()
		if err != nil {
			return 0, err
		}

		outputs, _ := res.Outputs()
		for i, output := range outputs {
			if len(output.NativeTokenList()) > 0 {
				continue
			}
			b.AddInput(&builder.TxInput{UnlockTarget: addr, Input: output, InputID: ids[i]})
			sum += output.Deposit()
			if sum >= 2*amount {
				break
			}
		}
	}
	if sum < amount {
		return amount, fmt.Errorf("balance amount is not enough")
	}
	return sum - amount, nil
}

func (w *IotaSmrWallet) getBasiceUnSpentOutputsWithOutputId(b *builder.TransactionBuilder, amount uint64, prefix iotago.NetworkPrefix, addr iotago.Address) (uint64, []byte, error) {
	indexer, err := w.nodeAPI.Indexer(context.Background())
	if err != nil {
		return 0, nil, err
	}

	notHas := false
	query := nodeclient.BasicOutputsQuery{
		AddressBech32: addr.Bech32(prefix),
		IndexerNativeTokenParas: nodeclient.IndexerNativeTokenParas{
			HasNativeTokens: &notHas,
		},
		IndexerTimelockParas: nodeclient.IndexerTimelockParas{
			HasTimelock: &notHas,
		},
		IndexerExpirationParas: nodeclient.IndexerExpirationParas{
			HasExpiration: &notHas,
		},
		IndexerStorageDepositParas: nodeclient.IndexerStorageDepositParas{
			HasStorageDepositReturn: &notHas,
		},
	}
	res, err := indexer.Outputs(context.Background(), &query)
	if err != nil {
		return 0, nil, err
	}
	sum := uint64(0)
	count := 0
	var newOutput []byte
	for res.Next() {
		ids, err := res.Response.Items.OutputIDs()
		if err != nil {
			return 0, nil, err
		}
		outputs, _ := res.Outputs()
		for i, output := range outputs {
			if len(output.NativeTokenList()) > 0 {
				continue
			}
			if t := output.UnlockConditionSet().Timelock(); t != nil {
				if t.UnixTime > uint32(time.Now().Unix()) {
					continue
				}
			}
			if bytes.Equal(ids[i][:], w.outputid) {
				return 0, nil, fmt.Errorf("output is old")
			}
			b.AddInput(&builder.TxInput{UnlockTarget: addr, Input: output, InputID: ids[i]})
			if len(newOutput) == 0 {
				newOutput = ids[i][:]
			}
			sum += output.Deposit()
			count++
			if count >= 64 {
				break
			}
		}
	}
	if sum < amount {
		return amount, nil, fmt.Errorf("balance amount is not enough")
	}
	return sum - amount, newOutput, nil
}

func (w *IotaSmrWallet) getBasiceTimeoutUnSpentOutputs(b *builder.TransactionBuilder, prefix iotago.NetworkPrefix, addr iotago.Address, filterTag []byte) (uint64, error) {
	indexer, err := w.nodeAPI.Indexer(context.Background())
	if err != nil {
		return 0, err
	}

	notHas := false
	has := true
	nowTs := uint32(time.Now().Unix())
	query := nodeclient.BasicOutputsQuery{
		AddressBech32: addr.Bech32(prefix),
		IndexerNativeTokenParas: nodeclient.IndexerNativeTokenParas{
			HasNativeTokens: &notHas,
		},
		IndexerTimelockParas: nodeclient.IndexerTimelockParas{
			HasTimelock:      &has,
			TimelockedBefore: nowTs,
		},
		IndexerExpirationParas: nodeclient.IndexerExpirationParas{
			HasExpiration: &notHas,
		},
		IndexerStorageDepositParas: nodeclient.IndexerStorageDepositParas{
			HasStorageDepositReturn: &notHas,
		},
	}
	res, err := indexer.Outputs(context.Background(), &query)
	if err != nil {
		return 0, err
	}
	extParas := iotago.ExternalUnlockParameters{
		ConfUnix: nowTs,
	}
	sum := uint64(0)
	for res.Next() {
		ids, err := res.Response.Items.OutputIDs()
		if err != nil {
			return 0, err
		}

		outputs, _ := res.Outputs()
		for i, output := range outputs {
			if len(output.NativeTokenList()) > 0 {
				continue
			}
			if output.UnlockConditionSet().TimelocksExpired(&extParas) != nil {
				continue
			}
			tag := output.FeatureSet().TagFeature()
			if tag != nil && bytes.Equal(tag.Tag, filterTag) {
				continue
			}
			b.AddInput(&builder.TxInput{UnlockTarget: addr, Input: output, InputID: ids[i]})
			sum += output.Deposit()
		}
	}
	return sum, nil
}

func (w *IotaSmrWallet) getNFTOutput(b *builder.TransactionBuilder, addr iotago.Address) (*iotago.NFTOutput, error) {
	indexer, err := w.nodeAPI.Indexer(context.Background())
	if err != nil {
		return nil, err
	}

	outputID, nftOutput, _, err := indexer.NFT(context.Background(), w.nftID)
	if err != nil {
		return nil, err
	}
	b.AddInput(&builder.TxInput{UnlockTarget: addr, Input: nftOutput, InputID: *outputID})

	return nftOutput, nil
}

func (w *IotaSmrWallet) getTimelockOutputs(b *builder.TransactionBuilder, lockedtime int64, addr iotago.Address) (uint64, error) {
	indexer, err := w.nodeAPI.Indexer(context.Background())
	if err != nil {
		return 0, err
	}

	notHas := false
	has := true
	query := nodeclient.BasicOutputsQuery{
		AddressBech32: w.addr,
		IndexerNativeTokenParas: nodeclient.IndexerNativeTokenParas{
			HasNativeTokens: &notHas,
		},
		IndexerTimelockParas: nodeclient.IndexerTimelockParas{
			HasTimelock:      &has,
			TimelockedBefore: uint32(time.Now().Unix() - lockedtime),
		},
		IndexerExpirationParas: nodeclient.IndexerExpirationParas{
			HasExpiration: &notHas,
		},
		IndexerStorageDepositParas: nodeclient.IndexerStorageDepositParas{
			HasStorageDepositReturn: &notHas,
		},
	}
	res, err := indexer.Outputs(context.Background(), &query)
	if err != nil {
		return 0, err
	}
	sum := uint64(0)
	count := 0
	for res.Next() {
		ids, err := res.Response.Items.OutputIDs()
		if err != nil {
			return 0, err
		}
		outputs, _ := res.Outputs()
		for i, output := range outputs {
			b.AddInput(&builder.TxInput{UnlockTarget: addr, Input: output, InputID: ids[i]})
			sum += output.Deposit()
			count++
			if count >= 64 {
				break
			}
		}
	}
	return sum, nil
}

func (w *IotaSmrWallet) GetNftOutputFromBlockID(id []byte) (string, error) {
	info, err := w.nodeAPI.Info(context.Background())
	if err != nil {
		return "", err
	}
	var blockID iotago.BlockID
	copy(blockID[:], id)
	block, err := w.nodeAPI.BlockByBlockID(context.Background(), blockID, &info.Protocol)
	if err != nil {
		return "", err
	}
	tx := block.Payload.(*iotago.Transaction)
	outputs := tx.Essence.Outputs
	nftOutput, b := outputs[0].(*iotago.NFTOutput)
	if !b {
		return "", fmt.Errorf("iotago.Output is not *iotago.NFTOutput. %s", blockID.String())
	}
	if !nftOutput.NFTID.Empty() {
		return nftOutput.NFTID.String(), nil
	}
	txid, _ := tx.ID()
	sum := blake2b.Sum256(common.FromHex(txid.ToHex() + "0000"))
	return hexutil.Encode(sum[:]), nil
}

// NewBlockBuilder builds the transaction with signature and then swaps to a BlockBuilder with
// the transaction set as its payload.
func NewBlockBuilder(protoParas *iotago.ProtocolParameters, tx *iotago.Transaction) (*builder.BlockBuilder, error) {
	if _, err := tx.Serialize(serializer.DeSeriModePerformValidation, protoParas); err != nil {
		return nil, err
	}
	blockBuilder := builder.NewBlockBuilder()
	return blockBuilder.ProtocolVersion(protoParas.Version).Payload(tx), nil
}
