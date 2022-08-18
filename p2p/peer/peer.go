package peer

import (
	"container/list"
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/saveio/carrier/network"
	dspMsg "github.com/saveio/dsp-go-sdk/network/message"
	uTime "github.com/saveio/dsp-go-sdk/utils/time"
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
	id           string        // msg id
	sessionId    string        // msg to session id
	msg          proto.Message // msg
	needReply    bool          // if msg need reply
	reply        chan MsgReply // msg reply channel
	createdAt    uint64        // msg created time
	writeTimeout time.Duration // timeout for msg to sent
	timeout      uint64        // msg send timeout
	retry        uint32        // msg retry times
	retryAt      uint64        // retry msg time stamp

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
	id           string              // peer id
	addr         string              // peer host address
	client       *network.PeerClient // network overlay peer client instance
	mq           *list.List          // message queue list
	lock         *sync.RWMutex       // lock for sync operation
	retry        map[string]int      // count recorder for retry send msg
	retrying     bool                // flag of service is retrying
	closeTime    time.Time           // peer closed time
	failedCount  *FailedCount        // peer QoS failed count
	state        ConnectState        // connect state
	receivedMsg  *lru.ARCCache       // received msg cache
	sessions     map[string]*Session // session map, session id <=> session struct
	waiting      bool                // waiting peer reconnect
	isBadQuality bool                // is bad quality once
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
		sessions:    make(map[string]*Session),
	}
	return p
}

// GetHostAddr.
func (p *Peer) GetHostAddr() string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.addr
}

func (p *Peer) SetPeerId(peerId string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.id = peerId
}

// GetPeerId.
func (p *Peer) GetPeerId() string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.id
}

func (p *Peer) ClientIsEmpty() bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.client == nil
}

func (p *Peer) IsBadQuality() bool {
	return p.isBadQuality
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
	log.Debugf("set peer client %p for %s", client, p.addr)
	go p.acceptAckNotify()
	if p.mq.Len() == 0 {
		return
	}
	// retrying the msg in the queue after client is set up
	go p.retryMsg()
}

func (p *Peer) GetClient() *network.PeerClient {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.client
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
		msgId = dspMsg.GenIdByTimestamp(rand.New(rand.NewSource(time.Now().UnixNano())))
	}
	ch := make(chan MsgReply, 1)
	log.Debugf("send msg %s to %s", msgId, p.addr)
	if err := p.addMsg(msgId, &MsgWrap{msg: msg, id: msgId, reply: ch}); err != nil {
		return err
	}
	go p.retryMsg()
	if p.client != nil {
		if err := p.client.AsyncSendAndWaitAck(context.Background(), msg, msgId); err != nil {
			p.failedCount.Send++
			log.Errorf("async send msg %s err %s", msgId, err)
		}
	} else {
		log.Warnf("msg %s can't be sent, because %s is closed, wait for retry", msgId, p.id)
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
		msgId = dspMsg.GenIdByTimestamp(rand.New(rand.NewSource(time.Now().UnixNano())))
	}
	ch := make(chan MsgReply, 1)
	log.Debugf("send msg %s and wait for reply to %s", msgId, p.addr)
	if err := p.addMsg(msgId, &MsgWrap{msg: msg, id: msgId, needReply: true, reply: ch}); err != nil {
		return nil, err
	}
	go p.retryMsg()
	if p.client != nil {
		if err := p.client.AsyncSendAndWaitAck(context.Background(), msg, msgId); err != nil {
			p.failedCount.Send++
			log.Errorf("async send msg %s err %s", msgId, err)
		}
	} else {
		log.Warnf("msg %s can't be sent, because %s is closed, wait for retry", msgId, p.id)
	}

	reply := <-ch
	log.Debugf("receive msg reply of %s from %s", msgId, p.addr)
	if reply.err != nil {
		p.failedCount.Send++
	}
	return reply.msg, reply.err
}

