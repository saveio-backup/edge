package dsp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	dspCom "github.com/saveio/dsp-go-sdk/common"
	"github.com/saveio/dsp-go-sdk/store"
	"github.com/saveio/dsp-go-sdk/task"
	dspUtils "github.com/saveio/dsp-go-sdk/utils"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp/actor/client"
	"github.com/saveio/edge/dsp/storage"
	sdkcom "github.com/saveio/themis-go-sdk/common"
	chainSdkFs "github.com/saveio/themis-go-sdk/fs"
	"github.com/saveio/themis/cmd/utils"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	fs "github.com/saveio/themis/smartcontract/service/native/savefs"
)

type DspFileListType int

const (
	DspFileListTypeAll DspFileListType = iota
	DspFileListTypeImage
	DspFileListTypeDoc
	DspFileListTypeVideo
	DspFileListTypeMusic
)

type DeleteFileResp struct {
	dspCom.DeleteUploadFileResp
	IsUploaded bool
}

type TransferType int

const (
	transferTypeComplete TransferType = iota
	transferTypeUploading
	transferTypeDownloading
	transferTypeAll
)

type NodeProgress struct {
	HostAddr         string
	UploadSize       uint64
	RealUploadSize   uint64
	RealDownloadSize uint64
	DownloadSize     uint64
	Speed            uint64
}

type Transfer struct {
	Id             string
	FileHash       string
	FileName       string
	Url            string
	Type           TransferType
	Status         store.TaskState
	DetailStatus   task.TaskProgressState
	CopyNum        uint32
	Path           string
	IsUploadAction bool
	UploadSize     uint64
	DownloadSize   uint64
	FileSize       uint64
	RealFileSize   uint64
	Nodes          []*NodeProgress
	Progress       float64
	CreatedAt      uint64
	UpdatedAt      uint64
	Result         interface{} `json:",omitempty"`
	ErrorCode      uint32
	ErrMsg         string `json:",omitempty"`
	StoreType      uint32
	Encrypted      bool
}

type TransferlistResp struct {
	IsTransfering bool
	Type          TransferType
	Transfers     []*Transfer
}

type DownloadFileInfo struct {
	Hash        string
	Name        string
	Ext         string
	Size        uint64
	Fee         uint64
	FeeFormat   string
	Path        string
	DownloadDir string
}

type FileShareIncome struct {
	Name         string
	OwnerAddress string
	Profit       uint64
	ProfitFormat string
	SharedAt     uint64
}

type FileShareIncomeResp struct {
	TotalIncome       uint64
	TotalIncomeFormat string
	Incomes           []*FileShareIncome
}

type FileResp struct {
	Hash          string
	Name          string
	Encrypt       bool
	Url           string
	Size          uint64
	DownloadCount uint64
	ExpiredAt     uint64
	UpdatedAt     uint64
	Profit        uint64
	Privilege     uint64
	CurrentHeight uint64
	ExpiredHeight uint64
	StoreType     fs.FileStoreType
	RealFileSize  uint64
	Nodes         []NodeProveDetail
}

type NodeProveDetail struct {
	HostAddr    string
	WalletAddr  string
	PdpProveNum uint64
	State       int
	Index       int
	UploadSize  uint64
}
type NodeProveDetails []NodeProveDetail

func (s NodeProveDetails) Len() int {
	return len(s)
}

func (s NodeProveDetails) Less(i, j int) bool {
	return s[i].Index < s[j].Index
}

func (s NodeProveDetails) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type DownloadFilesInfo struct {
	Hash          string
	Name          string
	Path          string
	OwnerAddress  string
	Url           string
	Size          uint64
	DownloadCount uint64
	DownloadAt    uint64
	LastShareAt   uint64
	Profit        uint64
	ProfitFormat  string
	Privilege     uint64
	RealFileSize  uint64
}

type WhiteListRule struct {
	Addr          string
	StartHeight   uint64
	ExpiredHeight uint64
}

type Userspace struct {
	Used          uint64
	Remain        uint64
	ExpiredAt     uint64
	Balance       uint64
	CurrentHeight uint64
	ExpiredHeight uint64
}

type UserspaceRecordResp struct {
	Size       uint64
	ExpiredAt  uint64
	Cost       int64
	CostFormat string
}

type CalculateResp struct {
	TxFee            uint64
	TxFeeFormat      string
	StorageFee       uint64
	StorageFeeFormat string
	ValidFee         uint64
	ValidFeeFormat   string
}

type UserspaceCostResp struct {
	Fee          uint64
	FeeFormat    string
	Refund       uint64
	RefundFormat string
	TransferType storage.UserspaceTransferType
}

type FsContractSettingResp struct {
	DefaultCopyNum     uint64
	MaxCopyNum         uint64
	DefaultProvePeriod uint64
	MinProveInterval   uint64
	MinVolume          uint64
}

type FileTask struct {
	Id       string
	FileName string
	State    int
	Result   interface{}
	Code     uint64
	Error    string
}

type FileTaskResp struct {
	Tasks []*FileTask
}

type UploadFileFilterType int

const (
	UploadFileFilterTypeAll UploadFileFilterType = iota
	UploadFileFilterTypeDoing
	UploadFileFilterTypeDone
)

func (this *Endpoint) UploadFile(path, desc string, durationVal, intervalVal, privilegeVal, copyNumVal,
	storageTypeVal, realFileSizeVal interface{}, encryptPwd, url string,
	whitelist []string, share bool) (*fs.UploadOption, *DspErr) {
	f, err := os.Stat(path)
	if err != nil {
		return nil, &DspErr{Code: FS_UPLOAD_FILEPATH_ERROR,
			Error: fmt.Errorf("os stat file %s error: %s", path, err.Error())}
	}
	log.Debugf("path: %v, isDir: %t", path, f.IsDir())
	if f.IsDir() {
		return nil, &DspErr{Code: FS_UPLOAD_FILEPATH_ERROR,
			Error: fmt.Errorf("uploadFile error: %s is a directory", path)}
	}
	if len(this.dspNet.GetProxyServer()) > 0 && !this.dspNet.IsConnectionReachable(this.dspNet.GetProxyServer()) {
		return nil, &DspErr{Code: NET_PROXY_DISCONNECTED,
			Error: fmt.Errorf("proxy %s is unreachable", this.dspNet.GetProxyServer())}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	currentAccount := dsp.CurrentAccount()
	fsSetting, err := dsp.GetFsSetting()
	if err != nil {
		return nil, &DspErr{Code: FS_GET_SETTING_FAILED, Error: err}
	}
	currentHeight, err := dsp.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	bal, err := dsp.BalanceOf(dsp.Address())
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	if bal == 0 {
		return nil, &DspErr{Code: INSUFFICIENT_BALANCE, Error: err}
	}
	interval, ok := intervalVal.(float64)
	interval = interval / float64(config.BlockTime())
	if !ok || interval == 0 {
		interval = float64(fsSetting.DefaultProvePeriod)
	}
	if uint64(interval) < fsSetting.MinProveInterval {
		return nil, &DspErr{Code: FS_UPLOAD_INTERVAL_TOO_SMALL, Error: ErrMaps[FS_UPLOAD_INTERVAL_TOO_SMALL]}
	}
	storageType, _ := storageTypeVal.(float64)
	realFileSize, _ := realFileSizeVal.(float64)
	var fileSizeInKB uint64
	if uint64(realFileSize) > 0 {
		fileSizeInKB = uint64(realFileSize)
	} else {
		fileSizeInKB = uint64(f.Size() / 1024)
		if fileSizeInKB == 0 {
			fileSizeInKB = 1
		}
	}
	opt := &fs.UploadOption{
		FileDesc:      []byte(desc),
		ProveInterval: uint64(interval),
		StorageType:   uint64(storageType),
		FileSize:      uint64(fileSizeInKB),
	}
	if fs.FileStoreType(storageType) == fs.FileStoreTypeNormal {
		userspace, err := dsp.GetUserSpace(currentAccount.Address.ToBase58())
		if err != nil {
			return nil, &DspErr{Code: FS_GET_USER_SPACE_FAILED, Error: err}
		}
		log.Debugf("storageType %v, userspace.ExpireHeight %d, current: %d",
			storageType, userspace.ExpireHeight, currentHeight)
		if userspace.ExpireHeight <= uint64(currentHeight) {
			return nil, &DspErr{Code: DSP_USER_SPACE_EXPIRED, Error: ErrMaps[DSP_USER_SPACE_EXPIRED]}
		}
		if userspace.Remain < uint64(fileSizeInKB) {
			return nil, &DspErr{Code: DSP_USER_SPACE_NOT_ENOUGH, Error: ErrMaps[DSP_USER_SPACE_NOT_ENOUGH]}
		}
		opt.ExpiredHeight = userspace.ExpireHeight
	} else {
		duration, _ := durationVal.(float64)
		opt.ExpiredHeight = uint64(currentHeight + uint32(duration/float64(config.BlockTime())))
	}
	log.Debugf("opt.ExpiredHeight :%d, minInterval :%d, current: %d",
		opt.ExpiredHeight, fsSetting.MinProveInterval, currentHeight)
	if opt.ExpiredHeight < fsSetting.MinProveInterval+uint64(currentHeight) {
		return nil, &DspErr{Code: DSP_CUSTOM_EXPIRED_NOT_ENOUGH, Error: ErrMaps[DSP_CUSTOM_EXPIRED_NOT_ENOUGH]}
	}
	privilege, ok := privilegeVal.(float64)
	if !ok {
		privilege = fs.PUBLIC
	}
	opt.Privilege = uint64(privilege)
	copyNum, ok := copyNumVal.(float64)
	if !ok {
		copyNum = float64(fsSetting.DefaultCopyNum)
	}
	opt.CopyNum = uint64(copyNum)
	if len(url) == 0 {
		// random
		b := make([]byte, common.DSP_URL_RAMDOM_NAME_LEN/2)
		_, err := rand.Read(b)
		if err != nil {
			return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
		url = dspCom.FILE_URL_CUSTOM_HEADER + hex.EncodeToString(b)
	}
	find, err := dsp.QueryUrl(url, dsp.Address())
	if find != nil || err == nil {
		return nil, &DspErr{Code: DSP_UPLOAD_URL_EXIST, Error: fmt.Errorf("url exist err %s", err)}
	}
	opt.DnsURL = []byte(url)
	opt.RegisterDNS = len(url) > 0
	opt.BindDNS = len(url) > 0
	// check whitelist format
	whitelistObj := fs.WhiteList{
		Num:  uint64(len(whitelist)),
		List: make([]fs.Rule, 0, uint64(len(whitelist))),
	}
	whitelistM := make(map[string]struct{}, 0)
	log.Debugf("whitelist :%v, len: %d %d", whitelist, len(whitelistObj.List), cap(whitelistObj.List))
	for i, whitelistAddr := range whitelist {
		addr, err := chainCom.AddressFromBase58(whitelistAddr)
		if err != nil {
			return nil, &DspErr{Code: INVALID_WALLET_ADDRESS, Error: err}
		}
		if _, ok := whitelistM[whitelistAddr]; ok {
			continue
		}
		whitelistM[whitelistAddr] = struct{}{}
		log.Debugf("index :%d", i)
		whitelistObj.List = append(whitelistObj.List, fs.Rule{
			Addr:         addr,
			BaseHeight:   uint64(currentHeight),
			ExpireHeight: opt.ExpiredHeight,
		})
	}
	opt.WhiteList = whitelistObj
	opt.Share = share
	opt.Encrypt = len(encryptPwd) > 0
	opt.EncryptPassword = []byte(encryptPwd)
	optBuf, _ := json.Marshal(opt)
	log.Debugf("path %s, UploadOption :%s\n", path, optBuf)
	taskExist, err := dsp.UploadTaskExist(path)
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	if taskExist {
		return nil, &DspErr{Code: DSP_UPLOAD_FILE_EXIST, Error: ErrMaps[DSP_UPLOAD_FILE_EXIST]}
	}
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Errorf("panic recover err %v", e)
			}
		}()
		log.Debugf("upload file path %s, this.Dsp: %t", path, dsp == nil)
		ret, err := dsp.UploadFile("", path, opt)
		if err != nil {
			log.Errorf("upload failed err %s", err)
			return
		} else {
			log.Infof("upload file success: %v", ret)
		}
	}()
	return opt, nil
}

