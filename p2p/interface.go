package p2p

import p2pHttp "github.com/saveio/edge/p2p/http"

//start p2p http
func StartP2PHttp() {
	hs := p2pHttp.InitHttpServer()
	go hs.Start()
}
