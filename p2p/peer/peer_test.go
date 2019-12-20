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
