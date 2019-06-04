package localrpc

import (
	"fmt"
	"net/http"

	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/http/base/rpc"
	"github.com/saveio/themis/common/log"
)

const (
	LOCAL_HOST string = "127.0.0.1"
	LOCAL_DIR  string = "/local"
)

func StartLocalRpcServer() error {
	log.Debug()
	http.HandleFunc(LOCAL_DIR, rpc.Handle)

	rpc.HandleFunc("setdebuginfo", rpc.SetDebugInfo)

	localRpc := fmt.Sprintf("127.0.0.1:%d", config.Parameters.BaseConfig.PortBase+uint32(config.Parameters.BaseConfig.LocalRpcPortOffset))
	err := http.ListenAndServe(localRpc, nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}
