package service

import (
	"sync"
	"sync/atomic"
)

var mintNameNftQueue *SendQueue
var mintPkNftQueue *SendQueue
var seeds [4]uint64

func SetSeeds(_seeds [4]uint64) {
	seeds = _seeds
}

// MintMsg
type MintMsg struct {
	Addr       string // mint to address, smr
	NftMeta    []byte // metadata
	NftTag     []byte // tag data
	ExpireDays int    // expire time, days
}

// SendQueue, sending queue
type SendQueue struct {
	queue  []*MintMsg  // metadata queue
	status atomic.Bool // false to stop the queue running
	sync.RWMutex
}

// NewSendQueue
// @rpc		: the shimmer rpc url
// @addr	: the proxy account, a shimmer address
// @enpk 	: the proxy account's private key
// @timeOut : the time of the meta can be remain int he net, a seconds number
// @sendTime: the queue's sending interval time, a seconds number
func NewSendQueue() *SendQueue {
	sq := &SendQueue{
		queue: make([]*MintMsg, 0),
	}
	return sq
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
func (sq *SendQueue) pushBack(msg *MintMsg) bool {
	sq.Lock()
	defer sq.Unlock()
	if sq.status.Load() {
		sq.queue = append(sq.queue, msg)
		return true
	}
	return false
}

// pop a meta from the head of queue
func (sq *SendQueue) pop() *MintMsg {
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
func (sq *SendQueue) pushFront(msg *MintMsg) {
	sq.Lock()
	defer sq.Unlock()
	q := make([]*MintMsg, 0, len(sq.queue)+1)
	q = append(q, msg)
	q = append(q, sq.queue...)
	sq.queue = q
}
