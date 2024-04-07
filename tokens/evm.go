package tokens

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// event BuySmr(address indexed user, uint256 amount);
var EventBuySmr = crypto.Keccak256Hash([]byte("BuySmr(address,uint256)"))

type Log struct {
	Type int    //0: infomation log; 1: connect error log; 2: order error log
	Info string // log data
}

type Order struct {
	ChainId string
	TxHash  common.Hash
	User    common.Address
	EdAddr  []byte
	Amount  *big.Int
}

// EvmToken
type EvmToken struct {
	rpc        string
	wss        string
	listenType int            // 0: listen event; 1: scanBlock to read event log
	chainid    string         // smr-evm, bsc-evm
	contract   common.Address // contract of groupfi
	stop       chan bool      // send stop signal
	run        atomic.Bool    // is the scanBlock running
	wg         sync.WaitGroup // listen work to wait
}

func NewEvmToken(_rpc, _wss, _chainid_, _contract string, _listenType int) *EvmToken {
	return &EvmToken{
		rpc:        _rpc,
		wss:        _wss,
		listenType: _listenType,
		chainid:    _chainid_,
		contract:   common.HexToAddress(_contract),
	}
}

func (t *EvmToken) StartListen() (chan Log, chan Order) {
	if t.listenType == 0 {
		return t.listenEvent()
	}
	return t.scanBlock()
}

func (t *EvmToken) listenEvent() (chan Log, chan Order) {
	//Set the query filter
	query := ethereum.FilterQuery{
		Addresses: []common.Address{t.contract},
		Topics:    [][]common.Hash{{EventBuySmr}},
	}

	// connetion err chan
	chLog := make(chan Log, 10)
	chOrder := make(chan Order, 10)

	t.wg.Add(1)
	t.stop = make(chan bool)
	go func() {
		defer t.wg.Done()
	StartFilter:
		// Create the ethclient
		c, err := ethclient.Dial(t.wss)
		if err != nil {
			chLog <- Log{Type: 1, Info: fmt.Sprintf("The EthWssClient redial error(%v). \nThe EthWssClient will be redialed at 5 seconds later...", err)}
			time.Sleep(time.Second * 5)
			goto StartFilter
		}
		eventLogChan := make(chan types.Log)
		sub, err := c.SubscribeFilterLogs(context.Background(), query, eventLogChan)
		if err != nil || sub == nil {
			chLog <- Log{Type: 1, Info: fmt.Sprintf("Get event logs from eth wss client error. %v", err)}
			time.Sleep(time.Second * 5)
			goto StartFilter
		}
		chLog <- Log{Type: 0, Info: fmt.Sprintf("Start to listen %s : %s", t.chainid, t.contract.Hex())}
		for {
			select {
			case err := <-sub.Err():
				chLog <- Log{Type: 1, Info: fmt.Sprintf("Event wss sub error(%v). \nThe EthWssClient will be redialed ...", err)}
				sub.Unsubscribe()
				time.Sleep(time.Second * 5)
				goto StartFilter
			case vLog := <-eventLogChan:
				t.dealEventOrder(&vLog, chOrder)
			case b := <-t.stop:
				if b {
					chLog <- Log{Type: 3, Info: "groupfi listen service stoped"}
					sub.Unsubscribe()
					return
				}
			}
		}
	}()
	return chLog, chOrder
}

func (t *EvmToken) scanBlock() (chan Log, chan Order) {
	c, err := ethclient.Dial(t.rpc)
	if err != nil {
		panic(err)
	}
	fromHeight, err := c.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}

	// Set the query filter
	query := ethereum.FilterQuery{
		Addresses: []common.Address{t.contract},
		Topics:    [][]common.Hash{{EventBuySmr}},
	}

	// connetion err chan
	chLog := make(chan Log, 10)
	chOrder := make(chan Order, 10)

	t.wg.Add(1)
	t.run.Store(true)
	go func() {
		defer t.wg.Done()
		chLog <- Log{Type: 0, Info: fmt.Sprintf("Start to scan %s : %s ...", t.chainid, t.contract.Hex())}
		for t.run.Load() {
			time.Sleep(10 * time.Second)
			var toHeight uint64
			if toHeight, err = c.BlockNumber(context.Background()); err != nil {
				chLog <- Log{Type: 1, Info: fmt.Sprintf("BlockNumber error. %v", err)}
				continue
			} else if toHeight < fromHeight {
				continue
			}

			query.FromBlock = new(big.Int).SetUint64(fromHeight)
			query.ToBlock = new(big.Int).SetUint64(toHeight)
			logs, err := c.FilterLogs(context.Background(), query)
			if err != nil {
				chLog <- Log{Type: 1, Info: fmt.Sprintf("FilterLogs error. %v", err)}
				continue
			}
			for i := range logs {
				t.dealEventOrder(&logs[i], chOrder)
			}
			fromHeight = toHeight + 1
		}
		chLog <- Log{Type: 3, Info: "listen groupfi service stoped"}
	}()
	return chLog, chOrder
}

func (t *EvmToken) StopListen() {
	if t.listenType == 0 {
		t.stop <- true
	} else if t.listenType == 1 {
		t.run.Store(false)
	}
	t.wg.Wait()
}

func (t *EvmToken) dealEventOrder(vLog *types.Log, chOrder chan Order) {
	chOrder <- Order{
		ChainId: t.chainid,
		TxHash:  vLog.TxHash,
		User:    common.BytesToAddress(vLog.Topics[1].Bytes()),
		EdAddr:  vLog.Topics[2].Bytes(),
		Amount:  new(big.Int).SetBytes(vLog.Data),
	}
}

func (t *EvmToken) FilterERC20Addresses(addrs []common.Address, c common.Address, threshold *big.Int) ([]uint16, error) {
	client, err := ethclient.Dial(t.rpc)
	if err != nil {
		return nil, err
	}
	gp, err := NewGroupFi(t.contract, client)
	if err != nil {
		return nil, err
	}
	res, err := gp.FilterERC20Addresses(&bind.CallOpts{}, addrs, c, threshold)
	if err != nil {
		return nil, err
	}
	return res.Indexes[:res.Count], nil
}

func (t *EvmToken) FilterERC721Addresses(addrs []common.Address, c common.Address) ([]uint16, error) {
	client, err := ethclient.Dial(t.rpc)
	if err != nil {
		return nil, err
	}
	gp, err := NewGroupFi(t.contract, client)
	if err != nil {
		return nil, err
	}
	res, err := gp.FilterERC721Addresses(&bind.CallOpts{}, addrs, c)
	if err != nil {
		return nil, err
	}
	return res.Indexes[:res.Count], nil
}

func (t *EvmToken) CheckERC20Addresses(adds, subs []common.Address, c common.Address, threshold *big.Int) (int8, error) {
	client, err := ethclient.Dial(t.rpc)
	if err != nil {
		return 0, err
	}
	gp, err := NewGroupFi(t.contract, client)
	if err != nil {
		return 0, err
	}
	res, err := gp.CheckERC20Group(&bind.CallOpts{}, adds, subs, c, threshold)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (t *EvmToken) CheckERC721Addresses(adds, subs []common.Address, c common.Address) (int8, error) {
	client, err := ethclient.Dial(t.rpc)
	if err != nil {
		return 0, err
	}
	gp, err := NewGroupFi(t.contract, client)
	if err != nil {
		return 0, err
	}
	res, err := gp.CheckERC721Group(&bind.CallOpts{}, adds, subs, c)
	if err != nil {
		return 0, err
	}
	return res, nil
}
