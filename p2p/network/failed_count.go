package network

import "github.com/saveio/themis/common/log"

type FailedCount struct {
	dial       int
	send       int
	recv       int
	disconnect int
}

func (this *Network) AddDialFailedCnt(addr string) {
	val, _ := this.peerFailedCount.LoadOrStore(addr, &FailedCount{})
	cnt, ok := val.(*FailedCount)
	if !ok {
		return
	}
	cnt.dial++
	this.peerFailedCount.Store(addr, val)
}

func (this *Network) AddSendFailedCnt(addr string) {
	val, _ := this.peerFailedCount.LoadOrStore(addr, &FailedCount{})
	cnt, ok := val.(*FailedCount)
	if !ok {
		return
	}
	cnt.send++
	this.peerFailedCount.Store(addr, val)
}

func (this *Network) AddRecvFailedCnt(addr string) {
	val, _ := this.peerFailedCount.LoadOrStore(addr, &FailedCount{})
	cnt, ok := val.(*FailedCount)
	if !ok {
		return
	}
	cnt.recv++
	this.peerFailedCount.Store(addr, val)
}

func (this *Network) AddDisconnectCnt(addr string) {
	val, _ := this.peerFailedCount.LoadOrStore(addr, &FailedCount{})
	cnt, ok := val.(*FailedCount)
	if !ok {
		return
	}
	cnt.disconnect++
	this.peerFailedCount.Store(addr, val)
}

func (this *Network) IsPeerNetQualityBad(addr string) bool {
	totalFailed := &FailedCount{}
	peerCnt := 0
	var peerFailed *FailedCount
	this.peerFailedCount.Range(func(key, value interface{}) bool {
		peerCnt++
		peerAddr, _ := key.(string)
		cnt, _ := value.(*FailedCount)
		totalFailed.dial += cnt.dial
		totalFailed.send += cnt.send
		totalFailed.recv += cnt.recv
		totalFailed.disconnect += cnt.disconnect
		if peerAddr == addr {
			peerFailed = cnt
		}
		return false
	})
	if peerFailed == nil {
		return false
	}
	log.Debugf("peer failed %v, totalFailed %v, peer cnt %v", peerFailed, totalFailed, peerCnt)
	if peerFailed.dial >= totalFailed.dial/peerCnt ||
		peerFailed.send >= totalFailed.send/peerCnt ||
		peerFailed.disconnect >= totalFailed.disconnect/peerCnt ||
		peerFailed.recv >= totalFailed.recv/peerCnt {
		return true
	}
	return false
}
