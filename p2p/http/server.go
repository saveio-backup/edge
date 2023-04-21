package p2pHttp

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"github.com/saveio/dsp-go-sdk/types/state"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/p2p/http/transport"
	"github.com/saveio/themis/common/log"
)

type UploadRespanseData struct {
	Hashs    []string `json:"hashs"`
	FileHash string   `json:"fileHash"`
	TxHash   string   `json:"txHash"`
	Indexs   []uint64 `json:"indexs"`
	PeerAddr string   `json:"peerAddr"`
	Blocks   []string `json:"blocks"`
	Tags     []string `json:"tags"`
	// Offset    int64
	// PaymentId int32
}
type DownloadRespanseData struct {
	Hashs          []string `json:"hashs"` //block hash
	FileHash       string   `json:"fileHash"`
	TxHash         string   `json:"txHash"` //block txhash
	Indexs         []uint64 `json:"indexs"`
	PeerAddr       string   `json:"peerAddr"`
	PaymentId      int32    `json:"paymentId"`
	DownloadTaskId int32    `json:"downloadTaskId"`
}

type JsonResult struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type UploadResultData struct {
	HaveBeen  interface{} `json:"haveBeen"`
	PutSucess interface{} `json:"putSuccess"`
	PutError  interface{} `json:"putError"`
}
type PaymentRespanseData struct {
	Hashs []string `json:"hashs"`
}

type DownloadResultData struct {
	Hashs  []string `json:"hashs"`
	Blocks []string `json:"blocks"`
	Tags   []string `json:"Tags"`
}
type DeleteRespanseData struct {
	TaskTypes      []string `json:"taskTypes"`
	TransportState string   `json:"transportState"`
	PeerAddr       string   `json:"peerAddr"`
	TaskIds        []string `json:"taskIds"`
}

func NodeState(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	moduleState := dsp.DspService.GetNodeState()
	w.Header().Set("content-type", "application/json;charset=utf-8")
	m := make(map[string]state.ModuleState)
	m["state"] = moduleState
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)
}

func CreateUploadTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	peerAddr := r.URL.Query().Get("peerAddr")
	fileHash := r.URL.Query().Get("fileHash")
	prefix := r.URL.Query().Get("prefix")
	w.Header().Set("content-type", "application/json;charset=utf-8")
	if peerAddr == "" || fileHash == "" || prefix == "" {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	task := transport.NewTask()
	err := task.CreateUploadTask(peerAddr, fileHash, prefix)

	if err != nil {
		m := make(map[string]string)
		m["error"] = "create upload task error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	m := make(map[string]int32)
	m["createState"] = 1
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)
}

func NodeVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	query := r.URL.Query()
	fileHash := query.Get("fileHash")
	peerAddr := query.Get("peerAddr")
	blockHeight, _ := strconv.ParseUint(query.Get("blockHeight"), 10, 64)
	w.Header().Set("content-type", "application/json;charset=utf-8")
	if fileHash == "" || peerAddr == "" || blockHeight <= 0 {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	m := make(map[string]string)

	err := dsp.DspService.GetChain().WaitForTxConfirmed(blockHeight)
	if err != nil {
		log.Errorf("get block height err %s", err)
		m["verification"] = "0"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	info, err := dsp.DspService.GetChain().GetFileInfo(fileHash)
	if info == nil {
		log.Errorf("get file info err : %s", err)
		m["verification"] = "0"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	if info.FileOwner.ToBase58() != peerAddr && len(info.PrimaryNodes.AddrList) > 0 {
		log.Errorf("verification wrong peerAddr: %s,", peerAddr)
		m["verification"] = "0"
		m["error"] = "verification wrong peerAddr: " + peerAddr
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	m["verification"] = "1"
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)
}

func FileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {

		log.Error(err)
	}
	var resp UploadRespanseData
	json.Unmarshal([]byte(body), &resp)

	w.Header().Set("content-type", "application/json;charset=utf-8")
	if resp.FileHash == "" || resp.PeerAddr == "" || len(resp.Indexs) <= 0 || len(resp.Tags) <= 0 || len(resp.Blocks) <= 0 || len(resp.Hashs) <= 0 || resp.TxHash == "" {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	//result data
	uploadResultData := UploadResultData{}

	if len(resp.Hashs) < 1 {
		res, _ := json.Marshal(JsonResult{
			Code: 200,
			Msg:  "success",
			Data: uploadResultData,
		})
		w.Write(res)
		return
	}
	notContainIndexs, containIndexs := transport.FilterUploadHashs(resp.PeerAddr, resp.FileHash, resp.Hashs)
	if len(containIndexs) > 0 {
		containMap := make(map[string]interface{})
		var hashs []string
		var indexs []uint64
		for _, i := range containIndexs {
			hashs = append(hashs, resp.Hashs[i])
			indexs = append(indexs, resp.Indexs[i])
		}
		containMap["hashs"] = hashs
		containMap["indexs"] = indexs
		uploadResultData.HaveBeen = containMap
	}
	if len(notContainIndexs) < 1 {
		res, _ := json.Marshal(JsonResult{
			Code: 200,
			Msg:  "success",
			Data: uploadResultData,
		})
		w.Write(res)
		return
	}
	fs := dsp.DspService.GetFS()
	blkInfos := make([]*transport.UploadFileDetail, 0)

	putSuccessMap := make(map[string]interface{})
	var unknowStateHashs []string
	var unknowStateIndexs []uint64
	var unknowStateErrors []string
	putErrorMap := make(map[string]interface{})
	var errorHashs []string
	var errorIndexs []uint64
	var errorInfos []string
	for _, i := range notContainIndexs {
		block, _ := hex.DecodeString(resp.Blocks[i])
		blk := fs.EncodedToBlockWithCid(block, resp.Hashs[i])
		if blk.Cid().String() != resp.Hashs[i] {
			log.Errorf("receive a wrong block: %s, expected: %s", blk.Cid().String(), resp.Hashs[i])
			errorHashs = append(errorHashs, resp.Hashs[i])
			errorIndexs = append(errorIndexs, resp.Indexs[i])
			errorInfos = append(errorInfos, "receive a wrong block")
			continue
		}
		if err := fs.PutBlock(blk); err != nil {
			log.Error(err)
			errorHashs = append(errorHashs, resp.Hashs[i])
			errorIndexs = append(errorIndexs, resp.Indexs[i])
			errorInfos = append(errorInfos, err.Error())
			continue
		}
		log.Debugf("put block success %v-%s-%d", resp.Blocks[i], resp.Hashs[i], resp.Indexs[i])

		tag, _ := hex.DecodeString(resp.Tags[i])
		if err := fs.PutTag(resp.Hashs[i], resp.FileHash, resp.Indexs[i], tag); err != nil {
			errorHashs = append(errorHashs, resp.Hashs[i])
			errorIndexs = append(errorIndexs, resp.Indexs[i])
			errorInfos = append(errorInfos, err.Error())
			continue
		}
		log.Debugf(" put tag or done %s-%s-%d", resp.FileHash, resp.Hashs[i], resp.Indexs[i])

		blkInfos = append(blkInfos, &transport.UploadFileDetail{
			PeerAddr: resp.PeerAddr,
			FileHash: resp.FileHash,
			Hash:     resp.Hashs[i],
			Index:    resp.Indexs[i],
			// DataOffset: uint64(resp.Offset),
			NodeList: []string{resp.PeerAddr},
		})
		unknowStateHashs = append(unknowStateHashs, resp.Hashs[i])
		unknowStateIndexs = append(unknowStateIndexs, resp.Indexs[i])
		unknowStateErrors = append(unknowStateErrors, "save put info to loacl error")
	}
	if len(blkInfos) > 0 {
		task := transport.NewTask()

		err := task.UploadPutBlocks(blkInfos)
		if err != nil {
			errorHashs = append(errorHashs, unknowStateHashs...)
			errorIndexs = append(errorIndexs, unknowStateIndexs...)
			errorInfos = append(errorInfos, unknowStateErrors...)
		} else {
			putSuccessMap["hashs"] = unknowStateHashs
			putSuccessMap["indexs"] = unknowStateIndexs
			uploadResultData.PutSucess = putSuccessMap
		}
	}
	if len(errorHashs) > 0 {
		putErrorMap["hashs"] = errorHashs
		putErrorMap["indexs"] = errorIndexs
		putErrorMap["infos"] = errorInfos
		uploadResultData.PutError = putErrorMap
	}

	log.Debug("fileUplaod success")
	//w.WriteHeader(200)
	w.Header().Set("content-type", "application/json;charset=utf-8")
	res, _ := json.Marshal(JsonResult{
		Code: 200,
		Msg:  "success",
		Data: uploadResultData,
	})
	w.Write(res)
}

func UploadCompleted(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	query := r.URL.Query()
	fileHash := query.Get("fileHash")
	peerAddr := query.Get("peerAddr")
	w.Header().Set("content-type", "application/json;charset=utf-8")
	if fileHash == "" || peerAddr == "" {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}

	fs := dsp.DspService.GetFS()
	err := fs.StartPDPVerify(fileHash)
	m := make(map[string]string)
	if err != nil {
		log.Errorf("start pdp verify err:%s", err)
		m["err"] = "start pdp verify err"
		res, _ := json.Marshal(JsonResult{Code: 500, Msg: "error", Data: m})
		w.Write(res)
		return
	}
	task := transport.NewTask()
	err = task.TransportCompleted("0", peerAddr, fileHash)
	if err != nil {
		if err.Error() == "leveldb: not found" {
			log.Errorf("loacal data storage err :%s", err)
			m["data"] = "0"
			m["err"] = "not found doing file,fileHash:" + fileHash
			res, _ := json.Marshal(JsonResult{Code: 200, Msg: "error", Data: m})
			w.Write(res)
		} else {
			log.Errorf("loacal data storage err :%s", err)
			m["err"] = "loacal data storage err"
			res, _ := json.Marshal(JsonResult{Code: 500, Msg: "error", Data: m})
			w.Write(res)
		}

		return
	}
	m["data"] = "1"
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)
}

//local upload record
func UploadRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}

	query := r.URL.Query()

	peerAddr := query.Get("peerAddr")
	taskState := query.Get("taskState")
	w.Header().Set("content-type", "application/json;charset=utf-8")
	if taskState == "" || peerAddr == "" {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	task := transport.NewTask()
	uploadFileList := task.UploadTaskRecord(peerAddr, taskState)
	m := make(map[string][]*transport.UploadFile)
	m["uploadFileList"] = uploadFileList
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)
}

func UploadRecordDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	query := r.URL.Query()
	peerAddr := query.Get("peerAddr")
	fileHash := query.Get("fileHash")
	taskState := query.Get("taskState")
	w.Header().Set("content-type", "application/json;charset=utf-8")
	if taskState == "" || peerAddr == "" || fileHash == "" {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	task := transport.NewTask()
	uploadFile, _ := task.GetUploadTaskDetail(peerAddr, fileHash, taskState)
	m := make(map[string]*transport.UploadFile)
	m["uploadFile"] = uploadFile
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)
}

func CreateDownloadTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	fileOwnerAddr := r.URL.Query().Get("fileOwnerAddr")
	peerAddr := r.URL.Query().Get("peerAddr")
	fileHash := r.URL.Query().Get("fileHash")
	w.Header().Set("content-type", "application/json;charset=utf-8")
	if fileHash == "" || peerAddr == "" || fileOwnerAddr == "" {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	m := make(map[string]interface{})
	task := transport.NewTask()
	hashs, indexs, prefix, err := task.GetUploadBlockHashs(fileOwnerAddr, fileHash)
	if err != nil {
		m["err"] = err.Error()
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	downloadTaskId, err := task.CreateDownloadTask(peerAddr, fileHash)
	if err != nil {
		m["err"] = err.Error()
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}

	m["createState"] = 1
	m["downloadTaskId"] = downloadTaskId
	m["prefix"] = prefix
	m["hashs"] = hashs
	m["indexs"] = indexs
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)
}

func GetPaymentId(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {

		log.Error(err)
	}
	var resp PaymentRespanseData
	json.Unmarshal([]byte(body), &resp)
	m := make(map[string]interface{})
	task := transport.NewTask()
	paymentId, ammount, err := task.GetPaymentId(resp.Hashs)
	if err != nil {
		m["err"] = err.Error()
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	m["paymentId"] = paymentId
	m["ammount"] = ammount
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)
}