func (this *Endpoint) PauseUploadFile(taskIds []string) (*FileTaskResp, *DspErr) {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}
		exist := dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := dsp.PauseUpload(id)
		if err != nil {
			taskResp.Code = DSP_PAUSE_UPLOAD_FAIELD
			taskResp.Error = err.Error()
		}
		state, err := dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_RESUME_UPLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("pause upload failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	go this.notifyUploadingTransferList()
	return resp, nil
}

func (this *Endpoint) ResumeUploadFile(taskIds []string) (*FileTaskResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	bal, err := dsp.BalanceOf(dsp.Address())
	if err != nil || bal == 0 {
		return nil, &DspErr{Code: INSUFFICIENT_BALANCE, Error: ErrMaps[INSUFFICIENT_BALANCE]}
	}
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}
		exist := dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := dsp.ResumeUpload(id)
		log.Debugf("resume upload err %v", err)
		if err != nil {
			taskResp.Code = DSP_RESUME_UPLOAD_FAIELD
			taskResp.Error = err.Error()
		}
		state, err := dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_RESUME_UPLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("resume upload failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp, nil
}

func (this *Endpoint) RetryUploadFile(taskIds []string) (*FileTaskResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	bal, err := dsp.BalanceOf(dsp.Address())
	if err != nil || bal == 0 {
		return nil, &DspErr{Code: INSUFFICIENT_BALANCE, Error: ErrMaps[INSUFFICIENT_BALANCE]}
	}
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}
		exist := dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := dsp.RetryUpload(id)
		if err != nil {
			taskResp.Code = DSP_RETRY_UPLOAD_FAIELD
			taskResp.Error = err.Error()
		}
		state, err := dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_RETRY_UPLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("retry upload failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp, nil
}

func (this *Endpoint) CancelUploadFile(taskIds []string, gasLimit uint64) (*FileTaskResp, *DspErr) {
	defer func() {
		go this.notifyUploadingTransferList()
	}()
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	bal, err := dsp.BalanceOf(dsp.Address())
	if err != nil || bal == 0 {
		return nil, &DspErr{Code: INSUFFICIENT_BALANCE, Error: ErrMaps[INSUFFICIENT_BALANCE]}
	}
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}

	args := make([][]interface{}, 0, len(taskIds))
	for _, id := range taskIds {
		args = append(args, []interface{}{id})
	}
	request := func(arg []interface{}, respCh chan *dspUtils.RequestResponse) {
		taskResp := &FileTask{
			State: int(store.TaskStateCancel),
		}
		if len(arg) != 1 {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			respCh <- &dspUtils.RequestResponse{
				Result: taskResp,
			}
			return
		}
		id, ok := arg[0].(string)
		if !ok {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			respCh <- &dspUtils.RequestResponse{
				Result: taskResp,
			}
			return
		}
		taskResp.Id = id
		taskResp.FileName = dsp.GetTaskFileName(id)
		exist := dsp.IsTaskExist(id)
		if !exist {
			err := dsp.DeleteTaskIds([]string{id})
			if err != nil {
				taskResp.Code = DSP_CANCEL_TASK_FAILED
				taskResp.Error = err.Error()
			}
			log.Debugf("cancel no exist in memory task, upload file, id %s, resp %v", id, taskResp)
			respCh <- &dspUtils.RequestResponse{
				Result: taskResp,
			}
			return
		}
		deleteResp, err := dsp.CancelUpload(id, gasLimit)
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
			respCh <- &dspUtils.RequestResponse{
				Result: taskResp,
			}
			return
		}
		taskResp.Result = deleteResp
		err = dsp.DeleteTaskIds([]string{id})
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
		}
		respCh <- &dspUtils.RequestResponse{
			Result: taskResp,
		}
	}
	requestResps := dspUtils.CallRequestWithArgs(request, args)
	for _, r := range requestResps {
		resp.Tasks = append(resp.Tasks, r.Result.(*FileTask))
	}
	return resp, nil
}

func (this *Endpoint) DeleteUploadFile(fileHash string, gasLimit uint64) (*DeleteFileResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	fi, err := dsp.GetFileInfo(fileHash)
	if fi == nil && dsp.IsFileInfoDeleted(err) {
		log.Debugf("file info is deleted: %v, %s", fi, err)
		return nil, nil
	}
	if fi != nil && err == nil && fi.FileOwner.ToBase58() == dsp.WalletAddress() {
		taskId := dsp.GetUploadTaskId(fileHash)
		if len(taskId) == 0 {
			tx, _, deletErr := dsp.DeleteUploadFilesFromChain([]string{fileHash}, gasLimit)
			if deletErr != nil {
				return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: deletErr}
			}
			resp := &DeleteFileResp{IsUploaded: true}
			resp.Tx = tx
			resp.FileHash = fileHash
			return resp, nil
		}
		result, err := dsp.DeleteUploadedFileByIds([]string{taskId}, gasLimit)
		if err != nil {
			log.Errorf("[Endpoint DeleteUploadFile] delete upload file failed, err %s", err)
			return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: err}
		}
		if len(result) == 0 {
			return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: ErrMaps[DSP_DELETE_FILE_FAILED]}
		}
		deleteResp := result[0]
		resp := &DeleteFileResp{IsUploaded: true}
		resp.Tx = deleteResp.Tx
		resp.FileHash = deleteResp.FileHash
		resp.FileName = deleteResp.FileName
		resp.Nodes = deleteResp.Nodes
		return resp, nil
	}
	log.Debugf("fi :%v, err :%v", fi, err)
	return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: err}
}

func (this *Endpoint) DeleteDownloadFile(fileHash string) (*DeleteFileResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	err := dsp.DeleteDownloadedLocalFile(fileHash)
	if err != nil {
		return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: err}
	}
	return nil, nil
}

