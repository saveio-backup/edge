package dsp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	dspConsts "github.com/saveio/dsp-go-sdk/consts"
	"github.com/saveio/dsp-go-sdk/store"
	dspTypes "github.com/saveio/dsp-go-sdk/task/types"
	dspLink "github.com/saveio/dsp-go-sdk/types/link"
	dspPrefix "github.com/saveio/dsp-go-sdk/types/prefix"
	"github.com/saveio/dsp-go-sdk/utils/async"
	dspTask "github.com/saveio/dsp-go-sdk/utils/task"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp/actor/client"
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
	dspTypes.DeleteUploadFileResp
	IsUploaded bool `json:"IsUploaded,omitempty"`
}

type TransferType int

const (
	transferTypeComplete TransferType = iota
	transferTypeUploading
	transferTypeDownloading
	transferTypeAll
)

// Upload or download progress of a remote peer
type NodeProgress struct {
	HostAddr         string
	UploadSize       uint64
	RealUploadSize   uint64
	RealDownloadSize uint64
	DownloadSize     uint64
	Speed            uint64
}

// Transfer detail
type Transfer struct {
	Id             string                     // task id
	FileHash       string                     // file hash
	FileName       string                     // file name
	Url            string                     // file download url
	Type           TransferType               // transfer type (uploading, downloading or complete)
	Status         store.TaskState            // task state
	DetailStatus   dspTypes.TaskProgressState // detail status of a task
	CopyNum        uint32                     // file copy number
	Path           string                     // file upload or download path
	IsUploadAction bool                       // is upload action or not
	UploadSize     uint64                     // upload size
	DownloadSize   uint64                     // download size
	FileSize       uint64                     // file size (block count * 256KiB)
	RealFileSize   uint64                     // file real size
	Fee            uint64                     // file download fee
	FeeFormat      string                     // file download fee with format
	Nodes          []*NodeProgress            // remote peer transfer progress
	Progress       float64                    // total task progress
	CreatedAt      uint64                     // task createdAt timestamp
	UpdatedAt      uint64                     // task updatedAt timestamp
	Result         interface{}                `json:",omitempty"` // transfer result
	ErrorCode      uint32                     // transfer error code
	ErrMsg         string                     `json:",omitempty"` // transfer error message
	StoreType      uint32                     // upload task store type
	Encrypted      bool                       // is encrypted or not
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
	TotalFile         int
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
	CreatedAt     uint64
	UpdatedAt     uint64
	Profit        uint64
	Privilege     uint64
	ProveLevel    uint64
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
	TransferType common.UserspaceTransferType
}

