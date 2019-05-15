package dsp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/saveio/carrier/network"
	"github.com/saveio/dsp-go-sdk/network/common"
	"github.com/saveio/dsp-go-sdk/network/message"
	"github.com/saveio/themis/common/log"
)

var node1ListAddr = "udp://127.0.0.1:50001"
var node2ListAddr = "udp://127.0.0.1:50002"

func TestNetworkReceiveMsg(t *testing.T) {
	n := NewNetwork()
	n.Start(node1ListAddr)
	counter := 0
	n.handler = func(ctx *network.ComponentContext) {
		msg := message.ReadMessage(ctx.Message())
		if msg == nil {
			return
		}
		fmt.Printf("receive msg:%v, from address %s， id:%s, pubkey:%s\n", msg, ctx.Client().Address, ctx.Client().ID.String(), ctx.Client().ID.PublicKeyHex())
		counter++
		if counter == 1 {
			return
		}

		fmt.Printf("reply\n")
		err := ctx.Reply(context.Background(), msg.ToProtoMsg())
		if err != nil {
			fmt.Printf("reply err:%v\n", err)
		}
	}
	tick := time.NewTicker(time.Second)
	for {
		<-tick.C
	}
}

func TestNetworkSendMsg(t *testing.T) {
	n := NewNetwork()
	n.Start(node2ListAddr)
	n.Connect(node1ListAddr)
	msg := &message.Message{}
	msg.Header = &message.Header{
		Version:   "0",
		Type:      common.MSG_TYPE_BLOCK,
		MsgLength: 0,
	}
}

func TestDialIP(t *testing.T) {
	log.InitLog(1)
	n := NewNetwork()
	n.Start(node2ListAddr)
	addr := "udp://127.0.0.1:13004"
	err := n.Connect(addr)
	if err != nil {
		t.Fatal(err)
	}
	listen := n.IsPeerListenning(addr)
	fmt.Printf("connected %v\n", listen)

	time.Sleep(time.Duration(3) * time.Second)
	err = n.Disconnect(addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("disconnected")
	time.Sleep(time.Duration(3) * time.Second)
}