func (this *Endpoint) DeleteUploadFiles(fileHashes []string, gasLimit uint64) ([]*DeleteFileResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	taskIds := make([]string, 0, len(fileHashes))
	for _, fileHash := range fileHashes {
		taskId := dsp.GetUploadTaskId(fileHash)
		taskHash := dsp.GetTaskFileHash(taskId)
		if len(taskId) == 0 || fileHash != taskHash {
			continue
		}
		taskIds = append(taskIds, taskId)
	}
	if len(taskIds) == 0 {
		tx, _, serr := dsp.DeleteUploadFilesFromChain(fileHashes, gasLimit)
		if serr != nil {
			return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: ErrMaps[DSP_DELETE_FILE_FAILED]}
		}
		resps := make([]*DeleteFileResp, 0, len(fileHashes))
		for _, hash := range fileHashes {
			resp := &DeleteFileResp{IsUploaded: true}
			resp.Tx = tx
			resp.FileHash = hash
			resps = append(resps, resp)
		}
		return resps, nil
	}
	result, err := dsp.DeleteUploadedFileByIds(taskIds, gasLimit)
	if err != nil {
		return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: err}
	}
	if len(result) == 0 {
		return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: ErrMaps[DSP_DELETE_FILE_FAILED]}
	}
	resps := make([]*DeleteFileResp, 0, len(result))
	for _, r := range result {
		resp := &DeleteFileResp{IsUploaded: true}
		resp.Tx = r.Tx
		resp.FileHash = r.FileHash
		resp.FileName = r.FileName
		resp.Nodes = r.Nodes
		resps = append(resps, resp)
	}
	return resps, nil
}

func (this *Endpoint) CalculateDeleteFilesFee(fileHashes []string) (*dspCom.Gas, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	preExecFee, err := dsp.GetDeleteFilesStorageFee(this.getDspWalletAddr(), fileHashes)
	if err != nil {
		return &dspCom.Gas{GasPrice: sdkcom.GAS_PRICE, GasLimit: preExecFee}, &DspErr{Code: FS_DELETE_CALC_FEE_FAILED, Error: err}
	}
	return &dspCom.Gas{GasPrice: sdkcom.GAS_PRICE, GasLimit: preExecFee}, nil
}

func (this *Endpoint) GetFsConfig() (*FsContractSettingResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	set, err := dsp.GetFsSetting()
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	info, err := dsp.GetNodeList()
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	maxCopyNum := uint64(0)
	if info.NodeNum >= 1 {
		maxCopyNum = info.NodeNum - 1
	}
	return &FsContractSettingResp{
		DefaultCopyNum:     set.DefaultCopyNum,
		MaxCopyNum:         maxCopyNum,
		DefaultProvePeriod: set.DefaultProvePeriod * config.BlockTime(),
		MinProveInterval:   set.MinProveInterval,
		MinVolume:          set.MinVolume,
	}, nil
}

func (this *Endpoint) DownloadFile(fileHash, url, linkStr, password string, max uint64,
	setFileName, inOrder bool) *DspErr {
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	// if balance of current channel is not enough, reject
	if !dsp.HasDNS() {
		return &DspErr{Code: DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST, Error: ErrMaps[DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST]}
	}
	if !this.channelNet.IsConnectionReachable(dsp.CurrentDNSHostAddr()) {
		return &DspErr{Code: DSP_CHANNEL_DNS_OFFLINE, Error: ErrMaps[DSP_CHANNEL_DNS_OFFLINE]}
	}

	fileInfo, err := this.GetDownloadFileInfo(url)
	if err != nil {
		return err
	}
	if len(this.dspNet.GetProxyServer()) > 0 &&
		!this.dspNet.IsConnectionReachable(this.dspNet.GetProxyServer()) {
		return &DspErr{Code: NET_PROXY_DISCONNECTED,
			Error: fmt.Errorf("proxy %s is unreachable", this.dspNet.GetProxyServer())}
	}
	if len(this.channelNet.GetProxyServer()) > 0 &&
		!this.channelNet.IsConnectionReachable(this.channelNet.GetProxyServer()) {
		return &DspErr{Code: NET_PROXY_DISCONNECTED,
			Error: fmt.Errorf("proxy %s is unreachable", this.channelNet.GetProxyServer())}
	}

	canDownload := false
	//[NOTE] when this.QueryChannel works, replace this.GetAllChannels logic
	all, getChannelErr := dsp.AllChannels()
	if getChannelErr != nil {
		return &DspErr{Code: INTERNAL_ERROR, Error: getChannelErr}
	}
	if all == nil || len(all.Channels) == 0 {
		return &DspErr{Code: DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH, Error: ErrMaps[DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH]}
	}

	for _, ch := range all.Channels {
		if dsp.IsDNS(ch.Address) && ch.Balance >= fileInfo.Fee {
			canDownload = true
			break
		}
	}

	if !canDownload {
		return &DspErr{Code: DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH, Error: ErrMaps[DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH]}
	}
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return syncErr
	}
	if syncing {
		return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}

	if len(url) > 0 {
		hash := dsp.GetFileHashFromUrl(url)
		if len(hash) == 0 {
			return &DspErr{Code: INTERNAL_ERROR, Error: fmt.Errorf("file hash not found for url %s", url)}
		}
		info, _ := dsp.GetFileInfo(hash)
		if info != nil && !dsp.CheckFilePrivilege(info, hash, dsp.WalletAddress()) {
			return &DspErr{Code: DSP_NO_PRIVILEGE_TO_DOWNLOAD,
				Error: fmt.Errorf("user %s has no privilege to download this file", dsp.WalletAddress())}
		}

		go func() {
			defer func() {
				if e := recover(); e != nil {
					log.Errorf("panic recover err %v", e)
				}
			}()
			err := dsp.DownloadFileByUrl(url, dspCom.ASSET_USDT, inOrder, password, false, setFileName, int(max))
			if err != nil {
				log.Errorf("Downloadfile from url failed %s", err)
			}
		}()
		return nil
	}

	if len(fileHash) > 0 {
		info, _ := dsp.GetFileInfo(fileHash)
		if info != nil && !dsp.CheckFilePrivilege(info, fileHash, dsp.WalletAddress()) {
			return &DspErr{Code: DSP_NO_PRIVILEGE_TO_DOWNLOAD,
				Error: fmt.Errorf("user %s has no privilege to download this file", dsp.WalletAddress())}
		}
		go func() {
			defer func() {
				if e := recover(); e != nil {
					log.Errorf("panic recover err %v", e)
				}
			}()
			err := dsp.DownloadFileByHash(fileHash, dspCom.ASSET_USDT, inOrder, password, false, setFileName, int(max))
			if err != nil {
				log.Errorf("Downloadfile from url failed %s", err)
			}
		}()
		return nil
	}

	if len(linkStr) > 0 {
		link, err := dspUtils.DecodeLinkStr(linkStr)
		if err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
		hash := link.FileHashStr
		if len(hash) == 0 {
			return &DspErr{Code: INTERNAL_ERROR, Error: fmt.Errorf("file hash not found for url %s", hash)}
		}
		info, _ := dsp.GetFileInfo(hash)
		if info != nil && !dsp.CheckFilePrivilege(info, hash, dsp.WalletAddress()) {
			return &DspErr{Code: DSP_NO_PRIVILEGE_TO_DOWNLOAD,
				Error: fmt.Errorf("user %s has no privilege to download this file", dsp.WalletAddress())}
		}
		go func() {
			defer func() {
				if e := recover(); e != nil {
					log.Errorf("panic recover err %v", e)
				}
			}()
			err := dsp.DownloadFileByLink(linkStr, dspCom.ASSET_USDT, inOrder, password, false, setFileName, int(max))
			if err != nil {
				log.Errorf("Downloadfile from url failed %s", err)
			}
		}()
		return nil
	}
	return nil
}

func (this *Endpoint) PauseDownloadFile(taskIds []string) (*FileTaskResp, *DspErr) {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}

		exist := dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := dsp.PauseDownload(id)
		if err != nil {
			taskResp.Code = DSP_PAUSE_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
		}
		state, err := dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_PAUSE_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("pause download failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	go this.notifyDownloadingTransferList()
	return resp, nil
}

func (this *Endpoint) ResumeDownloadFile(taskIds []string) (*FileTaskResp, *DspErr) {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	if !dsp.HasDNS() {
		return nil, &DspErr{Code: DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST, Error: ErrMaps[DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST]}
	}
	canDownload := false
	all, getChannelErr := dsp.AllChannels()
	if getChannelErr != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: getChannelErr}
	}
	if all == nil || len(all.Channels) == 0 {
		return nil, &DspErr{Code: DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH, Error: ErrMaps[DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH]}
	}

	fee := uint64(0)
	for _, id := range taskIds {
		fee += uint64(dsp.GetDownloadTaskRemainSize(id))
	}
	for _, ch := range all.Channels {
		log.Debugf("ResumeDownloadFile %v ch.Balance : %v fileinfo.fee %v ", ch.Address, ch.Balance, fee, dsp.IsDNS(ch.Address))
		if dsp.IsDNS(ch.Address) && ch.Balance >= fee {
			canDownload = true
			break
		}
	}
	canDownload = true
	if !canDownload {
		return nil, &DspErr{Code: DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH, Error: ErrMaps[DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH]}
	}
	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return nil, syncErr
	}
	if syncing {
		return nil, &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}

	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}
		exist := dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := dsp.ResumeDownload(id)
		if err != nil {
			taskResp.Code = DSP_RESUME_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
		}
		state, err := dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_RESUME_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("resume download failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp, nil
}

