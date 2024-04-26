package service

import (
	"encoding/hex"
	"fmt"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/wallet"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

func SendTxEssence(signAcc string, txEssenceBytes []byte, asyn bool) ([]byte, []byte, error) {
	// get proxy from signAcc
	proxy, err := model.GetProxyAccount(signAcc)
	if err != nil {
		return nil, nil, err
	}
	if proxy == nil {
		return nil, nil, fmt.Errorf("proxy account is not exist")
	}

	// deserialize the transaction essence bytes
	essence := &iotago.TransactionEssence{}
	if _, err := essence.Deserialize(txEssenceBytes, serializer.DeSeriModeNoValidation, nil); err != nil {
		return nil, nil, err
	}

	// verify the outputs
	if len(essence.Outputs) < 1 {
		return nil, nil, fmt.Errorf("illegal essence outputs(%d)", len(essence.Outputs))
	}
	for i := 0; i < len(essence.Outputs); i++ {
		if !verifyMsgOutput(proxy.Smr, essence.Outputs[i]) {
			return nil, nil, fmt.Errorf("illegal essence output(%d)", i)
		}
	}

	// sign the transaction essence
	signature, err := wallet.SignIotaSmrHashWithPK(essence, proxy.EnPk)
	if err != nil {
		return nil, nil, fmt.Errorf("signIotaSmrHashWithPK error. %v", err)
	}
	tx := newTransaction(essence, signature)
	txid, _ := tx.ID()

	// send the tx to network
	w := wallet.NewIotaSmrWallet(config.ShimmerRpc, "", "", "0x00")
	var blockId []byte
	if asyn {
		go func() {
			if blockId, err = w.SendSignedTxData(tx); err != nil {
				gl.OutLogger.Error("w.SendSignedTxData error. %s, %v", proxy.Smr, err)
			} else {
				gl.OutLogger.Info("send msg meta output. 0x%s", hex.EncodeToString(blockId))
			}
		}()
	} else {
		if blockId, err = w.SendSignedTxData(tx); err != nil {
			gl.OutLogger.Error("w.SendSignedTxData error. %s, %v", proxy.Smr, err)
		} else {
			gl.OutLogger.Info("send msg meta output. 0x%s", hex.EncodeToString(blockId))
		}
	}

	return txid[:], blockId[:], nil
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

func verifyMsgOutput(to string, op iotago.Output) bool {
	if !verifyBalanceOutput(to, op) {
		return false
	}
	tl := op.UnlockConditionSet().Timelock()
	if tl == nil {
		return true
	}
	if (int64(tl.UnixTime) - time.Now().Unix()) > config.MaxMsgLockTime {
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
	if op.UnlockConditionSet().HasStorageDepositReturnCondition() {
		return false
	}
	if op.UnlockConditionSet().HasExpirationCondition() {
		return false
	}
	return true
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
