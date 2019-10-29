package websocket

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/saveio/themis/common/log"
)

func TestWsServer(t *testing.T) {
	StartServer()
	waitToExit()
}

func TestWsClient(t *testing.T) {
	log.InitLog(1, log.Stdout)
	var c *websocket.Conn
	var err error
	for {
		c, _, err = websocket.DefaultDialer.Dial("ws://localhost:10339", nil)
		if err != nil {
			fmt.Printf("dial %s\n", err)
			time.Sleep(time.Duration(2) * time.Second)
			continue
		}
		break
	}
	defer c.Close()
	m := map[string]interface{}{
		"Action": "subscribe",
	}
	data, _ := json.Marshal(m)
	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		t.Fatal(err)
	}
	for {
		msgType, message, err := c.ReadMessage()
		if err != nil {
			t.Fatal("ReadMessage:", err)
		}
		log.Infof("type: %d, recv: %s\n", msgType, message)
	}
}
func waitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("do received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}
