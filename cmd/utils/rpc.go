package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/themis-go-sdk/client"
	rpcerr "github.com/saveio/themis/http/base/error"

	"github.com/urfave/cli"
)

var (
	dspRestAddr string
	callDspRest bool
)

//JsonRpc version
const JSON_RPC_VERSION = "2.0"

const (
	ERROR_INVALID_PARAMS   = rpcerr.INVALID_PARAMS
	ERROR_ONTOLOGY_COMMON  = 10000
	ERROR_ONTOLOGY_SUCCESS = 0
)

type OntologyError struct {
	ErrorCode int64
	Error     error
}

func NewOntologyError(err error, errCode ...int64) *OntologyError {
	ontErr := &OntologyError{Error: err}
	if len(errCode) > 0 {
		ontErr.ErrorCode = errCode[0]
	} else {
		ontErr.ErrorCode = ERROR_ONTOLOGY_COMMON
	}
	if err == nil {
		ontErr.ErrorCode = ERROR_ONTOLOGY_SUCCESS
	}
	return ontErr
}

//JsonRpcRequest object in rpc
type JsonRpcRequest struct {
	Version string        `json:"jsonrpc"`
	Id      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

//JsonRpcResponse object response for JsonRpcRequest
type JsonRpcResponse struct {
	Error  int64           `json:"error"`
	Desc   string          `json:"desc"`
	Result json.RawMessage `json:"result"`
}

func sendRpcRequest(method string, params []interface{}) ([]byte, *OntologyError) {
	rpcReq := &JsonRpcRequest{
		Version: JSON_RPC_VERSION,
		Id:      "cli",
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, NewOntologyError(fmt.Errorf("JsonRpcRequest json.Marshal error:%s", err))
	}

	addr := fmt.Sprintf(config.Parameters.BaseConfig.ChainRestAddr)
	resp, err := http.Post(addr, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return nil, NewOntologyError(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, NewOntologyError(fmt.Errorf("read rpc response body error:%s", err))
	}
	rpcRsp := &JsonRpcResponse{}
	err = json.Unmarshal(body, rpcRsp)
	if err != nil {
		return nil, NewOntologyError(fmt.Errorf("json.Unmarshal JsonRpcResponse:%s error:%s", body, err))
	}
	if rpcRsp.Error != 0 {
		return nil, NewOntologyError(fmt.Errorf("%s", strings.ToLower(rpcRsp.Desc)), rpcRsp.Error)
	}
	return rpcRsp.Result, nil
}

//Api for cmd sending rest request to dsp daemon
func BeforeFunc(ctx *cli.Context) error {
	if ctx.IsSet(flags.GetFlagName(flags.DspRestAddrFlag)) {
		addr := ctx.String(flags.GetFlagName(flags.DspRestAddrFlag))
		SetAddress(addr)
		callDspRest = true
	}
	return nil
}

func CallDspRest() bool {
	return callDspRest
}

func SetAddress(addr string) {
	if addr == "" {
		addr = "127.0.0.1"
	}
	dspRestAddr = fmt.Sprintf("%s:%d", addr, config.Parameters.BaseConfig.PortBase+uint32(config.Parameters.BaseConfig.HttpRestPortOffset))
}

//[TODO] move below functions to themis-go-sdk
func GetRequestUrl(reqPath string, values ...*url.Values) (string, error) {
	addr := dspRestAddr

	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
	reqUrl, err := new(url.URL).Parse(addr)
	if err != nil {
		return "", fmt.Errorf("Parse address:%s error:%s", addr, err)
	}
	reqUrl.Path = reqPath
	if len(values) > 0 {
		first := values[0]
		if first != nil {
			reqUrl.RawQuery = first.Encode()
		}
	}
	return reqUrl.String(), nil
}

func DealRestResponse(body io.Reader) ([]byte, error) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read http body error:%s", err)
	}
	restRsp := &client.RestfulResp{}
	err = json.Unmarshal(data, restRsp)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal RestfulResp:%s error:%s", body, err)
	}
	if restRsp.Error != 0 {
		return nil, fmt.Errorf("sendRestRequest error code:%d desc:%s result:%s", restRsp.Error, restRsp.Desc, restRsp.Result)
	}
	return restRsp.Result, nil
}

func SendRestGetRequest(reqPath string, values ...*url.Values) ([]byte, error) {
	reqUrl, err := GetRequestUrl(reqPath, values...)
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("send http get request error:%s", err)
	}
	defer resp.Body.Close()
	return DealRestResponse(resp.Body)
}