// Send. send msg without wait it's reply
func (p *Peer) StreamSend(sessionId, msgId string, msg proto.Message, sendTimeout time.Duration) error {
	defer func() {
		if e := recover(); e != nil {
			log.Errorf("send panic recover err %v", e)
		}
	}()
	if p == nil {
		return fmt.Errorf("peer is nil")
	}
	if len(msgId) == 0 {
		msgId = dspMsg.GenIdByTimestamp(rand.New(rand.NewSource(time.Now().UnixNano())))
	}
	if len(sessionId) == 0 {
		sessionId = dspMsg.GenIdByTimestamp(rand.New(rand.NewSource(time.Now().UnixNano())))
	}
	ch := make(chan MsgReply, 1)
	streamMsg := &MsgWrap{
		msg:          msg,
		id:           msgId,
		sessionId:    sessionId,
		reply:        ch,
		writeTimeout: sendTimeout,
		timeout:      calcTimeout(0, sendTimeout),
	}
	if err := p.addMsg(msgId, streamMsg); err != nil {
		return err
	}
	go p.retryMsg()
	log.Debugf("send msg %s to %s, sessionId %s sendTimeout %d", msgId, p.addr, sessionId, sendTimeout)
	if _, err := p.streamAsyncSendAndWaitAck(msg, sessionId, msgId); err != nil {
		log.Errorf("stream send msg %s err %s", msgId, err)
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
func (p *Peer) StreamSendAndWaitReply(sessionId, msgId string, msg proto.Message, sendTimeout time.Duration) (
	proto.Message, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Errorf("send and wait reply recover err %v", e)
		}
	}()
	if p == nil {
		return nil, fmt.Errorf("peer is nil")
	}
	if len(msgId) == 0 {
		msgId = dspMsg.GenIdByTimestamp(rand.New(rand.NewSource(time.Now().UnixNano())))
	}
	if len(sessionId) == 0 {
		sessionId = dspMsg.GenIdByTimestamp(rand.New(rand.NewSource(time.Now().UnixNano())))
	}
	ch := make(chan MsgReply, 1)
	log.Debugf("send msg %s and wait for reply to %s", msgId, p.addr)
	msgWrap := &MsgWrap{
		msg:          msg,
		id:           msgId,
		sessionId:    sessionId,
		needReply:    true,
		reply:        ch,
		writeTimeout: sendTimeout,
		timeout:      calcTimeout(0, sendTimeout),
	}
	if err := p.addMsg(msgId, msgWrap); err != nil {
		return nil, err
	}
	go p.retryMsg()
	if _, err := p.streamAsyncSendAndWaitAck(msg, sessionId, msgId); err != nil {
		log.Errorf("stream send msg %s err %s", msgId, err)
	}
	reply := <-ch
	log.Debugf("stream receive msg reply of %s from %s", msgId, p.addr)
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
	p.replyMsgs([]*list.Element{e}, MsgReply{msg: replyMsg})
	log.Debugf("after remove msg %s-%s len %d", origId, msgWrap.id, p.mq.Len())
}

// CloseSession. close session and its stream
func (p *Peer) CloseSession(sessionId string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.closeSession(sessionId)
}

func (p *Peer) closeSession(sessionId string) error {
	s, ok := p.sessions[sessionId]
	if !ok {
		return fmt.Errorf("session %s not exist", sessionId)
	}
	if err := s.Close(); err != nil {
		log.Errorf("close session %s failed", sessionId)
		return err
	}
	delete(p.sessions, sessionId)
	return nil
}

// GetSessionTxSpeed. get session send data speed
func (p *Peer) GetSessionSpeed(sessionId string) (uint64, uint64, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	ses, ok := p.sessions[sessionId]
	if !ok {
		return 0, 0, fmt.Errorf("session %s not found", sessionId)
	}
	return ses.GetTxAvgSpeed(), ses.GetRxAvgSpeed(), nil
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
			msgWrap, fails := p.getMsgToRetry()
			if fails != nil {
				p.lock.Lock()
				p.isBadQuality = true
				p.replyMsgs(fails, MsgReply{err: fmt.Errorf("bad network quality with %s in retryMsg", p.addr)})
				p.lock.Unlock()
			}
			if msgWrap == nil {
				continue
			}
			if p.client == nil {
				log.Debugf("retry msg %s, but client is disconnect", msgWrap.id)
				continue
			}
			log.Debugf("get msg to retry %s timeout %d", msgWrap.id, msgWrap.writeTimeout/time.Second)
			if msgWrap.writeTimeout > 0 {
				if _, err := p.streamAsyncSendAndWaitAck(msgWrap.msg, msgWrap.sessionId, msgWrap.id); err != nil {
					log.Errorf("send msg to %s err in retry service %s", p.addr, err)
				}
				continue
			}
			if err := p.client.AsyncSendAndWaitAck(context.Background(), msgWrap.msg, msgWrap.id); err != nil {
				log.Errorf("send msg to %s err in retry service %s", p.addr, err)
			}
		case <-closeSignal:
			log.Debugf("exit retry msg service, because client %s is closed", p.addr)
			go p.lostConn()
			return
		}
	}
}

