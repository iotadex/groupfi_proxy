package service

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/tools"
	"gproxy/wallet"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

func SendTxEssence(signAcc string, txEssenceBytes []byte) ([]byte, error) {
	// get proxy from signAcc
	proxy, err := model.GetProxyAccount(signAcc)
	if err != nil {
		return nil, err
	}
	if proxy == nil {
		return nil, fmt.Errorf("proxy account is not exist")
	}

	// deserialize the transaction essence bytes
	essence := &iotago.TransactionEssence{}
	if _, err := essence.Deserialize(txEssenceBytes, serializer.DeSeriModeNoValidation, nil); err != nil {
		return nil, err
	}

	// verify the outputs
	if len(essence.Outputs) > 2 || len(essence.Outputs) < 1 {
		return nil, fmt.Errorf("illegal essence outputs(%d)", len(essence.Outputs))
	}
	if !verifyMsgOutput(proxy.Smr, essence.Outputs[0]) {
		return nil, fmt.Errorf("illegal essence output(0)")
	}
	if len(essence.Outputs) == 2 {
		if !verifyBalanceOutput(proxy.Smr, essence.Outputs[1]) {
			return nil, fmt.Errorf("illegal essence output(1)")
		}
	}

	// sign the transaction essence
	pk := tools.Aes.GetDecryptString(proxy.EnPk, seeds)
	signature, err := signIotaSmrHashWithPK(essence, pk)
	if err != nil {
		return nil, fmt.Errorf("signIotaSmrHashWithPK error. %v", err)
	}
	tx := newTransaction(essence, signature)
	// send the tx to network
	go func(tx *iotago.Transaction) {
		w := wallet.NewIotaSmrWallet(config.ShimmerRpc, "", "", "0x00")
		if id, err := w.SendSignedTxData(tx); err != nil {
			gl.OutLogger.Error("w.SendSignedTxData error. %s, %v", proxy.Smr, err)
		} else {
			gl.OutLogger.Info("send msg meta output. 0x%s", hex.EncodeToString(id))
		}
	}(tx)

	id, _ := tx.ID()
	return id[:], nil
}

func MintSignAccPkNft(signAcc string, metadata []byte) ([]byte, error) {
	proxy, err := model.GetProxyAccount(signAcc)
	if err != nil {
		return nil, err
	}
	if proxy == nil {
		return nil, fmt.Errorf("proxy account is not exist")
	}

	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, proxy.Smr, proxy.EnPk, "")
	id, err := w.MinPkCollectionNft(proxy.Smr, metadata, []byte(config.ProxyPkNftTag))
	if err != nil {
		return nil, err
	}
	return id, nil
}

func MintNameNft(to string, meta []byte) {
	mintNameNftQueue.pushBack(&MintMsg{
		Addr:       to,
		NftMeta:    meta,
		NftTag:     []byte("group-id"),
		ExpireDays: config.NameNftDays,
	})
}

func SignEd25519Hash(msg []byte) []byte {
	pk := common.FromHex(string(tools.Aes.GetDecryptString(config.SignEdPk, seeds)))
	signature := ed25519.Sign(pk, msg)
	return signature
}

func verifyMsgOutput(to string, op iotago.Output) bool {
	if !verifyBalanceOutput(to, op) {
		return false
	}
	if tl := op.UnlockConditionSet().Timelock(); tl == nil || (int64(tl.UnixTime)-time.Now().Unix()) > config.MaxMsgLockTime {
		return false
	}
	if meta := op.FeatureSet().MetadataFeature(); meta == nil || len(meta.Data) == 0 {
		return false
	}
	return true
}

func verifyBalanceOutput(to string, op iotago.Output) bool {
	if op.Type() != iotago.OutputBasic {
		return false
	}
	if len(op.NativeTokenList()) > 0 {
		return false
	}
	if addr := op.UnlockConditionSet().Address(); addr == nil || to != addr.Address.Bech32(iotago.PrefixShimmer) {
		return false
	}
	if op.UnlockConditionSet().HasTimelockCondition() {
		return false
	}
	if op.UnlockConditionSet().HasStorageDepositReturnCondition() {
		return false
	}
	if op.UnlockConditionSet().HasExpirationCondition() {
		return false
	}
	return true
}

func signIotaSmrHashWithPK(essence *iotago.TransactionEssence, pk []byte) (*iotago.Ed25519Signature, error) {
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

func newTransaction(txEssence *iotago.TransactionEssence, signature iotago.Signature) *iotago.Transaction {
	unlocks := iotago.Unlocks{}
	for i := range txEssence.Inputs {
		if i == 0 {
			unlocks = append(unlocks, &iotago.SignatureUnlock{Signature: signature})
		} else {
			unlocks = append(unlocks, &iotago.ReferenceUnlock{Reference: 0})
		}
	}
	return &iotago.Transaction{Essence: txEssence, Unlocks: unlocks}
}