func FileDownload(w http.ResponseWriter, r *http.Request) {
	log.Debug("enter FileDownload")
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {

		log.Error(err)
	}
	var resp DownloadRespanseData
	json.Unmarshal([]byte(body), &resp)
	log.Debugf(" start download file , peerAddr:%s ,fileHash:%s, txHash:%s ,paymentId:%v, downloadTaskId:%v, indexs:%v ", resp.PeerAddr, resp.FileHash, resp.TxHash, resp.PaymentId, resp.DownloadTaskId, resp.Indexs)
	w.Header().Set("content-type", "application/json;charset=utf-8")
	if resp.FileHash == "" || resp.PeerAddr == "" || resp.TxHash == "" || len(resp.Hashs) <= 0 || resp.DownloadTaskId <= 0 || resp.PaymentId <= 0 || len(resp.Indexs) < 0 {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}

	if len(resp.Hashs) < 1 {
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: ""})
		w.Write(res)
		return
	}

	haveHashIndexs, _ := transport.FilterDownloadHash(resp.PaymentId, resp.Hashs)
	if len(haveHashIndexs) != len(resp.Hashs) {
		m := make(map[string]interface{})
		m["err"] = "ready download hashs is not equal to paymentId hashs "
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
	} else {
		resultMap := make(map[string]interface{})
		fs := dsp.DspService.GetFS()

		downloadResultData := DownloadResultData{}
		var errorHashs []string
		var errorInfos []string
		downloadFileBlockInfos := make([]*transport.DownloadFileDetail, 0)
		event, err := dsp.DspService.GetChain().GetSmartContractEvent(resp.TxHash)
		if err != nil {
			log.Errorf("handle payment msg, get smart contract event err %s for tx %s", err, resp.TxHash)
		}

		valid := false
		for _, n := range event.Notify {
			if n == nil || n.States == nil {
				continue
			}
			s, ok := n.States.(map[string]interface{})
			if !ok {
				continue
			}
			paymentIdByChain, ok := s["paymentId"].(float64)
			if !ok {
				log.Errorf("payment id convert err %T", s["paymentId"])
				continue
			}
			log.Debugf("get payment id %v %T from event %v, paymentMsg.PaymentId %v",
				s["paymentId"], s["paymentId"], paymentIdByChain, resp.PaymentId)
			if int32(paymentIdByChain) == resp.PaymentId {
				valid = true
				break
			}
		}
		if !valid {
			log.Errorf("paymentId valid fail")
			errorHashs = append(errorHashs, resp.Hashs...)
			errorInfos = append(errorInfos, "paymentId valid fail")

		} else {
			//paymentId valid success
			for i, hash := range resp.Hashs {
				data := fs.GetBlock(hash).RawData()
				block := hex.EncodeToString(data)
				downloadResultData.Blocks = append(downloadResultData.Blocks, block)
				downloadResultData.Hashs = append(downloadResultData.Hashs, hash)
				tagBytes, err := fs.GetTag(hash, resp.FileHash, resp.Indexs[i])
				tag := hex.EncodeToString(tagBytes)
				if err != nil {
					log.Errorf("get hash: %s tag fail", hash)
					downloadResultData.Tags = append(downloadResultData.Tags, "-1")
				} else {

					downloadResultData.Tags = append(downloadResultData.Tags, tag)
				}
				downloadFileBlockInfos = append(downloadFileBlockInfos, &transport.DownloadFileDetail{
					DownloadTaskId: resp.DownloadTaskId,
					Hash:           hash,
					Index:          resp.Indexs[i],
					NodeList:       []string{resp.PeerAddr},
					TxHash:         resp.TxHash,
					PeerAddr:       resp.PeerAddr,
				})
			}
		}

		if len(errorHashs) > 0 {
			errMap := make(map[string][]string)
			errMap["errorHashs"] = errorHashs
			errMap["errorInfos"] = errorInfos
			resultMap["error"] = errMap
		}

		if len(downloadFileBlockInfos) > 0 {
			task := transport.NewTask()
			task.DownloadPutBlocks(downloadFileBlockInfos, resp.PaymentId)
		}
		resultMap["downloadData"] = downloadResultData
		log.Debugf("save download info to local success , peerAddr:%s ,fileHash:%s, txHash:%s ,paymentId:%v, downloadTaskId:%v , indexs:%v ", resp.PeerAddr, resp.FileHash, resp.TxHash, resp.PaymentId, resp.DownloadTaskId, resp.Indexs)
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: resultMap})
		w.Write(res)
	}
}

func DownloadCompleted(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	query := r.URL.Query()
	downloadTaskId := query.Get("downloadTaskId")
	peerAddr := query.Get("peerAddr")
	w.Header().Set("content-type", "application/json;charset=utf-8")
	if downloadTaskId == "" || peerAddr == "" {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	m := make(map[string]string)
	task := transport.NewTask()
	//
	err := task.TransportCompleted("1", peerAddr, downloadTaskId)
	if err != nil {
		if err.Error() == "leveldb: not found" {
			log.Errorf("loacal data storage err :%s", err)
			m["data"] = "0"
			m["err"] = "not found doing file,downloadTaskId:" + downloadTaskId
			res, _ := json.Marshal(JsonResult{Code: 200, Msg: "error", Data: m})
			w.Write(res)
		} else {
			log.Errorf("loacal data storage err :%s", err)
			m["err"] = "loacal data storage err"
			res, _ := json.Marshal(JsonResult{Code: 500, Msg: "error", Data: m})
			w.Write(res)
		}

		return
	}
	m["data"] = "1"
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)
}

func DownloadRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	query := r.URL.Query()

	peerAddr := query.Get("peerAddr")
	taskState := query.Get("taskState")
	w.Header().Set("content-type", "application/json;charset=utf-8")
	if taskState == "" || peerAddr == "" {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	task := transport.NewTask()
	downloadFileList := task.DownloadTaskRecord(peerAddr, taskState)
	m := make(map[string][]*transport.DownloadFile)
	m["downloadFileList"] = downloadFileList
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)
}

func DownloadRecordDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	query := r.URL.Query()

	peerAddr := query.Get("peerAddr")
	downloadTaskId := query.Get("downloadTaskId")
	taskState := query.Get("taskState")
	w.Header().Set("content-type", "application/json;charset=utf-8")
	if taskState == "" || peerAddr == "" || downloadTaskId == "" {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	task := transport.NewTask()
	downloadFile, _ := task.GetDownloadTaskDetail(peerAddr, downloadTaskId, taskState)
	m := make(map[string]*transport.DownloadFile)
	m["downloadFile"] = downloadFile
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)

}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {

		log.Error(err)
	}
	var resp DeleteRespanseData
	json.Unmarshal([]byte(body), &resp)
	w.Header().Set("content-type", "application/json;charset=utf-8")
	if resp.PeerAddr == "" || resp.TransportState == "" || len(resp.TaskIds) <= 0 || len(resp.TaskTypes) <= 0 {
		m := make(map[string]string)
		m["error"] = "parameter error"
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
		w.Write(res)
		return
	}
	task := transport.NewTask()
	deleteStateList := task.DeleteTransportTask(resp.TaskTypes, resp.TransportState, resp.PeerAddr, resp.TaskIds)
	w.Header().Set("content-type", "application/json;charset=utf-8")
	m := make(map[string][]bool)
	m["deleteStateList"] = deleteStateList
	res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: m})
	w.Write(res)
}

type HttpServer struct {
	router   *Router
	Listener net.Listener
	server   *http.Server
}

func InitHttpServer() HttpServer {
	hs := HttpServer{}
	hs.router = NewRouter()
	hs.initGetHandler()
	hs.initPostHandler()
	transport.InitDB()
	return hs
}
func test(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("tdrfsd"))
}

// init get handler
func (hs *HttpServer) initGetHandler() {
	// add get to router
	hs.router.Get("/api/v1/p2p/http/file/nodeState", NodeState)
	hs.router.Get("/api/v1/p2p/http/file/createUploadTask", CreateUploadTask)
	hs.router.Get("/api/v1/p2p/http/file/nodeVerification", NodeVerification)
	hs.router.Get("/api/v1/p2p/http/file/uploadCompleted", UploadCompleted)
	hs.router.Get("/api/v1/p2p/http/file/uploadRecord", UploadRecord)
	hs.router.Get("/api/v1/p2p/http/file/uploadRecordDetail", UploadRecordDetail)

	hs.router.Get("/api/v1/p2p/http/file/createDownloadTask", CreateDownloadTask)
	hs.router.Get("/api/v1/p2p/http/file/downloadCompleted", DownloadCompleted)
	hs.router.Get("/api/v1/p2p/http/file/downloadRecord", DownloadRecord)
	hs.router.Get("/api/v1/p2p/http/file/downloadRecordDetail", DownloadRecordDetail)
	hs.router.Get("/api/v1/p2p/http/test", test)

}
func (hs *HttpServer) initPostHandler() {
	//add post to router
	hs.router.Post("/api/v1/p2p/http/file/upload", FileUpload)

	hs.router.Post("/api/v1/p2p/http/file/getPaymentId", GetPaymentId)
	hs.router.Post("/api/v1/p2p/http/file/download", FileDownload)

	hs.router.Post("/api/v1/p2p/http/file/deleteTask", DeleteTask)
}
func (hs *HttpServer) Start() error {

	hs.server = &http.Server{Handler: hs.router}
	p2pPort := strconv.Itoa(int(config.Parameters.BaseConfig.PortBase + uint32(config.Parameters.BaseConfig.HttpP2pPortOffset)))
	var err error
	log.Infof("Start p2p http listen at %v", p2pPort)
	hs.Listener, err = net.Listen("tcp", fmt.Sprintf(":%v", p2pPort))
	if err != nil {
		log.Fatal("net.Listen: ", err.Error())
		return err
	}
	err = hs.server.Serve(hs.Listener)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
		return err
	}
	return nil
}
func (hs *HttpServer) Stop() {
	if hs.server != nil {
		hs.server.Shutdown(context.Background())
		log.Error("Close restful ")
	}

}