type FsContractSettingResp struct {
	DefaultCopyNum     uint64
	MaxCopyNum         uint64
	DefaultProvePeriod uint64
	DefaultProveLevel  uint64
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

// UploadFile: use dsp-go-sdk to create a upload task and upload the file
// taskId (string, optional): if it is not's specified, the sdk will create a random uuid string as taskId
// path   (string, required): file path
// dec    (string, optional): some description for the file, usually used as file name
// durationVal (uint64, required): duration for storing the file, uint second
// proveLevelVal (uint64, required): pdp prove level for file
// privilegeVal (uint64, required): privilege for file
// copyNumVal (uint64, required): copyNum of the file
// storageTypeVal (uint64, required): storage type, usespace mode or advance mode
// realFileSizeVal (uint64, required): real size for the file uint KiB
// encryptPwd (string, optional): encrypted password for the file
// url (string, optional): share url for the file
// whitelist ([]string, optional): if the file can only read by whitelist, here the wallet addresses
// share (bool, optional): reserved flag
func (this *Endpoint) UploadFile(taskId, path, desc string, durationVal, proveLevelVal, privilegeVal, copyNumVal,
	storageTypeVal, realFileSizeVal interface{}, encryptPwd, url string,
	whitelist []string, share bool) (*fs.UploadOption, *DspErr) {
	log.Debugf("upload task id %s", taskId)
	f, err := os.Stat(path)
	if err != nil {
		return nil, &DspErr{Code: FS_UPLOAD_FILEPATH_ERROR,
			Error: fmt.Errorf("os stat file %s error: %s", path, err.Error())}
	}
	if f.IsDir() {
		empty, err := IsDirEmpty(path)
		if err != nil {
			return nil, &DspErr{Code: FS_UPLOAD_FILEPATH_ERROR,
				Error: fmt.Errorf("check dir %s empty error: %s", path, err.Error())}
		}
		if empty {
			return nil, &DspErr{Code: FS_UPLOAD_FILEPATH_ERROR,
				Error: fmt.Errorf("dir %s is empty", path)}
		}
	}
	log.Debugf("path: %v, isDir: %t", path, f.IsDir())
	if len(this.dspNet.GetProxyServer().PeerID) > 0 &&
		!this.dspNet.IsConnReachable(this.dspNet.WalletAddrFromPeerId(this.dspNet.GetProxyServer().PeerID)) {
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
		return nil, &DspErr{Code: INSUFFICIENT_BALANCE, Error: ErrMaps[INSUFFICIENT_BALANCE]}
	}

	proveLevel, _ := ToUint64(proveLevelVal)
	if proveLevel == 0 {
		proveLevel = fsSetting.DefaultProveLevel
	}
	switch proveLevel {
	case fs.PROVE_LEVEL_HIGH:
	case fs.PROVE_LEVEL_MEDIEUM:
	case fs.PROVE_LEVEL_LOW:
	default:
		return nil, &DspErr{Code: FS_UPLOAD_INVALID_PROVE_LEVEL, Error: ErrMaps[FS_UPLOAD_INVALID_PROVE_LEVEL]}
	}
	storageType, _ := ToUint64(storageTypeVal)
	realFileSize, _ := ToUint64(realFileSizeVal)
	var fileSizeInKB uint64
	if uint64(realFileSize) > 0 {
		fileSizeInKB = uint64(realFileSize)
	} else {
		fileSizeInKB = uint64(f.Size() / 1000)
		if fileSizeInKB == 0 {
			fileSizeInKB = 1
		}
	}
	log.Debugf("fileSizeInKB %v, readlFileSizeVal %v", fileSizeInKB, realFileSizeVal)
	opt := &fs.UploadOption{
		FileDesc:    []byte(desc),
		ProveLevel:  uint64(proveLevel),
		StorageType: uint64(storageType),
		FileSize:    uint64(fileSizeInKB),
	}
	opt.ProveInterval = fs.GetProveIntervalByProveLevel(opt.ProveLevel)
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
		log.Debugf("opt.ExpiredHeight :%d, opt.Interval :%d, current: %d",
			opt.ExpiredHeight, opt.ProveInterval, currentHeight)
		if opt.ExpiredHeight < opt.ProveInterval+uint64(currentHeight) {
			return nil, &DspErr{Code: DSP_CUSTOM_EXPIRED_NOT_ENOUGH, Error: ErrMaps[DSP_CUSTOM_EXPIRED_NOT_ENOUGH]}
		}
	} else {
		duration, _ := ToUint64(durationVal)
		opt.ExpiredHeight = uint64(currentHeight + uint32(duration/config.BlockTime()))
		log.Debugf("opt.ExpiredHeight: %d, opt.Interval: %d, current: %d，duration: %d, blockTime: %d",
			opt.ExpiredHeight, opt.ProveInterval, currentHeight, duration, config.BlockTime())
		if opt.ExpiredHeight < opt.ProveInterval+uint64(currentHeight) {
			return nil, &DspErr{Code: DSP_USER_SPACE_PERIOD_NOT_ENOUGH, Error: ErrMaps[DSP_USER_SPACE_PERIOD_NOT_ENOUGH]}
		}
	}

	privilege, err := ToUint64(privilegeVal)
	if err != nil {
		privilege = fs.PUBLIC
	}
	opt.Privilege = uint64(privilege)
	if copyNumVal == nil {
		copyNumVal = fsSetting.DefaultCopyNum
	}
	log.Infof("copyNumVal+++ %v", copyNumVal)
	copyNum, err := ToUint64(copyNumVal)
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR,
			Error: fmt.Errorf("invalid copyNum %v error: %s", copyNumVal, err.Error())}
	}
	opt.CopyNum = uint64(copyNum)
	if len(url) == 0 {
		// random
		b := make([]byte, common.DSP_URL_RAMDOM_NAME_LEN/2)
		_, err := rand.Read(b)
		if err != nil {
			return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
		url = dspConsts.FILE_URL_CUSTOM_HEADER + hex.EncodeToString(b)
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
	whitelistM := make(map[string]struct{})
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
		return nil, &DspErr{Code: DSP_UPLOAD_FILE_EXIST, Error: fmt.Errorf("file %s %s", path, ErrMaps[DSP_UPLOAD_FILE_EXIST])}
	}
	go func() {
		// defer func() {
		// 	if e := recover(); e != nil {
		// 		log.Errorf("panic recover err %v", e)
		// 	}
		// }()
		log.Debugf("upload file path %s, this.Dsp: %t", path, dsp == nil)
		ret, err := dsp.UploadFile(true, taskId, path, opt)
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
	request := func(arg []interface{}, respCh chan *async.RequestResponse) {
		taskResp := &FileTask{
			State: int(store.TaskStateCancel),
		}
		if len(arg) != 1 {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			respCh <- &async.RequestResponse{
				Result: taskResp,
			}
			return
		}
		id, ok := arg[0].(string)
		if !ok {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			respCh <- &async.RequestResponse{
				Result: taskResp,
			}
			return
		}
		taskResp.Id = id
		taskResp.FileName = dsp.GetTaskFileName(id)
		exist := dsp.IsTaskExist(id)
		if !exist {
			err := dsp.CleanTasks([]string{id})
			if err != nil {
				taskResp.Code = DSP_CANCEL_TASK_FAILED
				taskResp.Error = err.Error()
			}
			log.Debugf("cancel no exist in memory task, upload file, id %s, resp %v", id, taskResp)
			respCh <- &async.RequestResponse{
				Result: taskResp,
			}
			return
		}
		deleteResp, err := dsp.CancelUpload(id, gasLimit)
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
			respCh <- &async.RequestResponse{
				Result: taskResp,
			}
			return
		}
		taskResp.Result = deleteResp
		err = dsp.CleanTasks([]string{id})
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
		}
		respCh <- &async.RequestResponse{
			Result: taskResp,
		}
	}
	requestResps := async.RequestWithArgs(request, args)
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
	defer func() {
		go this.notifyDownloadingTransferList()
	}()
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
			log.Errorf("delete upload files from chain err %s", serr)
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
		log.Errorf("delete upload files from ids err %s", err)
		return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: err}
	}
	if len(result) == 0 {
		log.Errorf("delete upload files from ids no result")
		return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: ErrMaps[DSP_DELETE_FILE_FAILED]}
	}
	resps := make([]*DeleteFileResp, 0, len(result))
	for _, r := range result {
		resp := &DeleteFileResp{IsUploaded: true}
		resp.Tx = r.Tx
		resp.FileHash = r.FileHash
		resp.FileName = r.FileName
		resps = append(resps, resp)
	}
	return resps, nil
}

