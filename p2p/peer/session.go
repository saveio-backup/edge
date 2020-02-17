package peer

import (
	"fmt"
	"sync"
	"time"

	"github.com/saveio/carrier/network"
	"github.com/saveio/edge/common"
	"github.com/saveio/themis/common/log"
)

type Session struct {
	id       string              // session id
	streamId string              // stream id
	client   *network.PeerClient // network overlay peer client instance
	txSpeeds []uint64            // send msg speed record
	rxSpeeds []uint64            // receive msg speed record
	lock     *sync.RWMutex       // stream lock
	closeCh  chan struct{}       // close signal
}

func NewSession(sessionId string) *Session {
	s := &Session{
		id:       sessionId,
		lock:     new(sync.RWMutex),
		txSpeeds: make([]uint64, 0, common.MAX_SESSION_RECORD_SPEED_LEN),
		rxSpeeds: make([]uint64, 0, common.MAX_SESSION_RECORD_SPEED_LEN),
	}
	return s
}

func (this *Session) SetClient(client *network.PeerClient) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.client = client
}

func (this *Session) Open() error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if len(this.streamId) != 0 {
		return nil
	}
	if this.client == nil {
		return fmt.Errorf("session %s, client is nil", this.id)
	}
	s := this.client.OpenStream()
	if s == nil {
		return fmt.Errorf("session %s, stream is nil", this.id)
	}
	this.streamId = s.ID
	this.closeCh = make(chan struct{})
	go this.startRecordSpeed()
	return nil
}

func (this *Session) Close() error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if len(this.streamId) == 0 {
		return nil
	}
	if this.client == nil {
		return fmt.Errorf("session %s, client is nil", this.id)
	}
	this.client.CloseStream(this.streamId)
	close(this.closeCh)
	return nil
}

func (this *Session) GetStreamId() string {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.streamId
}

func (this *Session) GetTxAvgSpeed() uint64 {
	sum := uint64(0)
	for _, r := range this.txSpeeds {
		sum += r
	}
	return sum / uint64(len(this.txSpeeds))
}

func (this *Session) GetRxAvgSpeed() uint64 {
	sum := uint64(0)
	for _, r := range this.rxSpeeds {
		sum += r
	}
	return sum / uint64(len(this.rxSpeeds))
}

func (this *Session) startRecordSpeed() {
	ti := time.NewTicker(time.Duration(common.SESSION_SPEED_RECORD_INTERVAL) * time.Second)
	defer ti.Stop()
	// TODO: pause ticker if peer hasn't send any data for a long time?
	for {
		select {
		case <-ti.C:
			streamId := this.GetStreamId()
			if len(streamId) == 0 {
				continue
			}
			if this.client == nil {
				log.Warnf("peer %s stream %s has no client", this.id, streamId)
				continue
			}
			data := this.client.StreamSendDataCnt(streamId)
			speed := data / common.SESSION_SPEED_RECORD_INTERVAL
			log.Debugf("record session tx speed %v", speed)
			this.txSpeeds = append(this.txSpeeds, speed)
			if len(this.txSpeeds) < common.MAX_SESSION_RECORD_SPEED_LEN {
				continue
			}
			this.txSpeeds = this.txSpeeds[1 : 1+common.MAX_SESSION_RECORD_SPEED_LEN]
		case <-this.closeCh:
			log.Debugf("stop record speed when stream is closed")
			return
		}
	}
}