func (this *Endpoint) RetryDownloadFile(taskIds []string) (*FileTaskResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}

	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}
		exist := dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := dsp.RetryDownload(id)
		if err != nil {
			taskResp.Code = DSP_RETRY_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("retry download failed %s", err)
		}
		state, err := dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_RETRY_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("retry download failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp, nil
}

func (this *Endpoint) CancelDownloadFile(taskIds []string) (*FileTaskResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id:       id,
			FileName: dsp.GetTaskFileName(id),
		}
		exist := dsp.IsTaskExist(id)
		if !exist {
			err := dsp.DeleteTaskIds([]string{id})

			if err != nil {
				taskResp.Code = DSP_CANCEL_TASK_FAILED
				taskResp.Error = err.Error()
			}
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := dsp.CancelDownload(id)
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
		}
		err = dsp.DeleteTaskIds([]string{id})
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
		}
		taskResp.State = int(store.TaskStateCancel)
		log.Debugf("cancel download file, id :%s, resp :%v", id, taskResp)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	go this.notifyDownloadingTransferList()
	return resp, nil
}

func (this *Endpoint) RegisterProgressCh() {
	dsp := this.getDsp()
	if dsp == nil {
		log.Errorf("dsp is nil, register progress channel failed")
		return
	}
	dsp.RegProgressChannel()
	for {
		select {
		case v, ok := <-dsp.ProgressChannel():
			// TODO: replace with list
			if !ok {
				log.Warnf("progress channel is closed")
				return
			}
			if v == nil {
				log.Warnf("progress channel receive nil info")
				continue
			}
			switch v.Type {
			case store.TaskTypeUpload:
				go this.notifyUploadingTransferList()
			case store.TaskTypeDownload:
				go this.notifyDownloadingTransferList()
			default:
			}
			for node, cnt := range v.Progress {
				switch v.Type {
				case store.TaskTypeUpload:
					log.Infof("file:%s, hash:%s, total:%d, peer:%v, uploaded:%d, progress:%f, speed: %d",
						v.FileName, v.FileHash, v.Total, node, cnt.Progress, float64(cnt.Progress)/float64(v.Total),
						cnt.AvgSpeed())
				case store.TaskTypeDownload:
					log.Infof("file:%s, hash:%s, total:%d, peer:%v, downloaded:%d, progress:%f, speed: %d",
						v.FileName, v.FileHash, v.Total, node, cnt.Progress, float64(cnt.Progress)/float64(v.Total),
						cnt.AvgSpeed())
				default:
				}
			}
			if v.Result != nil {
				switch v.Type {
				case store.TaskTypeUpload:
					go this.notifyUploadingTransferList()
				case store.TaskTypeDownload:
					go this.notifyDownloadingTransferList()
				default:
				}
			}
		case <-this.closeCh:
			dsp.CloseProgressChannel()
			return
		}
	}
}

func (this *Endpoint) DeleteTransferRecord(taskIds []string) (*FileTaskResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id:    id,
			State: int(store.TaskStateCancel),
		}
		err := dsp.DeleteTaskIds([]string{id})
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
		}
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp, nil
}

// GetTransferList. get transfer progress list
func (this *Endpoint) GetTransferList(pType TransferType, offset, limit uint32) (*TransferlistResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp := &TransferlistResp{
		IsTransfering: false,
		Type:          pType,
		Transfers:     []*Transfer{},
	}
	allType, reverse, includeFailed := false, false, true
	var infoType store.TaskType
	switch pType {
	case transferTypeUploading:
		infoType = store.TaskTypeUpload
	case transferTypeDownloading:
		infoType = store.TaskTypeDownload
	case transferTypeComplete:
		allType = true
		reverse = true
		includeFailed = false
	}
	ids := dsp.GetTaskIdList(offset, limit, infoType, allType, reverse, includeFailed)
	infos := make([]*Transfer, 0, len(ids))
	for idx, key := range ids {
		info := dsp.GetProgressInfo(key)
		if info == nil {
			log.Warnf("get progress failed %d for %s info %v", idx, key, info)
			continue
		}
		if len(info.TaskId) == 0 {
			continue
		}
		if pType == transferTypeUploading && info.Type != store.TaskTypeUpload {
			continue
		}
		if pType == transferTypeDownloading && info.Type != store.TaskTypeDownload {
			continue
		}
		pInfo := this.getTransferDetail(pType, info)
		if pInfo == nil {
			continue
		}
		if !resp.IsTransfering {
			resp.IsTransfering = (pType == transferTypeUploading || pType == transferTypeDownloading) && (pInfo.Status != store.TaskStateFailed && pInfo.Status != store.TaskStateDone)
		}
		infos = append(infos, pInfo)
	}
	resp.Transfers = infos
	return resp, nil
}

// GetTransferList. get transfer progress list
func (this *Endpoint) GetTransferDetail(pType TransferType, id string) (*Transfer, *DspErr) {
	if len(id) == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp := &Transfer{}
	info := dsp.GetProgressInfo(id)
	if info == nil {
		return resp, nil
	}
	pInfo := this.getTransferDetail(pType, info)
	if pInfo == nil {
		return resp, &DspErr{Code: INTERNAL_ERROR, Error: ErrMaps[INTERNAL_ERROR]}
	}
	return pInfo, nil
}

// GetTransferList. get transfer progress list
func (this *Endpoint) GetTransferDetailByUrl(pType TransferType, url string) (*Transfer, *DspErr) {
	if len(url) == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	id := dsp.GetDownloadTaskIdByUrl(url)
	log.Debugf("GetTransferDetailByUrl url %s id = %s", url, id)
	if len(id) == 0 {
		return nil, nil
	}
	tr, err := this.GetTransferDetail(pType, id)
	if tr != nil && tr.Url != url {
		log.Warnf("transfer old url is %s, but got %s", tr.Url, url)
		tr.Url = url
	}
	return tr, err
}

func (this *Endpoint) CalculateUploadFee(filePath string, durationVal, intervalVal, timesVal, copynumVal,
	whitelistVal, storeType interface{}) (*CalculateResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	currentAccount := dsp.CurrentAccount()
	fsSetting, err := dsp.GetFsSetting()
	if err != nil {
		return nil, &DspErr{Code: FS_GET_SETTING_FAILED, Error: err}
	}
	interval, err := OptionStrToFloat64(intervalVal, float64(fsSetting.DefaultProvePeriod))
	interval = interval / float64(config.BlockTime())
	if err != nil || interval == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	sType, _ := OptionStrToFloat64(storeType, 0)

	fi, err := os.Open(filePath)
	if err != nil {
		return nil, &DspErr{Code: FS_UPLOAD_GET_FILESIZE_FAILED, Error: err}
	}
	fileStat, err := fi.Stat()
	if err != nil {
		return nil, &DspErr{Code: FS_UPLOAD_GET_FILESIZE_FAILED, Error: err}
	}
	copyNum, err := OptionStrToFloat64(copynumVal, float64(fsSetting.DefaultCopyNum))
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	wh, err := OptionStrToFloat64(whitelistVal, 0)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	whitelist := fs.WhiteList{
		Num:  uint64(wh),
		List: make([]fs.Rule, uint64(wh)),
	}
	opt := &fs.UploadOption{
		FileDesc:        []byte{},
		FileSize:        uint64(fileStat.Size()),
		ProveInterval:   uint64(interval),
		CopyNum:         uint64(copyNum),
		StorageType:     uint64(sType),
		Encrypt:         false,
		EncryptPassword: []byte{},
		RegisterDNS:     true,
		BindDNS:         true,
		DnsURL:          nil,
		WhiteList:       whitelist,
		Share:           false,
		Privilege:       1,
	}

	currentHeight, err := dsp.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	if fs.FileStoreType(sType) == fs.FileStoreTypeNormal {
		userspace, err := dsp.GetUserSpace(currentAccount.Address.ToBase58())
		if err != nil {
			return nil, &DspErr{Code: FS_GET_USER_SPACE_FAILED, Error: err}
		}
		if userspace == nil {
			return nil, &DspErr{Code: FS_GET_USER_SPACE_FAILED, Error: err}
		}
		if userspace.ExpireHeight <= uint64(currentHeight) {
			return nil, &DspErr{Code: DSP_USER_SPACE_EXPIRED, Error: ErrMaps[DSP_USER_SPACE_EXPIRED]}
		}
		opt.ExpiredHeight = userspace.ExpireHeight
		log.Debugf("userspace.ExpireHeight %d, current %d, interval:%v",
			userspace.ExpireHeight, currentHeight, interval)
		fee, err := dsp.CalculateUploadFee(opt)
		if err != nil {
			log.Debugf("fee :%v, err %s", fee, err)
			return nil, &DspErr{Code: FS_UPLOAD_CALC_FEE_FAILED, Error: ErrMaps[FS_UPLOAD_CALC_FEE_FAILED]}
		}
		return &CalculateResp{
			TxFee:            fee.TxnFee,
			TxFeeFormat:      utils.FormatUsdt(fee.TxnFee),
			StorageFee:       fee.SpaceFee,
			StorageFeeFormat: utils.FormatUsdt(fee.SpaceFee),
			ValidFee:         fee.ValidationFee,
			ValidFeeFormat:   utils.FormatUsdt(fee.ValidationFee),
		}, nil
	}

	duration, err := OptionStrToFloat64(durationVal, 0)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	opt.ExpiredHeight = uint64(currentHeight) + uint64(duration/float64(config.BlockTime()))
	log.Debugf("opt :%v\n", opt)
	fee, err := dsp.CalculateUploadFee(opt)
	log.Debugf("fee :%v\n", fee)
	if err != nil {
		return nil, &DspErr{Code: DSP_CALC_UPLOAD_FEE_FAILED, Error: err}
	}
	return &CalculateResp{
		TxFee:            fee.TxnFee,
		TxFeeFormat:      utils.FormatUsdt(fee.TxnFee),
		StorageFee:       fee.SpaceFee,
		StorageFeeFormat: utils.FormatUsdt(fee.SpaceFee),
		ValidFee:         fee.ValidationFee,
		ValidFeeFormat:   utils.FormatUsdt(fee.ValidationFee),
	}, nil
}

