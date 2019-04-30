package http

import (
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/http/rest"
)

//start restful
func StartRestServer(dp *dsp.Endpoint) {

	rest.DspService = dp
	rt := rest.InitRestServer()
	go rt.Start()
}