func (p *Peer) acceptAckNotify() {
	log.Debugf("acceptAckNotify %v", p.client.Address)
	var closeSignal chan struct{}
	if p.client != nil {
		closeSignal = p.client.CloseSignal
	}
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
		case <-closeSignal:
			log.Debugf("exit accept ack notify service, because client %s is closed", p.addr)
			p.client = nil
			if err := p.closeAllSession(); err != nil {
				log.Errorf("close all session failed %s", err)
			}
			p.closeTime = time.Now()
			p.failedCount.Disconnect++
			return
		}
	}
}

// lostConn. connection is lost, wait for 1 minute in reconnecting. if failed. remove all pending msgs
func (p *Peer) lostConn() {
	if p.getMsgLen() == 0 {
		return
	}
	if p.IsWaitingReconnect() {
		return
	}
	p.SetPeerWaiting(true)
	defer p.SetPeerWaiting(false)
	ti := time.NewTicker(time.Second)
	defer ti.Stop()
	timeout := 0
	for {
		<-ti.C
		timeout++
		if timeout > common.MAX_PEER_RECONNECT_TIMEOUT {
			break
		}
		// check client is connected or not
		if p.ClientIsEmpty() {
			continue
		}
		break
	}
	if timeout < common.MAX_PEER_RECONNECT_TIMEOUT {
		return
	}
	// connection is lost
	p.lock.Lock()
	defer p.lock.Unlock()
	msgs := make([]*list.Element, 0, p.mq.Len())
	for e := p.mq.Front(); e != nil; e = e.Next() {
		msgWrap, ok := e.Value.(*MsgWrap)
		if ok {
			log.Debugf("peer %s is lost, remove msg %s", p.id, msgWrap.id)
		}
		msgs = append(msgs, e)
	}
	p.isBadQuality = true
	p.replyMsgs(msgs, MsgReply{err: fmt.Errorf("bad network quality with %s in lostConn", p.addr)})
}

func (p *Peer) SetPeerWaiting(w bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.waiting = w
}

func (p *Peer) IsWaitingReconnect() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.waiting
}

// addMsg. add msg to list
func (p *Peer) addMsg(msgId string, msg *MsgWrap) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.retry[msgId]; ok {
		return fmt.Errorf("add a duplicated msg %s", msgId)
	}
	msg.createdAt = uTime.GetMilliSecTimestamp()
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

func (p *Peer) getMsgToRetry() (*MsgWrap, []*list.Element) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if p.client == nil {
		log.Warnf("get msg to retry but client is nil %s", p.addr)
		return nil, nil
	}
	var retryMsgWrap *MsgWrap
	fails := make([]*list.Element, 0)
	for e := p.mq.Front(); e != nil; e = e.Next() {
		msgWrap := e.Value.(*MsgWrap)
		if msgWrap.retry > common.MAX_MSG_RETRY {
			log.Debugf("msg %s retry too many, append to remove list", msgWrap.id)
			fails = append(fails, e)
			continue
		}
		nowTimestamp := uTime.GetMilliSecTimestamp()
		// just created, delay it
		if nowTimestamp < msgWrap.createdAt+(common.ACK_MSG_CHECK_INTERVAL*1000) {
			log.Debugf("msg %s just send, delay retry it", msgWrap.id)
			continue
		}
		if msgWrap.writeTimeout > 0 && msgWrap.timeout <= nowTimestamp {
			log.Errorf("sessionId %s send msg %s timeout++++ %d", msgWrap.sessionId, msgWrap.id, msgWrap.timeout)
			p.closeSession(msgWrap.sessionId)
			log.Errorf("sessionId %s send msg %s timeout++++ close session success %d", msgWrap.sessionId, msgWrap.id, msgWrap.timeout)
			msgWrap.timeout = calcTimeout(msgWrap.retry, msgWrap.writeTimeout)
		}
		// in retry service
		retrying := msgWrap.retryAt > 0 && (nowTimestamp < msgWrap.retryAt+(common.MAX_ACK_MSG_TIMEOUT*1000))
		// msg is sending
		if _, ok := p.client.SyncWaitAck.Load(msgWrap.id); ok && (msgWrap.retryAt == 0 || retrying) {
			log.Debugf("msg %s already in component retry queue", msgWrap.id)
			if msgWrap.retryAt == 0 {
				msgWrap.retryAt = nowTimestamp
			}
			continue
		}
		if msgWrap.retryAt == 0 {
			msgWrap.retryAt = msgWrap.createdAt
			log.Debugf("msg %s has send success in 20s, wait for next round", msgWrap.id)
			continue
		}
		if retrying {
			log.Debugf("msg %s is in sync wait retry service", msgWrap.id)
			continue
		}
		log.Debugf("msg id %s retry %d times, last retry at %d", msgWrap.id, msgWrap.retry, msgWrap.retryAt)
		msgWrap.retryAt = nowTimestamp
		msgWrap.retry++
		// msg is waiting
		retryMsgWrap = msgWrap
		break
	}
	return retryMsgWrap, fails
}

