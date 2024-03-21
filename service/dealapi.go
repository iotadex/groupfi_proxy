package service

import (
	"fmt"
	"gproxy/config"
	"gproxy/model"
	"gproxy/tools"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

func SignMetaMsg(signAcc string, txEssenceBytes []byte) ([]byte, error) {
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
	hash, err := essence.SigningMessage()
	if err != nil {
		return nil, fmt.Errorf("essence.SigningMessage error. %v", err)
	}
	pk := tools.Aes.GetDecryptString(proxy.EnPk, seeds)
	signedHash, err := signIotaSmrHashWithPK(hash, pk)
	if err != nil {
		return nil, fmt.Errorf("signIotaSmrHashWithPK error. %v", err)
	}

	return signedHash, nil
}

func MintSignAccPkNft(signAcc string, metadata []byte) error {
	proxy, err := model.GetProxyAccount(signAcc)
	if err != nil {
		return err
	}
	if proxy == nil {
		return fmt.Errorf("proxy account is not exist")
	}

	mintPkNftQueue.pushBack(&MintMsg{
		Addr:       proxy.Smr,
		NftMeta:    metadata,
		NftTag:     []byte("GroupFi-PK"),
		ExpireDays: 0,
	})
	return nil
}

func MintNameNft(to string, meta []byte) {
	mintNameNftQueue.pushBack(&MintMsg{
		Addr:       to,
		NftMeta:    meta,
		NftTag:     []byte("group-id"),
		ExpireDays: config.Days,
	})
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

func signIotaSmrHashWithPK(msg []byte, pk []byte) ([]byte, error) {
	addr := iotago.Ed25519AddressFromPubKey(pk[32:])
	addrKeys := iotago.NewAddressKeysForEd25519Address(&addr, pk)
	signer := iotago.NewInMemoryAddressSigner(addrKeys)
	signature, err := signer.Sign(&addr, msg)
	if err != nil {
		return nil, err
	}
	return signature.(*iotago.Ed25519Signature).Signature[:], nil
}
