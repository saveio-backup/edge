package network

type FailedCount struct {
	dial       int
	send       int
	recv       int
	disconnect int
}

func NewFailedCount() *FailedCount {
	return &FailedCount{}
}

func (this *FailedCount) DialFailed() {
	this.dial++
}

func (this *FailedCount) SendFailed() {
	this.dial++
}

func (this *FailedCount) RecvFailed() {
	this.recv++
}

func (this *FailedCount) Disconnect() {
	this.disconnect++
}

func (this *Network) PeerDialFailed(addr string) {
	// cnt, _ := this.peerFailedCount.LoadOrStore(addr, &FailedCount{})
}
