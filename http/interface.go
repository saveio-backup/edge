package http

import (
	"github.com/saveio/edge/http/rest"
)

//start restful
func StartRestServer() {
	rt := rest.InitRestServer()
	go rt.Start()
}
