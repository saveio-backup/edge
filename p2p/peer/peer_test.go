package peer

import (
	"container/list"
	"fmt"
	"testing"
	"time"
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

func TestCalcTimeout(t *testing.T) {
	a := time.Duration(128) * time.Second
	fmt.Printf("a= %d\n", a)
	for i := 1; i < 10; i++ {
		b := calcTimeout(i, a)
		fmt.Printf("b= %v\n", b/time.Second)
	}
}
