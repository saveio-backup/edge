package network

import (
	"fmt"
	"sync"
	"testing"
)

func TestFailedCount(t *testing.T) {
	n := Network{}
	n.peerFailedCount = new(sync.Map)
	n.PeerDialFailed("1")
	val, ok := n.peerFailedCount.Load("1")
	if !ok {
		t.Fatal("convert failed")
	}
	cnt, ok := val.(*FailedCount)
	if !ok {
		t.Fatal("convert failed")
	}
	fmt.Println(cnt.dial)
	if cnt.dial != 1 {
		t.Fatal("no match")
	}
}