func (this *Endpoint) GetDownloadFileInfo(url string) (*DownloadFileInfo, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	info := &DownloadFileInfo{}
	var fileLink string
	if strings.HasPrefix(url, dspCom.FILE_URL_CUSTOM_HEADER) ||
		strings.HasPrefix(url, dspCom.FILE_URL_CUSTOM_HEADER_PROTOCOL) {
		fileLink = dsp.GetLinkFromUrl(url)
	} else if strings.HasPrefix(url, dspCom.FILE_LINK_PREFIX) {
		fileLink = url
	} else if strings.HasPrefix(url, dspCom.PROTO_NODE_PREFIX) ||
		strings.HasPrefix(url, dspCom.RAW_NODE_PREFIX) {
		// TODO support get download file info from hash
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	if len(fileLink) == 0 {
		return nil, &DspErr{Code: DSP_GET_FILE_LINK_FAILED, Error: ErrMaps[DSP_GET_FILE_LINK_FAILED]}
	}
	link, err := dsp.GetLinkValues(fileLink)
	if err != nil {
		return nil, &DspErr{Code: DSP_GET_FILE_LINK_FAILED, Error: ErrMaps[DSP_GET_FILE_LINK_FAILED]}
	}
	info.Hash = link.FileHashStr
	info.Name = link.FileName
	info.Size = link.BlockNum * dspCom.CHUNK_SIZE / 1024
	extParts := strings.Split(info.Name, ".")
	if len(extParts) > 1 {
		info.Ext = extParts[len(extParts)-1]
	}
	info.Fee = link.BlockNum * dspCom.CHUNK_SIZE * common.DSP_DOWNLOAD_UNIT_PRICE
	info.FeeFormat = utils.FormatUsdt(info.Fee)
	info.Path = this.getDownloadFilePath(info.Name)
	info.DownloadDir = this.getDownloadFilePath("")
	return info, nil
}

func (this *Endpoint) EncryptFile(path, password string) *DspErr {
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	err := dsp.AESEncryptFile(path, password, path+".temp")
	if err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	err = os.Rename(path+".temp", path)
	if err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	return nil
}

func (this *Endpoint) DecryptFile(path, fileName, password string) *DspErr {
	filePrefix, prefix, err := dspUtils.GetPrefixFromFile(path)
	if err != nil {
		return &DspErr{Code: DSP_DECRYPTED_FILE_FAILED, Error: err}
	}
	if !dspUtils.VerifyEncryptPassword(password, filePrefix.EncryptSalt, filePrefix.EncryptHash) {
		return &DspErr{Code: DSP_FILE_DECRYPTED_WRONG_PWD, Error: ErrMaps[DSP_FILE_DECRYPTED_WRONG_PWD]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	if len(fileName) == 0 {
		fileName = filePrefix.FileName
	}
	err = dsp.AESDecryptFile(path, string(prefix), password, dspUtils.GetDecryptedFilePath(path, fileName))
	log.Debugf("decrypted file output %s", dspUtils.GetDecryptedFilePath(path, fileName))
	if err != nil {
		return &DspErr{Code: DSP_DECRYPTED_FILE_FAILED, Error: err}
	}
	return nil
}

func (this *Endpoint) GetFileRevene() (uint64, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	sum, err := dsp.SumRecordsProfit()
	if err != nil {
		return 0, &DspErr{Code: DB_SUM_SHARE_PROFIT_FAILED, Error: err}
	}
	return uint64(sum), nil
}

func (this *Endpoint) GetFileShareIncome(start, end, offset, limit uint64) (*FileShareIncomeResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp := &FileShareIncomeResp{}
	records, err := dsp.FineShareRecordsByCreatedAt(int64(start), int64(end), int64(offset), int64(limit))
	if err != nil {
		return nil, &DspErr{Code: DB_FIND_SHARE_RECORDS_FAILED, Error: err}
	}
	resp.Incomes = make([]*FileShareIncome, 0, len(records))
	for _, record := range records {
		if record.Profit == 0 {
			continue
		}
		resp.TotalIncome += record.Profit
		resp.Incomes = append(resp.Incomes, &FileShareIncome{
			Name:         record.FileName,
			OwnerAddress: record.FileOwner,
			Profit:       record.Profit,
			ProfitFormat: utils.FormatUsdt(record.Profit),
			SharedAt:     uint64(record.CreatedAt),
		})
	}
	resp.TotalIncomeFormat = utils.FormatUsdt(resp.TotalIncome)
	return resp, nil
}

func (this *Endpoint) RegisterShareNotificationCh() {
	dsp := this.getDsp()
	if dsp == nil {
		log.Errorf("dsp is nil")
		return
	}
	dsp.RegShareNotificationChannel()
	for {
		select {
		case v, ok := <-dsp.ShareNotificationChannel():
			if !ok {
				break
			}
			log.Debugf("share notification taskkey=%s, filehash=%s, walletaddr=%s, state=%d, amount=%d",
				v.TaskKey, v.FileHash, v.ToWalletAddr, v.State, v.PaymentAmount)
			client.EventNotifyRevenue()

		case <-this.closeCh:
			dsp.CloseShareNotificationChannel()
			return
		}
	}
}

func (this *Endpoint) GetUploadFiles(fileType DspFileListType, offset, limit uint64,
	filterType UploadFileFilterType) ([]*FileResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	// rpc request
	curBlockHeight, err := dsp.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	taskInfos, err := dsp.GetUploadTaskInfos()
	if err != nil {
		return nil, &DspErr{Code: DSP_FILE_INFO_NOT_FOUND, Error: err}
	}
	files := make([]*FileResp, 0, limit)
	offsetCnt := uint64(0)
	for _, info := range taskInfos {
		if info == nil {
			continue
		}
		if info.ExpiredHeight < uint64(curBlockHeight) {
			continue
		}
		// 0: all, 1. image, 2. document. 3. video, 4. music
		if !FileNameMatchType(fileType, info.FileName) {
			continue
		}
		fileHashStr := info.FileHash
		if len(fileHashStr) == 0 {
			log.Warnf("task %s file hash is empty ", info.Id)
			continue
		}
		log.Debugf("get file of %s", fileHashStr)
		downloadedCount, _ := dsp.CountRecordByFileHash(fileHashStr)
		profit, _ := dsp.SumRecordsProfitByFileHash(fileHashStr)
		// rpc request
		proveDetail, err := dsp.GetFileProveDetails(fileHashStr)
		if err != nil {
			log.Errorf("get prove detail failed err %s", err)
			continue
		}
		nodesDetail := make([]NodeProveDetail, 0, proveDetail.ProveDetailNum)
		primaryNodeM := make(map[chainCom.Address]NodeProveDetail, 0)

		for index, addr := range info.PrimaryNodes {
			walletAddr, err := chainCom.AddressFromBase58(addr)
			if err != nil {
				continue
			}
			primaryNodeM[walletAddr] = NodeProveDetail{
				Index: index,
			}
		}
		fileHasUploaded := false
		for _, detail := range proveDetail.ProveDetails {
			nodeState := 2
			if detail.ProveTimes > 0 {
				nodeState = 3
				fileHasUploaded = true
			}
			uploadSize, _ := dsp.GetFileUploadSize(fileHashStr, string(detail.NodeAddr))
			if uploadSize > 0 {
				uploadSize /= 1024 // convert to KB
			}
			if detail.ProveTimes > 0 {
				uploadSize = info.FileSize
			}
			nodesDetail = append(nodesDetail, NodeProveDetail{
				HostAddr:    string(detail.NodeAddr),
				WalletAddr:  detail.WalletAddr.ToBase58(),
				PdpProveNum: detail.ProveTimes,
				State:       nodeState,
				Index:       primaryNodeM[detail.WalletAddr].Index,
				UploadSize:  uploadSize,
			})
			delete(primaryNodeM, detail.WalletAddr)
		}
		if filterType == UploadFileFilterTypeDoing && len(primaryNodeM) == 0 {
			continue
		}
		if filterType == UploadFileFilterTypeDone && len(primaryNodeM) > 0 {
			continue
		}
		if len(primaryNodeM) > 0 {
			unprovedNodeWallets := make([]chainCom.Address, 0)
			for addr, _ := range primaryNodeM {
				unprovedNodeWallets = append(unprovedNodeWallets, addr)
			}
			hostAddrs, err := dsp.GetNodeHostAddrListByWallets(unprovedNodeWallets)
			if err != nil {
				continue
			}
			for i, wallet := range unprovedNodeWallets {
				nodeDetail := primaryNodeM[wallet]
				nodeDetail.HostAddr = hostAddrs[i]
				nodeDetail.WalletAddr = wallet.ToBase58()
				uploadSize, _ := dsp.GetFileUploadSize(fileHashStr, string(nodeDetail.HostAddr))
				log.Debugf("file: %s, wallet %v, uploadsize %d", fileHashStr, wallet, uploadSize)
				if uploadSize > 0 {
					uploadSize /= 1024 // convert to KB
				}
				if uploadSize > info.FileSize {
					log.Warnf("update size is wrong %d, file size %d", uploadSize, info.FileSize)
					uploadSize = info.FileSize
				}
				nodeDetail.UploadSize = uploadSize
				primaryNodeM[wallet] = nodeDetail
			}
			for _, nodeDetail := range primaryNodeM {
				nodesDetail = append(nodesDetail, nodeDetail)
			}
		}
		if offsetCnt < offset {
			offsetCnt++
			continue
		}
		offsetCnt++
		sort.Sort(NodeProveDetails(nodesDetail))
		fileUrl := ""
		if fileHasUploaded && info.TaskState == store.TaskStateDone {
			fileUrl = info.Url
		}
		log.Debugf("fileHasUploaded %t, state %d", fileHasUploaded, info.TaskState)
		fr := &FileResp{
			Hash:          fileHashStr,
			Name:          info.FileName,
			Encrypt:       info.Encrypt,
			Url:           fileUrl,
			Size:          info.FileSize,
			DownloadCount: downloadedCount,
			ExpiredAt:     blockHeightToTimestamp(uint64(curBlockHeight), info.ExpiredHeight),
			UpdatedAt:     info.UpdatedAt / 1000,
			Profit:        profit,
			Privilege:     info.Privilege,
			CurrentHeight: uint64(curBlockHeight),
			ExpiredHeight: info.ExpiredHeight,
			StoreType:     fs.FileStoreType(info.StoreType),
			RealFileSize:  info.RealFileSize,
			Nodes:         nodesDetail,
		}
		files = append(files, fr)
		if limit > 0 && uint64(len(files)) >= limit {
			return files, nil
		}
	}
	return files, nil
}

type fileInfoResp struct {
	FileHash      string
	CreatedAt     uint64
	CopyNum       uint64
	Interval      uint64
	ProveTimes    uint64
	ExpiredHeight uint64
	Privilege     uint64
	OwnerAddress  string
	Whitelist     []string
	ExpiredAt     uint64
	CurrentHeight uint64
	Size          uint64
	RealFileSize  uint64
	StoreType     uint64
	BlocksRoot    string
}

func (this *Endpoint) GetFileInfo(fileHashStr string) (*fileInfoResp, *DspErr) {
	if len(fileHashStr) == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	info, err := dsp.GetFileInfo(fileHashStr)
	if err != nil {
		return nil, &DspErr{Code: DSP_FILE_INFO_NOT_FOUND, Error: ErrMaps[DSP_FILE_INFO_NOT_FOUND]}
	}

	now, err := dsp.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	expiredAt := blockHeightToTimestamp(uint64(now), info.ExpiredHeight)
	result := &fileInfoResp{
		FileHash:      string(info.FileHash),
		CopyNum:       info.CopyNum,
		Interval:      info.ProveInterval * config.BlockTime(),
		ProveTimes:    info.ProveTimes,
		ExpiredHeight: info.ExpiredHeight,
		Privilege:     info.Privilege,
		OwnerAddress:  info.FileOwner.ToBase58(),
		Whitelist:     []string{},
		ExpiredAt:     expiredAt,
		CurrentHeight: uint64(now),
		Size:          info.FileBlockNum * info.FileBlockSize,
		RealFileSize:  info.RealFileSize,
		StoreType:     info.StorageType,
		BlocksRoot:    string(info.BlocksRoot),
	}
	block, _ := dsp.GetBlockByHeight(uint32(info.BlockHeight))
	if block == nil {
		result.CreatedAt = uint64(time.Now().Unix())
	} else {
		result.CreatedAt = uint64(block.Header.Timestamp)
	}
	if info.Privilege != fs.WHITELIST {
		return result, nil
	}

	whitelist, err := dsp.GetWhiteList(fileHashStr)
	if err != nil || whitelist == nil {
		return result, nil
	}
	list := make([]string, 0, len(whitelist.List))
	for _, rule := range whitelist.List {
		list = append(list, rule.Addr.ToBase58())
	}
	result.Whitelist = list
	return result, nil
}

func (this *Endpoint) GetDownloadFiles(fileType DspFileListType, offset, limit uint64) (
	[]*DownloadFilesInfo, *DspErr) {
	fileInfos := make([]*DownloadFilesInfo, 0)
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	infos, _, err := dsp.AllDownloadFiles()
	if err != nil {
		return nil, &DspErr{Code: DB_GET_FILEINFO_FAILED, Error: ErrMaps[DB_GET_FILEINFO_FAILED]}
	}
	offsetCnt := uint64(0)
	for _, info := range infos {
		if info == nil {
			continue
		}
		exist := chainCom.FileExisted(info.FilePath)
		if !exist {
			log.Debugf("file not exist %s", info.FilePath)
			continue
		}
		file := info.FileHash
		url := info.Url

		// 0: all, 1. image, 2. document. 3. video, 4. music
		fileNameFromPath := filepath.Base(info.FilePath)
		if len(fileNameFromPath) == 0 {
			log.Warnf("can't get file name path :%s %s", info.FilePath, fileNameFromPath)
			fileNameFromPath = info.FileName
		}
		fileName := strings.ToLower(fileNameFromPath)
		if !FileNameMatchType(fileType, fileName) {
			continue
		}
		if offsetCnt < offset {
			offsetCnt++
			continue
		}
		offsetCnt++
		downloadedCount, _ := dsp.CountRecordByFileHash(file)
		profit, _ := dsp.SumRecordsProfitByFileHash(file)
		lastSharedAt, _ := dsp.FindLastShareTime(file)
		// TODO: get owner and privilege from DB
		fileInfo, _ := dsp.GetFileInfo(file)
		owner := ""
		privilege := uint64(fs.PUBLIC)
		if fileInfo != nil {
			owner = fileInfo.FileOwner.ToBase58()
			privilege = fileInfo.Privilege
		}
		if owner == "" && len(info.FileOwner) > 0 {
			owner = info.FileOwner
		}
		filePrefix := &dspUtils.FilePrefix{}
		filePrefix.Deserialize(info.Prefix)
		fileInfos = append(fileInfos, &DownloadFilesInfo{
			Hash:          file,
			Name:          fileNameFromPath,
			OwnerAddress:  owner,
			Url:           url,
			Size:          uint64(info.TotalBlockCount * dspCom.CHUNK_SIZE / 1024),
			DownloadCount: downloadedCount,
			DownloadAt:    info.CreatedAt / dspCom.MILLISECOND_PER_SECOND,
			LastShareAt:   lastSharedAt,
			Profit:        profit,
			ProfitFormat:  utils.FormatUsdt(profit),
			Path:          info.FilePath,
			Privilege:     privilege,
			RealFileSize:  filePrefix.FileSize,
		})
		if uint64(len(fileInfos)) > limit {
			break
		}
	}
	return fileInfos, nil
}

func (this *Endpoint) WhiteListOperation(fileHash string, op uint64, list []*WhiteListRule) (string, *DspErr) {
	if len(list) == 0 {
		return "", &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	li := fs.WhiteList{
		Num:  uint64(len(list)),
		List: make([]fs.Rule, 0, len(list)),
	}
	for _, l := range list {
		address, err := chainCom.AddressFromBase58(l.Addr)
		if err != nil {
			return "", &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
		}
		li.List = append(li.List, fs.Rule{
			Addr:         address,
			BaseHeight:   l.StartHeight,
			ExpireHeight: l.ExpiredHeight,
		})
	}
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.WhiteListOp(fileHash, op, li)
	if err != nil {
		return "", &DspErr{Code: DSP_WHITELIST_OP_FAILED, Error: err}
	}
	return tx, nil
}

func (this *Endpoint) GetWhitelist(fileHash string) ([]*WhiteListRule, *DspErr) {
	if len(fileHash) == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	list, err := dsp.GetWhiteList(fileHash)
	if err != nil {
		if strings.Contains(err.Error(), "not found!") {
			emptyList := make([]*WhiteListRule, 0)
			return emptyList, nil
		}
		return nil, &DspErr{Code: DSP_GET_WHITELIST_FAILED, Error: ErrMaps[DSP_GET_WHITELIST_FAILED]}
	}
	li := make([]*WhiteListRule, 0, list.Num)
	for _, l := range list.List {
		li = append(li, &WhiteListRule{
			Addr:          l.Addr.ToBase58(),
			StartHeight:   l.BaseHeight,
			ExpiredHeight: l.ExpireHeight,
		})
	}
	return li, nil
}

func (this *Endpoint) SetUserSpace(walletAddr string, size, sizeOpType, blockCount, countOpType uint64) (
	string, *DspErr) {
	if sizeOpType == uint64(fs.UserSpaceNone) && countOpType == uint64(fs.UserSpaceNone) {
		return "", &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	if sizeOpType == uint64(fs.UserSpaceNone) {
		size = 0
	}
	if countOpType == uint64(fs.UserSpaceNone) {
		blockCount = 0
	}
	blockCount = blockCount / config.BlockTime()
	tx, err := dsp.UpdateUserSpace(walletAddr, size, sizeOpType, blockCount, countOpType)
	if err != nil {
		return tx, ParseContractError(err)
	}
	_, err = dsp.PollForTxConfirmed(time.Duration(common.POLL_TX_COMFIRMED_TIMEOUT)*time.Second, tx)
	if err != nil {
		return "", &DspErr{Code: CHAIN_WAIT_TX_COMFIRMED_TIMEOUT, Error: err}
	}
	event, err := dsp.GetSmartContractEvent(tx)
	if err != nil || event == nil {
		log.Debugf("get event err %s, event :%v", err, event)
		if err := dsp.InsertUserspaceRecord(tx, walletAddr, size, store.UserspaceOperation(sizeOpType),
			blockCount*config.BlockTime(), store.UserspaceOperation(countOpType), 0, store.TransferTypeNone); err != nil {
			log.Errorf("insert userspace record err %s", err)
			return "", &DspErr{Code: DB_ADD_USER_SPACE_RECORD_FAILED, Error: err}
		}
		return tx, nil
	}
	hasTransfer := false
	for _, n := range event.Notify {
		states, ok := n.States.([]interface{})
		if !ok {
			continue
		}
		if len(states) < 4 || states[0] != "transfer" {
			continue
		}
		from := states[1].(string)
		to := states[2].(string)
		if to != chainSdkFs.FS_CONTRACT_ADDRESS.ToBase58() && to != dsp.WalletAddress() {
			continue
		}
		hasTransfer = true
		amount := states[3].(uint64)
		transferType := store.TransferTypeIn
		if to == walletAddr {
			transferType = store.TransferTypeOut
		}
		if err := dsp.InsertUserspaceRecord(tx, walletAddr, size, store.UserspaceOperation(sizeOpType),
			blockCount*config.BlockTime(), store.UserspaceOperation(countOpType), amount,
			transferType); err != nil {
			log.Errorf("insert userspace record err %s", err)
		}
		log.Debugf("from %s to %s amount %d", from, to, amount)
	}
	if len(event.Notify) == 0 || !hasTransfer {
		if err := dsp.InsertUserspaceRecord(tx, walletAddr, size, store.UserspaceOperation(sizeOpType),
			blockCount*config.BlockTime(), store.UserspaceOperation(countOpType),
			0, store.TransferTypeNone); err != nil {
			log.Errorf("insert userspace record err %s", err)
			return "", &DspErr{Code: DB_ADD_USER_SPACE_RECORD_FAILED, Error: err}
		}
		return tx, nil
	}
	return tx, nil
}

func (this *Endpoint) GetUserSpaceCost(walletAddr string, size, sizeOpType, blockCount, countOpType uint64) (
	*UserspaceCostResp, *DspErr) {
	if sizeOpType == uint64(fs.UserSpaceNone) && countOpType == uint64(fs.UserSpaceNone) {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	if sizeOpType == uint64(fs.UserSpaceNone) {
		size = 0
	}
	if countOpType == uint64(fs.UserSpaceNone) {
		blockCount = 0
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	blockCount = blockCount / config.BlockTime()
	cost, err := dsp.GetUpdateUserSpaceCost(walletAddr, size, sizeOpType, blockCount, countOpType)
	log.Debugf("cost %d %v %v %v %v %v, err %s", cost, walletAddr, size, sizeOpType, blockCount, countOpType, err)
	if err != nil {
		return nil, ParseContractError(err)
	}
	if cost.From.ToBase58() == dsp.WalletAddress() {
		return &UserspaceCostResp{
			Fee:          cost.Value,
			FeeFormat:    utils.FormatUsdt(cost.Value),
			TransferType: storage.TransferTypeIn,
		}, nil
	} else if cost.To.ToBase58() == dsp.WalletAddress() {
		return &UserspaceCostResp{
			Refund:       cost.Value,
			RefundFormat: utils.FormatUsdt(cost.Value),
			TransferType: storage.TransferTypeOut,
		}, nil
	}
	return nil, &DspErr{Code: INTERNAL_ERROR, Error: ErrMaps[INTERNAL_ERROR]}
}

func (this *Endpoint) GetUserSpace(addr string) (*Userspace, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	if len(addr) == 0 {
		addr = this.getDspWalletAddress()
	}
	space, err := dsp.GetUserSpace(addr)
	if err != nil || space == nil {
		return &Userspace{
			Used:      0,
			Remain:    0,
			ExpiredAt: 0,
			Balance:   0,
		}, nil
	}
	currentHeight, err := dsp.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	expiredAt := uint64(0)
	updateHeight := space.UpdateHeight
	now := uint64(time.Now().Unix())
	log.Debugf("space.ExpireHeight %v\n", space.ExpireHeight)
	if space.ExpireHeight > uint64(currentHeight) {
		expiredAt = blockHeightToTimestamp(uint64(currentHeight), space.ExpireHeight)
		log.Debugf("expiredAt %d currentHeight %d, expiredheight %d updatedheight %d",
			expiredAt, currentHeight, space.ExpireHeight, updateHeight)
	} else {
		spaceRecord, err := this.GetUserspaceRecords(addr, 0, 1)
		if err != nil || len(spaceRecord) == 0 {
			expiredAt = now
			log.Debugf("no space expiredAt %d ", expiredAt)
		} else {
			expiredAt = spaceRecord[0].ExpiredAt
			log.Debugf(" space[0] expiredAt %d ", expiredAt)
		}
	}
	log.Debugf("expiredAt %d, now %d  space.ExpireHeight:%d", expiredAt, now, space.ExpireHeight)
	if space.ExpireHeight <= uint64(currentHeight) {
		if expiredAt < now {
			return &Userspace{
				Used:          0,
				Remain:        0,
				ExpiredAt:     expiredAt,
				Balance:       space.Balance,
				CurrentHeight: uint64(currentHeight),
				ExpiredHeight: space.ExpireHeight,
			}, nil
		}
		return &Userspace{
			Used:          0,
			Remain:        0,
			ExpiredAt:     now - 1,
			Balance:       space.Balance,
			CurrentHeight: uint64(currentHeight),
			ExpiredHeight: space.ExpireHeight,
		}, nil
	}
	return &Userspace{
		Used:          space.Used,
		Remain:        space.Remain,
		ExpiredAt:     expiredAt,
		Balance:       space.Balance,
		CurrentHeight: uint64(currentHeight),
		ExpiredHeight: space.ExpireHeight,
	}, nil
}

func (this *Endpoint) GetUserspaceRecords(walletAddr string, offset, limit uint64) ([]*UserspaceRecordResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	records, err := dsp.SelectUserspaceRecordByWalletAddr(walletAddr, offset, limit)
	if err != nil {
		return nil, &DspErr{Code: DB_FIND_USER_SPACE_RECORD_FAILED, Error: err}
	}
	var resp []*UserspaceRecordResp
	if limit == 0 {
		resp = make([]*UserspaceRecordResp, 0)
	} else {
		resp = make([]*UserspaceRecordResp, 0, limit)
	}

	for _, record := range records {
		amount := int64(record.Amount)
		amountFormat := utils.FormatUsdt(record.Amount)
		if record.TransferType == store.TransferTypeOut {
			amount = -amount
			amountFormat = fmt.Sprintf("-%s", amountFormat)
		}
		resp = append(resp, &UserspaceRecordResp{
			Size:       record.TotalSize,
			ExpiredAt:  record.ExpiredAt,
			Cost:       amount,
			CostFormat: amountFormat,
		})
	}
	return resp, nil
}

func (this *Endpoint) GetStorageNodesInfo() (map[string]interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	info, err := dsp.GetNodeList()
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	m := make(map[string]interface{})
	m["Count"] = info.NodeNum
	return m, nil
}

func (this *Endpoint) GetProveDetail(fileHashStr string) (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	details, err := dsp.GetFileProveDetails(fileHashStr)
	if err != nil {
		return nil, &DspErr{Code: CONTRACT_ERROR, Error: err}
	}
	return details, nil
}

func (this *Endpoint) GetPeerCountOfHash(fileHashStr string) (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	return len(dsp.GetPeerFromTracker(fileHashStr, dsp.GetTrackerList())), nil
}

func (this *Endpoint) GetFileHashFromUrl(url string) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	return dsp.GetFileHashFromUrl(url), nil
}

func (this *Endpoint) UpdateFileUrlLink(url, hash, fileName string, fileSize, totalCount uint64) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	fileOwner := dsp.WalletAddress()
	blocksRoot := ""
	if fileSize == 0 || totalCount == 0 {
		info, _ := dsp.GetFileInfo(hash)
		if info != nil {
			fileSize = info.RealFileSize
			totalCount = info.FileBlockNum
			fileOwner = info.FileOwner.ToBase58()
			blocksRoot = string(info.BlocksRoot)
		} else {
			totalCount = fileSize * 1024 / dspCom.CHUNK_SIZE
		}
	}
	link := dsp.GenLink(hash, fileName, blocksRoot, fileOwner, uint64(fileSize), totalCount)
	tx, err := dsp.BindFileUrl(url, link)
	if err != nil {
		tx, err = dsp.RegisterFileUrl(url, link)
		if err != nil {
			return "", &DspErr{Code: CONTRACT_ERROR, Error: err}
		}
	}
	return tx, nil
}

func (this *Endpoint) getTransferDetail(pType TransferType, info *task.ProgressInfo) *Transfer {
	dsp := this.getDsp()
	if dsp == nil {
		return nil
	}
	if info.TaskState != store.TaskStateDone && info.TaskState != store.TaskStateFailed {
		// update state by task cache
		state, err := dsp.GetTaskState(info.TaskId)
		if err == nil {
			info.TaskState = state
		}
	}
	sum := uint64(0)
	nPros := make([]*NodeProgress, 0)
	for hAddr, cnt := range info.Progress {
		sum += uint64(cnt.Progress)
		pros := &NodeProgress{
			HostAddr: hAddr,
			Speed:    cnt.AvgSpeed(),
		}
		if info.Type == store.TaskTypeUpload {
			pros.UploadSize = uint64(cnt.Progress) * dspCom.CHUNK_SIZE / 1024
			pros.RealUploadSize = uint64(float64(cnt.Progress) / float64(info.Total) * float64(info.RealFileSize))
		} else if info.Type == store.TaskTypeDownload {
			pros.DownloadSize = uint64(cnt.Progress) * dspCom.CHUNK_SIZE / 1024
			pros.RealDownloadSize = uint64(float64(cnt.Progress) / float64(info.Total) * float64(info.RealFileSize))
		}
		nPros = append(nPros, pros)
	}
	pInfo := &Transfer{
		Id:           info.TaskId,
		FileHash:     info.FileHash,
		FileName:     info.FileName,
		Path:         info.FilePath,
		Url:          info.Url,
		CopyNum:      info.CopyNum,
		Type:         pType,
		StoreType:    info.StoreType,
		Status:       info.TaskState,
		DetailStatus: info.ProgressState,
		FileSize:     info.FileSize,
		RealFileSize: info.RealFileSize,
		Nodes:        nPros,
		CreatedAt:    info.CreatedAt,
		UpdatedAt:    info.UpdatedAt,
	}
	pInfo.IsUploadAction = (info.Type == store.TaskTypeUpload)
	pInfo.Progress = 0
	// log.Debugf("get transfer %s detail total %d sum %d ret %v err %s info.type %d", info.TaskKey, info.Total, sum, info.Result, info.ErrorMsg, info.Type)
	switch pType {
	case transferTypeUploading:
		if info.Total > 0 && sum >= uint64(info.Total) && info.Result != nil && len(info.ErrorMsg) == 0 {
			return nil
		}
		pInfo.UploadSize = sum * dspCom.CHUNK_SIZE / 1024
		if len(pInfo.Nodes) > 0 && pInfo.FileSize > 0 {
			pInfo.Progress = (float64(pInfo.UploadSize) / float64(pInfo.FileSize))
		}
		if pInfo.Progress == 1 && info.Result != nil {
			log.Warnf("info error msg of a success uploaded task %s", info.ErrorMsg)
			return nil
		}
		log.Debugf("info.total %d, sum :%d, info.result : %v, errormsg: %v, progress: %v",
			info.Total, sum, info.Result, info.ErrorMsg, pInfo.Progress)
	case transferTypeDownloading:
		if info.Total > 0 && sum > uint64(info.Total) {
			return nil
		}
		if info.TaskState == store.TaskStateDone {
			return nil
		}
		pInfo.DownloadSize = sum * dspCom.CHUNK_SIZE / 1024
		if pInfo.FileSize > 0 {
			pInfo.Progress = float64(pInfo.DownloadSize) / float64(pInfo.FileSize)
		}
	case transferTypeComplete:
		if sum < uint64(info.Total) || info.Total == 0 {
			return nil
		}
		if info.TaskState == store.TaskStateFailed {
			return nil
		}
		if info.Type == store.TaskTypeUpload {
			if info.Result == nil {
				return nil
			}
			pInfo.UploadSize = sum * dspCom.CHUNK_SIZE / 1024
			if pInfo.UploadSize == 0 {
				return nil
			}
			if pInfo.Status != store.TaskStateDone && pInfo.FileSize > 0 && pInfo.UploadSize == pInfo.FileSize {
				log.Warnf("task:%s taskstate is %d, status:%d, but it has done", info.TaskId, info.TaskState, pInfo.Status)
				pInfo.Status = store.TaskStateDone
			}
			if len(pInfo.Nodes) > 0 && pInfo.FileSize > 0 {
				pInfo.Progress = (float64(pInfo.UploadSize) / float64(pInfo.FileSize))
			}
		} else if info.Type == store.TaskTypeDownload {
			pInfo.DownloadSize = sum * dspCom.CHUNK_SIZE / 1024
			if pInfo.DownloadSize == 0 {
				return nil
			}
			if pInfo.Status != store.TaskStateDone && pInfo.FileSize > 0 && pInfo.DownloadSize == pInfo.FileSize {
				pInfo.Status = store.TaskStateDone
				log.Warnf("task:%s taskstate is %d, but it has done", info.TaskId, info.TaskState)
			}
			if pInfo.FileSize > 0 {
				pInfo.Progress = float64(pInfo.DownloadSize) / float64(pInfo.FileSize)
			}
			pInfo.Encrypted = dsp.IsFileEncrypted(pInfo.Path)
		}
	case transferTypeAll:
		if info.Type == store.TaskTypeUpload {
			pInfo.UploadSize = sum * dspCom.CHUNK_SIZE / 1024
			if pInfo.Status != store.TaskStateDone && pInfo.FileSize > 0 && pInfo.UploadSize == pInfo.FileSize {
				log.Warnf("task:%s taskstate is %d, status:%d, but it has done",
					info.TaskId, info.TaskState, pInfo.Status)
				pInfo.Status = store.TaskStateDone
			}
			if len(pInfo.Nodes) > 0 && pInfo.FileSize > 0 {
				pInfo.Progress = (float64(pInfo.UploadSize) / float64(pInfo.FileSize))
			}
		} else if info.Type == store.TaskTypeDownload {
			pInfo.DownloadSize = sum * dspCom.CHUNK_SIZE / 1024
			if pInfo.Status != store.TaskStateDone && pInfo.FileSize > 0 && pInfo.DownloadSize == pInfo.FileSize {
				pInfo.Status = store.TaskStateDone
				log.Warnf("task:%s taskstate is %d, but it has done", info.TaskId, info.TaskState)
			}
			if pInfo.FileSize > 0 {
				pInfo.Progress = float64(pInfo.DownloadSize) / float64(pInfo.FileSize)
			}
			pInfo.Encrypted = dsp.IsFileEncrypted(pInfo.Path)
		}
	}
	if info.TaskState == store.TaskStateFailed {
		pInfo.ErrMsg = info.ErrorMsg
		pInfo.ErrorCode = info.ErrorCode
	}
	if info.Result != nil {
		pInfo.Result = info.Result
	}
	return pInfo
}

func (this *Endpoint) getDownloadFilePath(fileName string) string {
	if len(fileName) == 0 {
		return config.FsFileRootPath()
	}
	return config.FsFileRootPath() + "/" + fileName
}

func blockHeightToTimestamp(curBlockHeight, endBlockHeight uint64) uint64 {
	now := uint64(time.Now().Unix())
	if endBlockHeight > curBlockHeight {
		return now + uint64(endBlockHeight-curBlockHeight)*config.BlockTime()
	}
	return now - uint64(curBlockHeight-endBlockHeight)*config.BlockTime()
}
