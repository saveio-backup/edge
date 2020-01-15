package peer

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/saveio/carrier/network"
	"github.com/saveio/dsp-go-sdk/utils"
	"github.com/saveio/edge/common"
	"github.com/saveio/themis/common/log"
)

type MsgState int

const (
	MsgStateNone MsgState = iota
	MsgStateSending
	MsgStateSended
)

type MsgReply struct {
	id  string
	msg proto.Message
	err error
}

type MsgWrap struct {
	id        string
	msg       proto.Message
	state     MsgState
	needReply bool
	reply     chan MsgReply
	createdAt uint64
}

type FailedCount struct {
	Dial       int
	Send       int
	Recv       int
	Disconnect int
}

type ConnectState int

const (
	ConnectStateNone ConnectState = iota
	ConnectStateConnecting
	ConnectStateConnected
	ConnectStateFailed
	ConnectStateClose
)

type Peer struct {
	addr        string              // peer address
	client      *network.PeerClient // network overlay peer client instance
	mq          *list.List          // message queue list
	lock        *sync.RWMutex       // lock for sync operation
	retry       map[string]int      // count recorder for retry send msg
	retrying    bool                // flag of service is retrying
	closeTime   time.Time           // peer closed time
	failedCount *FailedCount        // peer QoS failed count
	state       ConnectState        // connect state
	receivedMsg *lru.ARCCache
}

func New(addr string) *Peer {
	cache, _ := lru.NewARC(common.MAX_RECEIVED_MSG_CACHE)
	p := &Peer{
		addr:        addr,
		mq:          list.New(),
		lock:        new(sync.RWMutex),
		failedCount: new(FailedCount),
		retry:       make(map[string]int),
		receivedMsg: cache,
	}
	return p
}

// SetClient. set peer client.
func (p *Peer) SetClient(client *network.PeerClient) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.client != nil {
		return
	}
	p.receivedMsg.Purge()
	p.client = client
	go p.acceptAckNotify()
	if p.mq.Len() == 0 {
		return
	}
	// retrying the msg in the queue after client is set up
	go p.retryMsg()
}

// ActiveTime. get the active time when peer receive msg
func (p *Peer) ActiveTime() time.Time {
	if p.client == nil {
		return p.closeTime
	}
	return p.client.Time
}

func (p *Peer) State() ConnectState {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.state
}

func (p *Peer) IsSendingMsg() bool {
	cnt := p.getMsgLen()
	return cnt > 0
}

func (p *Peer) SetState(s ConnectState) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.state = s
	if s == ConnectStateFailed {
		p.failedCount.Dial++
	}
}

func (p *Peer) GetFailedCnt() *FailedCount {
	return p.failedCount
}

func (p *Peer) AddReceivedMsg(msgId string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.receivedMsg.Add(msgId, "")
}

func (p *Peer) IsMsgReceived(msgId string) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	_, ok := p.receivedMsg.Get(msgId)
	return ok
}

// Send. send msg without wait it's reply
func (p *Peer) Send(msgId string, msg proto.Message) error {
	defer func() {
		if e := recover(); e != nil {
			log.Errorf("send panic recover err %v", e)
		}
	}()
	if p == nil {
		return fmt.Errorf("peer is nil")
	}
	if len(msgId) == 0 {
		msgId = utils.GenIdByTimestamp(rand.New(rand.NewSource(time.Now().UnixNano())))
	}
	if p.client == nil {
		return fmt.Errorf("client is nil")
	}
	ch := make(chan MsgReply, 1)
	log.Debugf("send msg %s to %s", msgId, p.addr)
	if err := p.addMsg(msgId, &MsgWrap{msg: msg, id: msgId, reply: ch}); err != nil {
		return err
	}
	go p.retryMsg()
	if err := p.client.AsyncSendAndWaitAck(context.Background(), msg, msgId); err != nil {
		p.failedCount.Send++
		return err
	}
	log.Debugf("wait for msg reply of %s", msgId)
	reply := <-ch
	log.Debugf("receive msg ack of %s from %s", msgId, p.addr)
	if reply.err != nil {
		p.failedCount.Send++
	}
	return reply.err
}

// SendAndWaitReply. send msg and wait the response reply msg
func (p *Peer) SendAndWaitReply(msgId string, msg proto.Message) (proto.Message, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Errorf("send and wait reply recover err %v", e)
		}
	}()
	if p == nil {
		return nil, fmt.Errorf("peer is nil")
	}
	if len(msgId) == 0 {
		msgId = utils.GenIdByTimestamp(rand.New(rand.NewSource(time.Now().UnixNano())))
	}
	if p.client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	ch := make(chan MsgReply, 1)
	log.Debugf("send msg %s and wait for reply to %s", msgId, p.addr)
	if err := p.addMsg(msgId, &MsgWrap{msg: msg, id: msgId, needReply: true, reply: ch}); err != nil {
		return nil, err
	}
	go p.retryMsg()
	if err := p.client.AsyncSendAndWaitAck(context.Background(), msg, msgId); err != nil {
		p.failedCount.Send++
		return nil, err
	}
	reply := <-ch
	log.Debugf("receive msg reply of %d from %s", msgId, p.addr)
	if reply.err != nil {
		p.failedCount.Send++
	}
	return reply.msg, reply.err
}