func (p *Peer) replyMsgs(msgs []*list.Element, reply MsgReply) {
	for _, e := range msgs {
		msg, ok := e.Value.(*MsgWrap)
		if ok {
			msg.reply <- reply
		}
		p.mq.Remove(e)
		delete(p.retry, msg.id)
	}
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
		// msgWrap.state = MsgStateSended
		if !msgWrap.needReply {
			p.replyMsgs([]*list.Element{e}, MsgReply{})
			log.Debugf("remove no-reply msg %s from queue, left %d msg", msgWrap.id, p.mq.Len())
		}
		return
	}
	if p.retry[notify.MessageID] >= common.MAX_MSG_RETRY_FAILED {
		p.replyMsgs([]*list.Element{e}, MsgReply{err: fmt.Errorf("bad network quality with %s in receiveMsgNotify", p.addr)})
		log.Debugf("remove %s msg which retry too much, left %d msg", msgWrap.id, p.mq.Len())
		p.failedCount.Recv++
		p.isBadQuality = true
		return
	}
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

// openSession. open a session for send msgs on a stream.
// return stream id or error
func (p *Peer) openSession(sessionId string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.sessions[sessionId]; ok {
		return nil
	}
	ses := NewSession(sessionId)
	ses.SetClient(p.client)
	if err := ses.Open(); err != nil {
		log.Errorf("open session %s failed, err %s", sessionId, err)
		return err
	}
	p.sessions[sessionId] = ses
	return nil
}

// closeSession. close session and its stream
func (p *Peer) closeAllSession() error {
	p.lock.Lock()
	defer p.lock.Unlock()
	for id, s := range p.sessions {
		if err := s.Close(); err != nil {
			log.Debugf("close session %s failed", id)
			return err
		}
		delete(p.sessions, id)
	}
	return nil
}

func (p *Peer) streamAsyncSendAndWaitAck(msg proto.Message, sessionId, msgId string) (int32, error) {
	if err := p.openSession(sessionId); err != nil {
		return 0, err
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	ses, ok := p.sessions[sessionId]
	if !ok {
		log.Errorf("get session %s to send msg %s failed, no session", sessionId, msgId)
		return 0, fmt.Errorf("get session %s to send msg %s failed, no session", sessionId, msgId)
	}
	streamId := ses.GetStreamId()
	if p == nil {
		return 0, fmt.Errorf("peer is nil")
	}
	if len(streamId) == 0 {
		log.Debugf("stream id is empty when send msg %s", msgId)
	}
	var wrote int32
	var err error
	log.Debugf("stream send mgs %s", msgId)
	if p.client != nil {
		if err, wrote = p.client.StreamAsyncSendAndWaitAck(streamId, context.Background(), msg, msgId); err != nil {
			p.failedCount.Send++
			log.Errorf("stream %s send msg %s failed %s", streamId, msgId, err)
		}
	} else {
		log.Warnf("msg %s can't be sent, because %s is closed, wait for retry", msgId, p.id)
	}

	return wrote, err
}

func calcTimeout(ratio uint32, timeout time.Duration) uint64 {
	result := float64(1+float64(ratio)*0.2) * float64(timeout)
	return uTime.GetMilliSecTimestamp() + uint64(time.Duration(result)/time.Second)*1000
}
