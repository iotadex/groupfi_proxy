package service

import (
	"encoding/hex"
	"gproxy/config"
	"gproxy/gl"
	"gproxy/wallet"
	"sync"
	"sync/atomic"
	"time"
)

var sendingMan *SendMetaManager

func init() {
	sendingMan = NewSendMetaManager(config.SmrRpc, config.QueueExpiredHours, config.SendIntervalTime)
}

// SendMetaManager, manage all the proxy account's sending queue
type SendMetaManager struct {
	rpcUrl     string                // shimmer rpc url
	queues     map[string]*SendQueue // send meta queue for every account, evm account -> sending queue
	activeTime int64                 // hours
	sendTime   int64                 // seconds
	sync.RWMutex
}

// NewSendMetaManager, tHours means the queue's idle time, hours number
func NewSendMetaManager(rpc string, tHours, sendIntervalTime int64) *SendMetaManager {
	smm := &SendMetaManager{
		rpcUrl:     rpc,
		queues:     make(map[string]*SendQueue),
		activeTime: tHours,
		sendTime:   sendIntervalTime,
	}
	go smm.active()
	return smm
}

// active, detect the queue's status, if queue is expired, stop and delete it
func (smm *SendMetaManager) active() {
	ticker := time.NewTicker(time.Hour * time.Duration(smm.activeTime))
	for range ticker.C {
		smm.RLock()
		for acc, sq := range smm.queues {
			if sq.expired(smm.activeTime * 3600) {
				if sq.stop() {
					delete(smm.queues, acc)
				}
			}
		}
		smm.RUnlock()
	}
}

// Push, push a meta to the sending queue
// @acc 	: user's account, (chain + address) of a evm network
// @rpc		: the shimmer rpc url
// @addr	: the proxy account, a shimmer address
// @enpk 	: the proxy account's private key
// @msg		: the metadata and lock time for sending
// @return true or false
func (smm *SendMetaManager) Push(acc, addr, enpk string, msg *MetaMsg) bool {
	var sq *SendQueue
	var exist bool
	smm.RLock()
	sq, exist = smm.queues[acc]
	smm.RUnlock()
	if !exist {
		smm.Lock()
		sq = NewSendQueue(smm.rpcUrl, addr, enpk, smm.sendTime)
		smm.queues[acc] = sq
		smm.Unlock()
	}
	return sq.pushBack(msg)
}

type MetaMsg struct {
	Meta     []byte // metadata
	lockTime int64  // lock time, seconds
}

// SendQueue, sending queue
type SendQueue struct {
	address  string                // smr address
	w        *wallet.IotaSmrWallet // wallet
	queue    []*MetaMsg            // metadata queue
	sendTime int64                 // seconds
	lastTime int64                 // last message sent time
	status   atomic.Bool           // false to stop the queue running
	sync.RWMutex
}

// NewSendQueue
// @rpc		: the shimmer rpc url
// @addr	: the proxy account, a shimmer address
// @enpk 	: the proxy account's private key
// @timeOut : the time of the meta can be remain int he net, a seconds number
// @sendTime: the queue's sending interval time, a seconds number
func NewSendQueue(rpc, addr, enpk string, sendTime int64) *SendQueue {
	w := wallet.NewIotaSmrWallet(rpc, addr, enpk, "0x0")
	sq := &SendQueue{
		address:  addr,
		w:        w,
		lastTime: time.Now().Unix(),
	}
	go sq.run()
	return sq
}

// run, timed transmission
func (sq *SendQueue) run() {
	sq.status.Store(true)
	ticker := time.NewTicker(time.Second * time.Duration(sq.sendTime))
	for range ticker.C {
		if !sq.status.Load() {
			return
		}
		msg := sq.pop()
		if msg == nil {
			continue
		}
		if id, err := sq.w.SendMetaOnly(msg.Meta, msg.lockTime); err != nil {
			gl.OutLogger.Error("sq.wallet.SendMetaOnly error. %s, %v", sq.address, err)
			//push the mete to the queue's front position
			sq.pushFront(msg)
		} else {
			gl.OutLogger.Info("Send : %s : 0x%s", sq.address, hex.EncodeToString(id))
		}
	}
}

// stop sending tx if thq queue is not empty
func (sq *SendQueue) stop() bool {
	sq.Lock()
	defer sq.Unlock()
	if len(sq.queue) > 0 {
		return false
	}
	sq.status.Store(false)
	return true
}

// push a meta to the back of queue and update the lastTime
func (sq *SendQueue) pushBack(msg *MetaMsg) bool {
	sq.Lock()
	defer sq.Unlock()
	if sq.status.Load() {
		sq.queue = append(sq.queue, msg)
		sq.lastTime = time.Now().Unix()
		return true
	}
	return false
}

// pop a meta from the head of queue
func (sq *SendQueue) pop() *MetaMsg {
	sq.Lock()
	defer sq.Unlock()
	if len(sq.queue) == 0 {
		return nil
	}
	m := sq.queue[0]
	sq.queue = sq.queue[1:]
	return m
}

// push a meta to the front of queue
func (sq *SendQueue) pushFront(msg *MetaMsg) {
	sq.Lock()
	defer sq.Unlock()
	q := make([]*MetaMsg, 0, len(sq.queue)+1)
	q = append(q, msg)
	q = append(q, sq.queue...)
	sq.queue = q
}

// judge the queue is expired or not
func (sq *SendQueue) expired(timeOut int64) bool {
	sq.RLock()
	defer sq.RUnlock()
	return time.Now().Unix()-sq.lastTime > timeOut
}
