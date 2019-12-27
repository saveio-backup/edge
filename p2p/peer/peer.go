package peer

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/saveio/carrier/network"
	"github.com/saveio/dsp-go-sdk/utils"
	"github.com/saveio/edge/common"
	"github.com/saveio/themis/common/log"
)

type MsgReply struct {
	id  string
	msg proto.Message
	err error
}

type MsgWrap struct {
	id        string
	msg       proto.Message
	sending   bool
	sended    bool
	needReply bool
	reply     chan MsgReply
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
}

func New(addr string) *Peer {
	p := &Peer{
		addr:        addr,
		mq:          list.New(),
		lock:        new(sync.RWMutex),
		failedCount: new(FailedCount),
		retry:       make(map[string]int),
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

// Send. send msg without wait it's reply
func (p *Peer) Send(msgId string, msg proto.Message) error {
	if p == nil {
		return fmt.Errorf("peer is nil")
	}
	if len(msgId) == 0 {
		msgId = utils.GenIdByTimestamp()
	}
	if p.client == nil {
		return fmt.Errorf("client is nil")
	}
	ch := make(chan MsgReply, 1)
	log.Debugf("send msg %s to %s", msgId, p.addr)
	p.addMsg(msgId, &MsgWrap{msg: msg, id: msgId, sending: true, reply: ch})
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
	if p == nil {
		return nil, fmt.Errorf("peer is nil")
	}
	if len(msgId) == 0 {
		msgId = utils.GenIdByTimestamp()
	}
	if p.client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	ch := make(chan MsgReply, 1)
	log.Debugf("send msg %s and wait for reply to %s", msgId, p.addr)
	p.addMsg(msgId, &MsgWrap{msg: msg, id: msgId, sending: true, needReply: true, reply: ch})
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
	if p.serviceRetrying() {
		return
	}
	p.setRetrying(true)
	ti := time.NewTicker(time.Duration(1) * time.Second)
	defer func() {
		ti.Stop()
		p.setRetrying(false)
	}()
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
				log.Errorf("send msg to %s err in retry service %s", p.client, err)
			}
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
			p.client = nil
			p.closeTime = time.Now()
			p.failedCount.Disconnect++
			return
		}
	}
}

// addMsg. add msg to list
func (p *Peer) addMsg(msgId string, msg *MsgWrap) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.retry[msgId]; ok {
		log.Warnf("msg exit %s", msgId)
		return
	}
	p.mq.PushBack(msg)
	log.Debugf("add msg %v, need reply %t, len %d", msgId, msg.needReply, p.mq.Len())
	p.retry[msgId] = 0
}

func (p *Peer) getMsgLen() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.mq.Len()
}

func (p *Peer) getMsgToRetry() *MsgWrap {
	p.lock.RLock()
	defer p.lock.RUnlock()
	for e := p.mq.Front(); e != nil; e = e.Next() {
		msgWrap := e.Value.(*MsgWrap)
		if msgWrap.sended {
			continue
		}
		// msg haven't sended
		if msgWrap.sending {
			return nil
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
	log.Debugf("receive msg ack notify %d, result %d, need reply %t", msgWrap.id, notify.Status, msgWrap.needReply)
	// receive ack msg, remove it from list
	if notify.Status == network.ACK_SUCCESS {
		msgWrap.sended = true
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
	msgWrap.sending = false
	p.retry[msgWrap.id]++
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