func (this *Endpoint) CalculateDeleteFilesFee(fileHashes []string) (*dspTypes.Gas, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	preExecFee, err := dsp.GetDeleteFilesStorageFee(this.getDspWalletAddr(), fileHashes)
	if err != nil {
		return &dspTypes.Gas{GasPrice: sdkcom.GAS_PRICE, GasLimit: preExecFee}, &DspErr{Code: FS_DELETE_CALC_FEE_FAILED, Error: err}
	}
	return &dspTypes.Gas{GasPrice: sdkcom.GAS_PRICE, GasLimit: preExecFee}, nil
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
		DefaultProveLevel:  set.DefaultProveLevel,
		MinVolume:          set.MinVolume,
	}, nil
}

func (this *Endpoint) getChannelMinBalance() (uint64, *DspErr) {
	//[NOTE] when this.QueryChannel works, replace this.GetAllChannels logic
	dsp := this.getDsp()
	if dsp == nil {
		return 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	all, getChannelErr := dsp.AllChannels()
	if getChannelErr != nil {
		return 0, &DspErr{Code: INTERNAL_ERROR, Error: getChannelErr}
	}
	if all == nil || len(all.Channels) == 0 {
		return 0, &DspErr{Code: DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH, Error: ErrMaps[DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH]}
	}

	minChannelBalance := uint64(0)
	for _, ch := range all.Channels {
		if dsp.IsDNS(ch.Address) && minChannelBalance < ch.Balance {
			minChannelBalance = ch.Balance
		}
	}
	log.Debugf("all channel min balance %v", minChannelBalance)
	return minChannelBalance, nil
}

func (this *Endpoint) DownloadFile(taskId, fileHash, url, linkStr, password string, max uint64,
	setFileName, inOrder bool) *DspErr {
	dsp := this.getDsp()
	if dsp == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	log.Debugf("downlaod task id %s", taskId)
	// if balance of current channel is not enough, reject
	if !dsp.HasDNS() {
		return &DspErr{Code: DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST, Error: ErrMaps[DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST]}
	}
	if !this.channelNet.IsConnReachable(dsp.CurrentDNSWallet()) {
		return &DspErr{Code: DSP_CHANNEL_DNS_OFFLINE, Error: ErrMaps[DSP_CHANNEL_DNS_OFFLINE]}
	}

	if len(this.dspNet.GetProxyServer().PeerID) > 0 &&
		!this.dspNet.IsConnReachable(this.dspNet.WalletAddrFromPeerId(this.dspNet.GetProxyServer().PeerID)) {
		return &DspErr{Code: NET_PROXY_DISCONNECTED,
			Error: fmt.Errorf("proxy %s is unreachable", this.dspNet.GetProxyServer())}
	}

	syncing, syncErr := this.IsChannelProcessBlocks()
	if syncErr != nil {
		return syncErr
	}
	if syncing {
		return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	minChannelBalance, err := this.getChannelMinBalance()
	if err != nil {
		return err
	}
	if len(url) > 0 {
		fileInfoFromUrl, err := this.GetDownloadFileInfo(url)
		if err != nil {
			return err
		}
		if fileInfoFromUrl != nil && fileInfoFromUrl.Fee > minChannelBalance {
			return &DspErr{Code: DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH, Error: ErrMaps[DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH]}
		}
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
			// defer func() {
			// if e := recover(); e != nil {
			// 	log.Errorf("panic recover err %v", e)
			// }
			// }()
			err := dsp.DownloadFileByUrl(taskId, url, dspConsts.ASSET_USDT, inOrder, password, false, setFileName, int(max))
			if err != nil {
				log.Errorf("download task %s from url failed %s", taskId, err)
			}
		}()
		return nil
	}

	if len(fileHash) > 0 {
		if len(fileHash) != dspConsts.PROTO_NODE_FILE_HASH_LEN {
			return &DspErr{Code: INVALID_PARAMS, Error: fmt.Errorf("invalid file hash")}
		}
		info, _ := dsp.GetFileInfo(fileHash)
		if info != nil && !dsp.CheckFilePrivilege(info, fileHash, dsp.WalletAddress()) {
			return &DspErr{Code: DSP_NO_PRIVILEGE_TO_DOWNLOAD,
				Error: fmt.Errorf("user %s has no privilege to download this file", dsp.WalletAddress())}
		}
		if info != nil {
			fmt.Printf("info.FileBlockNum*info.FileBlockSize %v\n", info.FileBlockNum*info.FileBlockSize)
			if info.FileBlockNum*info.FileBlockSize*1024 > minChannelBalance {
				return &DspErr{Code: DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH, Error: ErrMaps[DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH]}
			}
		}
		go func() {
			// defer func() {
			// 	if e := recover(); e != nil {
			// 		log.Errorf("panic recover err %v", e)
			// 	}
			// }()
			err := dsp.DownloadFileByHash(taskId, fileHash, dspConsts.ASSET_USDT, inOrder, password, false, setFileName, int(max))
			if err != nil {
				log.Errorf("Download file from file hash failed %s", err)
			}
		}()
		return nil
	}

	if len(linkStr) > 0 {
		link, err := dspLink.DecodeLinkStr(linkStr)
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
		if info != nil && info.FileBlockNum*info.FileBlockSize*1024 > minChannelBalance {
			return &DspErr{Code: DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH, Error: ErrMaps[DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH]}
		}
		go func() {
			// defer func() {
			// 	if e := recover(); e != nil {
			// 		log.Errorf("panic recover err %v", e)
			// 	}
			// }()
			err := dsp.DownloadFileByLink(taskId, linkStr, dspConsts.ASSET_USDT, inOrder, password, false, setFileName, int(max))
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
			err := dsp.CleanTasks([]string{id})

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
		err = dsp.CleanTasks([]string{id})
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
			if v.ProgressState == dspTypes.TaskCreate && v.Type == store.TaskTypeUpload {
				log.Debugf("notify new upload task %s", v.TaskId)
				go this.notifyNewTransferTask(transferTypeUploading, v.TaskId)
			}
			if v.ProgressState == dspTypes.TaskCreate && v.Type == store.TaskTypeDownload {
				log.Debugf("notify new download task %s", v.TaskId)
				go this.notifyNewTransferTask(transferTypeDownloading, v.TaskId)
			}
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
		err := dsp.HideTaskIds([]string{id})
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
		}
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp, nil
}

// GetTransferList. get transfer progress list
func (this *Endpoint) GetTransferList(pType TransferType, offset, limit uint32, createdAt, createdAtEnd, updatedAt, updatedAtEnd uint64) (
	*TransferlistResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp := &TransferlistResp{
		IsTransfering: false,
		Type:          pType,
		Transfers:     []*Transfer{},
	}
	complete, reverse, includeFailed, ignoreHide := false, false, true, true
	var infoType store.TaskType
	switch pType {
	case transferTypeUploading:
		infoType = store.TaskTypeUpload
	case transferTypeDownloading:
		infoType = store.TaskTypeDownload
	case transferTypeComplete:
		complete = true
		reverse = true
		includeFailed = false
	}
	ids := dsp.GetTaskIdList(offset, limit, createdAt, createdAtEnd, updatedAt, updatedAtEnd, infoType,
		complete, reverse, includeFailed, ignoreHide)
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
func (this *Endpoint) GetProgressById(id string) (*dspTypes.ProgressInfo, *DspErr) {
	if len(id) == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp := &dspTypes.ProgressInfo{}
	info := dsp.GetProgressInfo(id)
	log.Debugf("get progress %v by %v", info, id)
	if info == nil {
		return resp, nil
	}
	return info, nil
}

// GetTransferList. get transfer progress list
func (this *Endpoint) GetTaskInfoById(id string) (*store.TaskInfo, *DspErr) {
	if len(id) == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp := &store.TaskInfo{}
	info := dsp.GetTaskInfo(id)
	log.Debugf("get progress %v by %v", info, id)
	if info == nil {
		return resp, nil
	}
	return info, nil
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

func (this *Endpoint) CalculateUploadFee(filePath string, durationVal, proveLevelVal, timesVal, copynumVal,
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

	proveLevel, err := OptionStrToFloat64(proveLevelVal, float64(fsSetting.DefaultProveLevel))
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}

	switch proveLevel {
	case fs.PROVE_LEVEL_HIGH:
	case fs.PROVE_LEVEL_MEDIEUM:
	case fs.PROVE_LEVEL_LOW:
	default:
		return nil, &DspErr{Code: FS_UPLOAD_INVALID_PROVE_LEVEL, Error: ErrMaps[FS_UPLOAD_INVALID_PROVE_LEVEL]}
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
		FileSize:        uint64(fileStat.Size() / 1024),
		ProveLevel:      uint64(proveLevel),
		ProveInterval:   fs.GetProveIntervalByProveLevel(uint64(proveLevel)),
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
			userspace.ExpireHeight, currentHeight, opt.ProveInterval)
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
	if strings.HasPrefix(url, dspConsts.FILE_URL_CUSTOM_HEADER) ||
		strings.HasPrefix(url, dspConsts.FILE_URL_CUSTOM_HEADER_PROTOCOL) {
		fileLink = dsp.GetLinkFromUrl(url)
	} else if strings.HasPrefix(url, dspConsts.FILE_LINK_PREFIX) {
		fileLink = url
	} else if strings.HasPrefix(url, dspConsts.PROTO_NODE_PREFIX) ||
		strings.HasPrefix(url, dspConsts.RAW_NODE_PREFIX) {
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
	info.Size = link.BlockNum * dspConsts.CHUNK_SIZE / 1024
	extParts := strings.Split(info.Name, ".")
	if len(extParts) > 1 {
		info.Ext = extParts[len(extParts)-1]
	}
	info.Fee = link.BlockNum * dspConsts.CHUNK_SIZE * common.DSP_DOWNLOAD_UNIT_PRICE
	if link.FileSize != 0 {
		info.Fee = link.FileSize * 1024 * common.DSP_DOWNLOAD_UNIT_PRICE
		// dag link's size
		info.Fee += link.BlockNum * 50
		// mediate node' size calculate by children max
		info.Fee += link.BlockNum / 11 * 50
	}
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
	stat, err := os.Stat(path)
	if err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	prefix := dspPrefix.NewEncryptPrefix(password, this.getDspWalletAddr(), uint64(stat.Size()), stat.IsDir())
	if prefix == nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: errors.New("prefix is nil")}
	}
	tempOutput := path + ".temp"
	prefixBuf := prefix.Serialize()
	output := path + ".ept"
	if err := dsp.AESEncryptFile(path, password, tempOutput); err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	log.Debugf("+++++ prefix %s", prefixBuf)
	outputFile, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	defer outputFile.Close()
	if _, err := outputFile.Write(prefixBuf); err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	remain, err := os.Open(tempOutput)
	if err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	if _, err := outputFile.Seek(int64(len(prefixBuf)), io.SeekStart); err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	if _, err := io.Copy(outputFile, remain); err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	if err := os.Remove(tempOutput); err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	return nil
}

func (this *Endpoint) DecryptFile(path, fileName, password string) (string, *DspErr) {
	filePrefix, prefix, err := dspPrefix.GetPrefixFromFile(path)
	if err != nil {
		return "", &DspErr{Code: DSP_DECRYPTED_FILE_FAILED, Error: err}
	}
	if !dspPrefix.VerifyEncryptPassword(password, filePrefix.EncryptSalt, filePrefix.EncryptHash) {
		return "", &DspErr{Code: DSP_FILE_DECRYPTED_WRONG_PWD, Error: ErrMaps[DSP_FILE_DECRYPTED_WRONG_PWD]}
	}
	log.Debugf("verified %s, %s", password, prefix)
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	if len(fileName) == 0 {
		fileName = filePrefix.FileName
	}
	outPath := dspTask.GetDecryptedFilePath(path, fileName)
	err = dsp.AESDecryptFile(path, string(prefix), password, outPath)
	log.Debugf("decrypted file output %s", dspTask.GetDecryptedFilePath(path, fileName))
	if err != nil {
		return "", &DspErr{Code: DSP_DECRYPTED_FILE_FAILED, Error: err}
	}
	return outPath, nil
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
	records, totalFile, err := dsp.FindShareRecordsByCreatedAt(int64(start), int64(end), int64(offset), int64(limit))
	if err != nil {
		return nil, &DspErr{Code: DB_FIND_SHARE_RECORDS_FAILED, Error: err}
	}
	resp.Incomes = make([]*FileShareIncome, 0, len(records))
	for _, record := range records {
		if record.Profit == 0 {
			totalFile--
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
	if totalFile < 0 {
		totalFile = 0
	}
	resp.TotalIncomeFormat = utils.FormatUsdt(resp.TotalIncome)
	resp.TotalFile = totalFile
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

func (this *Endpoint) GetUploadFiles(fileType DspFileListType, offset, limit, createdAt, createdAtEnd, updatedAt, updatedAtEnd uint64,
	filterType UploadFileFilterType) ([]*FileResp, int, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	// rpc request
	curBlockHeight, err := dsp.GetCurrentBlockHeight()
	if err != nil {
		return nil, 0, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	taskInfos, err := dsp.GetUploadTaskInfos()
	if err != nil {
		return nil, 0, &DspErr{Code: DSP_FILE_INFO_NOT_FOUND, Error: err}
	}
	log.Debugf("total upload task info length %v", len(taskInfos))
	totalCount := 0
	files := make([]*FileResp, 0, limit)
	offsetCnt := uint64(0)
	for _, info := range taskInfos {
		if info == nil {
			// log.Debugf("get upload list skip because info is nil")
			continue
		}
		if info.ExpiredHeight < uint64(curBlockHeight) {
			// log.Debugf("get upload list skip because info is expired")
			continue
		}
		if createdAt != 0 && createdAtEnd != 0 && (info.CreatedAt <= createdAt || info.CreatedAt > createdAtEnd) {
			// log.Debugf("get upload list skip because info is out of query date range")
			continue
		}
		if updatedAt != 0 && updatedAtEnd != 0 && (info.UpdatedAt <= updatedAt || info.UpdatedAt > updatedAtEnd) {
			// log.Debugf("get upload list skip because info is out of query date range")
			continue
		}
		// 0: all, 1. image, 2. document. 3. video, 4. music
		if !FileNameMatchType(fileType, info.FileName) {
			// log.Debugf("get upload list skip because info is mismatch type")
			continue
		}
		fileHashStr := info.FileHash
		if len(fileHashStr) == 0 {
			log.Warnf("task %s file hash is empty ", info.Id)
			continue
		}
		downloadedCount, _ := dsp.CountRecordByFileHash(fileHashStr)
		profit, _ := dsp.SumRecordsProfitByFileHash(fileHashStr)
		// init primary node map
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
		// rpc request
		if limit == 0 || uint64(len(files)) < limit {
			proveDetail, err := dsp.GetFileProveDetails(fileHashStr)
			nodesDetail := make([]NodeProveDetail, 0, info.CopyNum+1)
			fileHasUploaded := false
			if proveDetail != nil && err == nil {
				// log.Debugf("proveDetail %v, proveDetail.details %v", proveDetail, len(proveDetail.ProveDetails))
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
			} else {
				// log.Errorf("get prove %s detail failed err %s", fileHashStr, err)
			}
			if filterType == UploadFileFilterTypeDoing && len(primaryNodeM) == 0 {
				log.Debugf("get upload list skip because info no primary node")
				continue
			}
			if filterType == UploadFileFilterTypeDone && len(primaryNodeM) > 0 {
				log.Debugf("get upload list skip because primary node bigger than zero")
				continue
			}
			if proveDetail != nil && len(primaryNodeM) > 0 {
				unprovedNodeWallets := make([]chainCom.Address, 0)
				for addr, _ := range primaryNodeM {
					unprovedNodeWallets = append(unprovedNodeWallets, addr)
				}
				hostAddrs, err := dsp.GetNodeHostAddrListByWallets(unprovedNodeWallets)
				if err != nil {
					log.Debugf("get upload list skip because info wrong unprovedNodeWallets %v, err %v", unprovedNodeWallets, err)
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
				totalCount++
				continue
			}

			offsetCnt++
			sort.Sort(NodeProveDetails(nodesDetail))
			fileUrl := ""
			if fileHasUploaded && info.TaskState == store.TaskStateDone {
				fileUrl = info.Url
				totalCount++
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
				CreatedAt:     info.CreatedAt / 1000,
				UpdatedAt:     info.UpdatedAt / 1000,
				Profit:        profit,
				Privilege:     info.Privilege,
				ProveLevel:    info.ProveLevel,
				CurrentHeight: uint64(curBlockHeight),
				ExpiredHeight: info.ExpiredHeight,
				StoreType:     fs.FileStoreType(info.StoreType),
				RealFileSize:  info.RealFileSize,
				Nodes:         nodesDetail,
			}
			files = append(files, fr)
		} else {
			totalCount++
		}

	}
	log.Debugf("files num %d %d", len(files), totalCount)

	return files, totalCount, nil
}

type fileInfoResp struct {
	FileHash        string
	CreatedAt       uint64
	CopyNum         uint64
	Interval        uint64
	ProveLevel      uint64
	ProveTimes      uint64
	ExpiredHeight   uint64
	Privilege       uint64
	OwnerAddress    string
	Whitelist       []string
	ExpiredAt       uint64
	CurrentHeight   uint64
	Size            uint64
	RealFileSize    uint64
	StoreType       uint64
	BlocksRoot      string
	TotalBlockCount uint64
	Encrypt         bool
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
	tsk := dsp.GetUploadTaskInfoByHash(fileHashStr)
	encrypt := false
	if tsk != nil {
		encrypt = tsk.Encrypt
	}
	result := &fileInfoResp{
		FileHash:        string(info.FileHash),
		CopyNum:         info.CopyNum,
		Interval:        info.ProveInterval * config.BlockTime(),
		ProveLevel:      info.ProveLevel,
		ProveTimes:      info.ProveTimes,
		ExpiredHeight:   info.ExpiredHeight,
		Privilege:       info.Privilege,
		OwnerAddress:    info.FileOwner.ToBase58(),
		Whitelist:       []string{},
		ExpiredAt:       expiredAt,
		CurrentHeight:   uint64(now),
		Size:            info.FileBlockNum * info.FileBlockSize,
		RealFileSize:    info.RealFileSize,
		StoreType:       info.StorageType,
		BlocksRoot:      string(info.BlocksRoot),
		Encrypt:         encrypt,
		TotalBlockCount: info.FileBlockNum,
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
	[]*DownloadFilesInfo, int, *DspErr) {
	fileInfos := make([]*DownloadFilesInfo, 0)
	dsp := this.getDsp()
	if dsp == nil {
		return nil, 0, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	infos, _, err := dsp.AllDownloadFiles()
	if err != nil {
		return nil, 0, &DspErr{Code: DB_GET_FILEINFO_FAILED, Error: ErrMaps[DB_GET_FILEINFO_FAILED]}
	}
	totalCount := 0
	offsetCnt := uint64(0)
	isClient := dsp.IsClient()
	for _, info := range infos {
		if info == nil {
			continue
		}
		if isClient {
			exist := chainCom.FileExisted(info.FilePath)
			if !exist {
				log.Debugf("file not exist %s", info.FilePath)
				continue
			}
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
			totalCount++
			continue
		}
		totalCount++
		if limit > 0 && uint64(len(fileInfos)) >= limit {
			continue
		}
		downloadedCount, _ := dsp.CountRecordByFileHash(file)
		profit, _ := dsp.SumRecordsProfitByFileHash(file)
		lastSharedAt, _ := dsp.FindLastShareTime(file)
		owner := info.FileOwner
		privilege := uint64(info.Privilege)
		filePrefix := &dspPrefix.FilePrefix{}
		filePrefix.Deserialize(info.Prefix)
		fileInfos = append(fileInfos, &DownloadFilesInfo{
			Hash:          file,
			Name:          fileNameFromPath,
			OwnerAddress:  owner,
			Url:           url,
			Size:          uint64(info.TotalBlockCount * dspConsts.CHUNK_SIZE / 1024),
			DownloadCount: downloadedCount,
			DownloadAt:    info.CreatedAt / dspConsts.MILLISECOND_PER_SECOND,
			LastShareAt:   lastSharedAt,
			Profit:        profit,
			ProfitFormat:  utils.FormatUsdt(profit),
			Path:          info.FilePath,
			Privilege:     privilege,
			RealFileSize:  filePrefix.FileSize,
		})
	}
	return fileInfos, totalCount, nil
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
	if len(walletAddr) == 0 {
		walletAddr = this.getDspWalletAddress()
	}
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
	if countOpType == uint64(fs.UserSpaceRevoke) {
		currentHeight, err := dsp.GetCurrentBlockHeight()
		if err != nil {
			return "", &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
		}
		oldUserSpace, err := dsp.GetUserSpace(walletAddr)
		if err != nil {
			return "", &DspErr{Code: FS_GET_USER_SPACE_FAILED, Error: err}
		}
		fsSetting, err := dsp.GetFsSetting()
		if err != nil {
			return "", &DspErr{Code: FS_GET_SETTING_FAILED, Error: err}
		}
		if oldUserSpace.ExpireHeight-blockCount < uint64(currentHeight)+fsSetting.DefaultProvePeriod {
			return "", &DspErr{Code: FS_USER_SPACE_SECOND_INVALID, Error: ErrMaps[FS_USER_SPACE_SECOND_INVALID]}
		}
	}
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
			TransferType: common.TransferTypeIn,
		}, nil
	} else if cost.To.ToBase58() == dsp.WalletAddress() {
		return &UserspaceCostResp{
			Refund:       cost.Value,
			RefundFormat: utils.FormatUsdt(cost.Value),
			TransferType: common.TransferTypeOut,
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
	return len(dsp.DNS.GetPeerFromTracker(fileHashStr)), nil
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
			totalCount = fileSize * 1024 / dspConsts.CHUNK_SIZE
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

func (this *Endpoint) getTransferDetail(pType TransferType, info *dspTypes.ProgressInfo) *Transfer {
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
	for walletAddr, cnt := range info.Progress {
		sum += uint64(cnt.Progress)
		avgSpeed := cnt.AvgSpeed()
		if info.TaskState == store.TaskStateFailed || info.TaskState == store.TaskStatePause {
			avgSpeed = 0
		}
		pros := &NodeProgress{
			HostAddr: info.NodeHostAddrs[walletAddr],
			Speed:    avgSpeed,
		}
		if info.Type == store.TaskTypeUpload {
			pros.UploadSize = uint64(cnt.Progress) * dspConsts.CHUNK_SIZE / 1024
			pros.RealUploadSize = uint64(float64(cnt.Progress) / float64(info.Total) * float64(info.RealFileSize))
		} else if info.Type == store.TaskTypeDownload {
			pros.DownloadSize = uint64(cnt.Progress) * dspConsts.CHUNK_SIZE / 1024
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
		Encrypted:    info.Encrypt,
		Nodes:        nPros,
		CreatedAt:    info.CreatedAt,
		UpdatedAt:    info.UpdatedAt,
	}
	fee := info.FileSize * 1024
	if pInfo.RealFileSize > 0 {
		fee = info.RealFileSize * 1024
	}
	feeFormat := utils.FormatUsdt(fee)
	pInfo.Fee = fee
	pInfo.FeeFormat = feeFormat
	pInfo.IsUploadAction = (info.Type == store.TaskTypeUpload)
	pInfo.Progress = 0
	// log.Debugf("get transfer %s detail total %d sum %d ret %v err %s info.type %d", info.TaskKey, info.Total, sum, info.Result, info.ErrorMsg, info.Type)
	switch pType {
	case transferTypeUploading:
		if info.Total > 0 && sum >= uint64(info.Total) && info.Result != nil && len(info.ErrorMsg) == 0 {
			return nil
		}
		pInfo.UploadSize = sum * dspConsts.CHUNK_SIZE / 1024
		if len(pInfo.Nodes) > 0 && pInfo.FileSize > 0 {
			pInfo.Progress = (float64(pInfo.UploadSize) / float64(pInfo.FileSize))
		}
		if pInfo.Progress == 1 && info.Result != nil && info.ErrorCode == 0 {
			log.Warnf("info error msg of a success uploaded task %s, pInfo.UploadSize %d, file size %d",
				info.ErrorMsg, pInfo.UploadSize, pInfo.FileSize)
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
		pInfo.DownloadSize = sum * dspConsts.CHUNK_SIZE / 1024
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
			pInfo.UploadSize = sum * dspConsts.CHUNK_SIZE / 1024
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
			pInfo.DownloadSize = sum * dspConsts.CHUNK_SIZE / 1024
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
			pInfo.UploadSize = sum * dspConsts.CHUNK_SIZE / 1024
			if pInfo.Status != store.TaskStateDone && pInfo.FileSize > 0 && pInfo.UploadSize == pInfo.FileSize {
				log.Warnf("task:%s taskstate is %d, status:%d, but it has done",
					info.TaskId, info.TaskState, pInfo.Status)
				pInfo.Status = store.TaskStateDone
			}
			if len(pInfo.Nodes) > 0 && pInfo.FileSize > 0 {
				pInfo.Progress = (float64(pInfo.UploadSize) / float64(pInfo.FileSize))
			}
		} else if info.Type == store.TaskTypeDownload {
			pInfo.DownloadSize = sum * dspConsts.CHUNK_SIZE / 1024
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
	return filepath.Join(config.FsFileRootPath(), "/", fileName)
}

func blockHeightToTimestamp(curBlockHeight, endBlockHeight uint64) uint64 {
	now := uint64(time.Now().Unix())
	if endBlockHeight > curBlockHeight {
		return now + uint64(endBlockHeight-curBlockHeight)*config.BlockTime()
	}
	return now - uint64(curBlockHeight-endBlockHeight)*config.BlockTime()
}