// ReceiveMsg. handle receive msg
func (p *Peer) Receive(origId string, replyMsg proto.Message) {
	if len(origId) == 0 {
		return
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	var e *list.Element
	var msgWrap *MsgWrap
	for e = p.mq.Front(); e != nil; e = e.Next() {
		msgWrap = e.Value.(*MsgWrap)
		if msgWrap.id == origId {
			break
		}
	}
	if e == nil {
		return
	}
	if !msgWrap.needReply {
		return
	}
	msgWrap.reply <- MsgReply{msg: replyMsg}
	p.mq.Remove(e)
	delete(p.retry, msgWrap.id)
	log.Debugf("after remove msg %s-%s len %d", origId, msgWrap.id, p.mq.Len())
}

func (p *Peer) retryMsg() {
	defer func() {
		if e := recover(); e != nil {
			log.Errorf("retry msg panic recover err %v", e)
		}
	}()
	if p.serviceRetrying() {
		return
	}
	p.setRetrying(true)
	ti := time.NewTicker(time.Duration(1) * time.Second)
	defer func() {
		ti.Stop()
		p.setRetrying(false)
	}()
	var closeSignal chan struct{}
	if p.client != nil {
		closeSignal = p.client.CloseSignal
	}
	for {
		select {
		case <-ti.C:
			if p.getMsgLen() == 0 {
				return
			}
			msgWrap := p.getMsgToRetry()
			if msgWrap == nil {
				continue
			}
			if p.client == nil {
				log.Debugf("retry msg %s, but client is disconnect", msgWrap.id)
				continue
			}
			log.Debugf("get msg to retry %s", msgWrap.id)
			if err := p.client.AsyncSendAndWaitAck(context.Background(), msgWrap.msg, msgWrap.id); err != nil {
				log.Errorf("send msg to %s err in retry service %s", p.addr, err)
			}
		case <-closeSignal:
			log.Debugf("exit retry msg service, because client %s is closed", p.addr)
			return
		}
	}
}

func (p *Peer) acceptAckNotify() {
	log.Debugf("acceptAckNotify %v", p.client.Address)
	for {
		select {
		case notify, ok := <-p.client.AckStatusNotify:
			log.Debugf("receive notify %v %v", notify, ok)
			if !ok {
				continue
			}
			if len(notify.MessageID) == 0 {
				log.Warnf("receive a empty msg id msg")
				continue
			}
			p.receiveMsgNotify(notify)
		case <-p.client.CloseSignal:
			log.Debugf("exit accept ack notify service, because client %s is closed", p.addr)
			p.resetAllMsgs()
			p.client = nil
			p.closeTime = time.Now()
			p.failedCount.Disconnect++
			return
		}
	}
}

// addMsg. add msg to list
func (p *Peer) addMsg(msgId string, msg *MsgWrap) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.retry[msgId]; ok {
		return fmt.Errorf("add a duplicated msg %s", msgId)
	}
	msg.createdAt = utils.GetMilliSecTimestamp()
	p.mq.PushBack(msg)
	log.Debugf("add msg %v, need reply %t, len %d", msgId, msg.needReply, p.mq.Len())
	p.retry[msgId] = 0
	return nil
}

func (p *Peer) getMsgLen() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.mq.Len()
}

func (p *Peer) getMsgToRetry() *MsgWrap {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if p.client == nil {
		log.Warnf("get msg to retry but client is nil %s", p.addr)
		return nil
	}
	for e := p.mq.Front(); e != nil; e = e.Next() {
		msgWrap := e.Value.(*MsgWrap)
		if msgWrap.state == MsgStateSended {
			log.Debugf("msg %s has sended, ignore retring", msgWrap.id)
			continue
		}
		// msg is sending
		if _, ok := p.client.SyncWaitAck.Load(msgWrap.id); ok {
			log.Debugf("msg %s already in component retry queue", msgWrap.id)
			continue
		}
		if utils.GetMilliSecTimestamp() < msgWrap.createdAt+(common.ACK_MSG_CHECK_INTERVAL*1000) {
			log.Debugf("msg %s just send, delay retry it after %ds", msgWrap.id, common.ACK_MSG_CHECK_INTERVAL)
			continue
		}
		// msg is waiting
		return msgWrap
	}
	return nil
}

// receive msg ack notify
func (p *Peer) receiveMsgNotify(notify network.AckStatus) {
	p.lock.Lock()
	defer p.lock.Unlock()
	var e *list.Element
	for e = p.mq.Front(); e != nil; e = e.Next() {
		msgWrap := e.Value.(*MsgWrap)
		if msgWrap.id == notify.MessageID {
			break
		}
	}
	if e == nil {
		log.Debugf("receive notify of msg id %s, but it's removed from the queue", notify.MessageID)
		return
	}
	msgWrap := e.Value.(*MsgWrap)
	log.Debugf("receive msg ack notify %s, result %d, need reply %t", msgWrap.id, notify.Status, msgWrap.needReply)
	// receive ack msg, remove it from list
	if notify.Status == network.ACK_SUCCESS {
		msgWrap.state = MsgStateSended
		if !msgWrap.needReply {
			msgWrap.reply <- MsgReply{}
			p.mq.Remove(e)
			delete(p.retry, msgWrap.id)
			log.Debugf("remove no-reply msg %s from queue, left %d msg", msgWrap.id, p.mq.Len())
		}
		return
	}
	if p.retry[notify.MessageID] >= common.MAX_MSG_RETRY {
		msgWrap.reply <- MsgReply{err: errors.New("retry too many")}
		p.mq.Remove(e)
		delete(p.retry, msgWrap.id)
		log.Debugf("remove %s msg which retry too much, left %d msg", msgWrap.id, p.mq.Len())
		p.failedCount.Recv++
		return
	}
	p.retry[msgWrap.id]++
}

// resetAllMsgs. reset all msg state to none
func (p *Peer) resetAllMsgs() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for e := p.mq.Front(); e != nil; e = e.Next() {
		msgWrap := e.Value.(*MsgWrap)
		if msgWrap == nil {
			continue
		}
		msgWrap.state = MsgStateNone
	}
}

func (p *Peer) serviceRetrying() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.retrying
}

func (p *Peer) setRetrying(r bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.retrying = r
}
