package utils

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/saveio/edge/common/config"
	"github.com/saveio/themis/common/log"
)

func UseHttpProfile() {
	r := http.NewServeMux()

	// Register pprof handlers
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	port := fmt.Sprintf(":%d", int(config.Parameters.BaseConfig.PortBase)+config.Parameters.BaseConfig.ProfilePortOffset)
	log.Infof("start profile at %v", port)
	http.ListenAndServe(port, r)

}
