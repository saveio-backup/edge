package peer

import (
	"container/list"
	"fmt"
	"testing"
)

func TestMQList(t *testing.T) {
	p := &Peer{}
	p.mq = list.New()
	v := p.mq.PushBack("123")
	fmt.Printf("val: %v, len: %d\n", v.Value, p.mq.Len())
	item := p.mq.Front()
	fmt.Printf("val: %v\n", item.Value)
	ret := p.mq.Remove(item)
	fmt.Printf("ret: %v len: %d\n", ret, p.mq.Len())
}

func TestNewSession(t *testing.T) {
	s := NewSession("123")
	for i := uint64(0); i < 20; i++ {
		s.txSpeeds = append(s.txSpeeds, i)
	}
	fmt.Printf("s.tx %v %d\n", s.txSpeeds, len(s.txSpeeds))
	fmt.Printf("s.tx %v %d\n", s.rxSpeeds, len(s.rxSpeeds))
}
