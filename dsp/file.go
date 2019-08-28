package dsp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dspCom "github.com/saveio/dsp-go-sdk/common"
	"github.com/saveio/dsp-go-sdk/task"
	dspUtils "github.com/saveio/dsp-go-sdk/utils"
	"github.com/saveio/edge/common"
	clicom "github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp/storage"
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
)

type NodeProgress struct {
	HostAddr     string
	UploadSize   uint64
	DownloadSize uint64
}

type Transfer struct {
	Id             string
	FileHash       string
	FileName       string
	Type           TransferType
	Status         task.TaskState
	DetailStatus   task.TaskProgressState
	CopyNum        uint64
	Path           string
	IsUploadAction bool
	UploadSize     uint64
	DownloadSize   uint64
	FileSize       uint64
	Nodes          []*NodeProgress
	Progress       float64
	CreatedAt      uint64
	UpdatedAt      uint64
	Result         interface{} `json:",omitempty"`
	ErrorCode      uint32
	ErrMsg         string `json:",omitempty"`
	StoreType      uint64
	Encrypted      bool
}

type TransferlistResp struct {
	IsTransfering bool
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

func (this *Endpoint) UploadFile(path, desc string, durationVal, intervalVal, privilegeVal, copyNumVal, storageTypeVal interface{},
	encryptPwd, url string, whitelist []string, share bool) (*fs.UploadOption, *DspErr) {
	f, err := os.Stat(path)
	if err != nil {
		return nil, &DspErr{Code: FS_UPLOAD_FILEPATH_ERROR, Error: fmt.Errorf("os stat file %s error: %s", path, err.Error())}
	}
	log.Debugf("path: %v, isDir: %t", path, f.IsDir())
	if f.IsDir() {
		return nil, &DspErr{Code: FS_UPLOAD_FILEPATH_ERROR, Error: fmt.Errorf("uploadFile error: %s is a directory", path)}
	}
	currentAccount := this.Dsp.CurrentAccount()
	fssetting, err := this.Dsp.Chain.Native.Fs.GetSetting()
	if err != nil {
		return nil, &DspErr{Code: FS_GET_SETTING_FAILED, Error: err}
	}
	currentHeight, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	interval, ok := intervalVal.(float64)
	if !ok || interval == 0 {
		interval = float64(fssetting.DefaultProvePeriod)
	}
	if uint64(interval) < fssetting.MinProveInterval {
		return nil, &DspErr{Code: FS_UPLOAD_INTERVAL_TOO_SMALL, Error: ErrMaps[FS_UPLOAD_INTERVAL_TOO_SMALL]}
	}
	storageType, _ := storageTypeVal.(float64)
	fileSizeInKB := f.Size() / 1024
	if fileSizeInKB == 0 {
		fileSizeInKB = 1
	}
	opt := &fs.UploadOption{
		FileDesc:      []byte(desc),
		ProveInterval: uint64(interval),
		StorageType:   uint64(storageType),
		FileSize:      uint64(fileSizeInKB),
	}
	if fs.FileStoreType(storageType) == fs.FileStoreTypeNormal {
		userspace, err := this.Dsp.Chain.Native.Fs.GetUserSpace(currentAccount.Address)
		if err != nil {
			return nil, &DspErr{Code: FS_GET_USER_SPACE_FAILED, Error: err}
		}
		log.Debugf("storageType %v, userspace.ExpireHeight %d, current: %d", storageType, userspace.ExpireHeight, currentHeight)
		if userspace.ExpireHeight <= uint64(currentHeight) {
			return nil, &DspErr{Code: DSP_USER_SPACE_EXPIRED, Error: ErrMaps[DSP_USER_SPACE_EXPIRED]}
		}
		opt.ExpiredHeight = userspace.ExpireHeight
	} else {
		duration, _ := durationVal.(float64)
		opt.ExpiredHeight = uint64(currentHeight) + uint64(duration)
	}
	log.Debugf("opt.ExpiredHeight :%d, minInterval :%d, current: %d", opt.ExpiredHeight, fssetting.MinProveInterval, currentHeight)
	if opt.ExpiredHeight < fssetting.MinProveInterval+uint64(currentHeight) {
		return nil, &DspErr{Code: DSP_CUSTOM_EXPIRED_NOT_ENOUGH, Error: ErrMaps[DSP_CUSTOM_EXPIRED_NOT_ENOUGH]}
	}
	privilege, ok := privilegeVal.(float64)
	if !ok {
		privilege = fs.PUBLIC
	}
	opt.Privilege = uint64(privilege)
	copyNum, ok := copyNumVal.(float64)
	if !ok {
		copyNum = float64(fssetting.DefaultCopyNum)
	}
	opt.CopyNum = uint64(copyNum)
	if len(url) == 0 {
		// random
		b := make([]byte, clicom.DSP_URL_RAMDOM_NAME_LEN/2)
		_, err := rand.Read(b)
		if err != nil {
			return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
		url = this.Dsp.Chain.Native.Dns.GetCustomUrlHeader() + hex.EncodeToString(b)
	}
	find, err := this.Dsp.Chain.Native.Dns.QueryUrl(url, this.Dsp.CurrentAccount().Address)
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
	log.Debugf("whitelist :%v, len: %d %d", whitelist, len(whitelistObj.List), cap(whitelistObj.List))
	for i, whitelistAddr := range whitelist {
		addr, err := chainCom.AddressFromBase58(whitelistAddr)
		if err != nil {
			return nil, &DspErr{Code: INVALID_WALLET_ADDRESS, Error: err}
		}
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
	taskExist, err := this.Dsp.UploadTaskExist(path)
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	if taskExist {
		return nil, &DspErr{Code: DSP_UPLOAD_FILE_EXIST, Error: ErrMaps[DSP_UPLOAD_FILE_EXIST]}
	}
	go func() {
		log.Debugf("upload file path %s, this.Dsp: %t", path, this.Dsp == nil)
		ret, err := this.Dsp.UploadFile("", path, opt)
		if err != nil {
			log.Errorf("upload failed err %s", err)
			return
		} else {
			log.Infof("upload file success: %v", ret)
		}
	}()
	return opt, nil
}

func (this *Endpoint) PauseUploadFile(taskIds []string) *FileTaskResp {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}

		exist := this.Dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := this.Dsp.PauseUpload(id)
		if err != nil {
			taskResp.Code = DSP_PAUSE_UPLOAD_FAIELD
			taskResp.Error = err.Error()
		}
		state, err := this.Dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_RESUME_UPLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("pause upload failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp
}

func (this *Endpoint) ResumeUploadFile(taskIds []string) *FileTaskResp {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}
		exist := this.Dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := this.Dsp.ResumeUpload(id)
		log.Debugf("resume upload err %v", err)
		if err != nil {
			taskResp.Code = DSP_RESUME_UPLOAD_FAIELD
			taskResp.Error = err.Error()
		}
		state, err := this.Dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_RESUME_UPLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("resume upload failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp
}

func (this *Endpoint) RetryUploadFile(taskIds []string) *FileTaskResp {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}
		exist := this.Dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := this.Dsp.RetryUpload(id)
		if err != nil {
			taskResp.Code = DSP_RETRY_UPLOAD_FAIELD
			taskResp.Error = err.Error()
		}
		state, err := this.Dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_RETRY_UPLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("retry upload failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp
}

func (this *Endpoint) CancelUploadFile(taskIds []string) *FileTaskResp {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}

	// send delete files tx
	fileHashes := make([]string, 0, len(taskIds))
	for _, id := range taskIds {
		fileHash := this.Dsp.GetTaskFileHash(id)
		fileHashes = append(fileHashes, fileHash)
	}
	_, _, deleteTxErr := this.Dsp.DeleteUploadFilesFromChain(fileHashes)

	for _, id := range taskIds {
		taskResp := &FileTask{
			Id:       id,
			FileName: this.Dsp.GetTaskFileName(id),
			State:    int(task.TaskStateCancel),
		}
		if deleteTxErr != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = deleteTxErr.Error.Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		exist := this.Dsp.IsTaskExist(id)
		if !exist {
			err := this.DeleteProgress([]string{id})
			if err != nil {
				taskResp.Code = DSP_CANCEL_TASK_FAILED
				taskResp.Error = err.Error()
			}
			log.Debugf("cancel upload file, id :%s, resp :%v", id, taskResp)
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		deleteResp, err := this.Dsp.CancelUpload(id)
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		taskResp.Result = deleteResp
		err = this.DeleteProgress([]string{id})
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
		}
		log.Debugf("cancel upload file, id :%s, resp :%v", id, taskResp)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp
}

func (this *Endpoint) DeleteUploadFile(fileHash string) (*DeleteFileResp, *DspErr) {
	fi, err := this.Dsp.Chain.Native.Fs.GetFileInfo(fileHash)
	if fi != nil && err == nil && fi.FileOwner.ToBase58() == this.Dsp.WalletAddress() {
		result, err := this.Dsp.DeleteUploadedFiles([]string{fileHash})
		if err != nil {
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
	return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: err}
}

func (this *Endpoint) DeleteDownloadFile(fileHash string) (*DeleteFileResp, *DspErr) {
	err := this.Dsp.DeleteDownloadedFile(fileHash)
	if err != nil {
		return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: err}
	}
	return &DeleteFileResp{IsUploaded: false}, nil
}

func (this *Endpoint) DeleteUploadFiles(fileHashs []string) ([]*DeleteFileResp, *DspErr) {
	result, err := this.Dsp.DeleteUploadedFiles(fileHashs)
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

func (this *Endpoint) GetFsConfig() (*FsContractSettingResp, *DspErr) {
	set, err := this.Dsp.Chain.Native.Fs.GetSetting()
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}

	return &FsContractSettingResp{
		DefaultCopyNum:     set.DefaultCopyNum,
		DefaultProvePeriod: set.DefaultProvePeriod,
		MinProveInterval:   set.MinProveInterval,
		MinVolume:          set.MinVolume,
	}, nil
}

func (this *Endpoint) IsChannelProcessBlocks() (bool, *DspErr) {
	if this.Dsp == nil || this.Dsp.Channel == nil {
		return false, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	filterBlockHeight := this.Dsp.Channel.GetCurrentFilterBlockHeight()
	now, getHeightErr := this.Dsp.Chain.GetCurrentBlockHeight()
	log.Debugf("IsChannelProcessBlocks filterBlockHeight: %d, now :%d", filterBlockHeight, now)
	if getHeightErr != nil {
		return false, &DspErr{Code: INTERNAL_ERROR, Error: ErrMaps[INTERNAL_ERROR]}
	}
	if filterBlockHeight+common.MAX_SYNC_HEIGHT_OFFSET <= now {
		this.SetFilterBlockRange()
		return true, nil
	}
	return false, nil
}

func (this *Endpoint) DownloadFile(fileHash, url, link, password string, max uint64, setFileName bool) *DspErr {
	// if balance of current channel is not enouth, reject
	if this.Dsp.DNS == nil || this.Dsp.DNS.DNSNode == nil || this.Dsp.DNS.DNSNode.WalletAddr == "" {
		return &DspErr{Code: DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST, Error: ErrMaps[DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST]}
	}

	fileInfo, err := this.GetDownloadFileInfo(url)
	if err != nil {
		return err
	}

	canDownload := false
	//[NOTE] when this.QueryChannel works, replace this.GetAllChannels logic
	all, getChannelErr := this.Dsp.Channel.AllChannels()
	if getChannelErr != nil {
		return &DspErr{Code: INTERNAL_ERROR, Error: getChannelErr}
	}
	if all == nil || len(all.Channels) == 0 {
		return &DspErr{Code: DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH, Error: ErrMaps[DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH]}
	}

	for _, ch := range all.Channels {
		if ch.Address == this.Dsp.DNS.DNSNode.WalletAddr && ch.Balance >= fileInfo.Fee {
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

	if len(fileHash) > 0 {
		go func() {
			err := this.Dsp.DownloadFileByHash(fileHash, dspCom.ASSET_USDT, true, password, false, setFileName, int(max))
			if err != nil {
				log.Errorf("Downloadfile from url failed %s", err)
			}
		}()
		return nil
	}
	if len(url) > 0 {
		hash := this.Dsp.GetFileHashFromUrl(url)
		if len(hash) == 0 {
			return &DspErr{Code: INTERNAL_ERROR, Error: fmt.Errorf("file hash not found for url %s", url)}
		}
		go func() {
			hash := this.Dsp.GetFileHashFromUrl(url)
			this.SetUrlForHash(hash, url)
			err := this.Dsp.DownloadFileByUrl(url, dspCom.ASSET_USDT, true, password, false, setFileName, int(max))
			if err != nil {
				log.Errorf("Downloadfile from url failed %s", err)
			}
		}()
		return nil
	}
	if len(link) > 0 {
		hash := dspUtils.GetFileHashFromLink(link)
		if len(hash) == 0 {
			return &DspErr{Code: INTERNAL_ERROR, Error: fmt.Errorf("file hash not found for url %s", hash)}
		}
		go func() {
			err := this.Dsp.DownloadFileByLink(link, dspCom.ASSET_USDT, true, password, false, setFileName, int(max))
			if err != nil {
				log.Errorf("Downloadfile from url failed %s", err)
			}
		}()
		return nil
	}
	return nil
}

func (this *Endpoint) PauseDownloadFile(taskIds []string) *FileTaskResp {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}

		exist := this.Dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := this.Dsp.PauseDownload(id)
		if err != nil {
			taskResp.Code = DSP_PAUSE_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
		}
		state, err := this.Dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_PAUSE_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("pause download failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp
}

func (this *Endpoint) ResumeDownloadFile(taskIds []string) *FileTaskResp {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}
		exist := this.Dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := this.Dsp.ResumeDownload(id)
		if err != nil {
			taskResp.Code = DSP_RESUME_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
		}
		state, err := this.Dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_RESUME_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("resume download failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp
}

func (this *Endpoint) RetryDownloadFile(taskIds []string) *FileTaskResp {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id: id,
		}
		exist := this.Dsp.IsTaskExist(id)
		if !exist {
			taskResp.Code = DSP_TASK_NOT_EXIST
			taskResp.Error = ErrMaps[DSP_TASK_NOT_EXIST].Error()
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := this.Dsp.RetryDownload(id)
		if err != nil {
			taskResp.Code = DSP_RETRY_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("retry download failed %s", err)
		}
		state, err := this.Dsp.GetTaskState(id)
		if err != nil {
			taskResp.Code = DSP_RETRY_DOWNLOAD_FAIELD
			taskResp.Error = err.Error()
			log.Errorf("retry download failed %s", err)
		}
		taskResp.State = int(state)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp
}

func (this *Endpoint) CancelDownloadFile(taskIds []string) *FileTaskResp {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id:       id,
			FileName: this.Dsp.GetTaskFileName(id),
		}
		exist := this.Dsp.IsTaskExist(id)
		if !exist {
			err := this.DeleteProgress([]string{id})
			if err != nil {
				taskResp.Code = DSP_CANCEL_TASK_FAILED
				taskResp.Error = err.Error()
			}
			resp.Tasks = append(resp.Tasks, taskResp)
			continue
		}
		err := this.Dsp.CancelDownload(id)
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
		}
		err = this.DeleteProgress([]string{id})
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
		}
		taskResp.State = int(task.TaskStateCancel)
		log.Debugf("cancel download file, id :%s, resp :%v", id, taskResp)
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp
}

func (this *Endpoint) RegisterProgressCh() {
	if this.Dsp == nil {
		log.Errorf("this.Dsp is nil")
		return
	}
	this.Dsp.RegProgressChannel()
	for {
		select {
		case v, ok := <-this.Dsp.ProgressChannel():
			// TODO: replace with list
			err := this.AddProgress(v)
			if err != nil {
				log.Errorf("add progress err %s", err)
			}
			infoRet, ok := v.Result.(*dspCom.UploadResult)
			fmt.Printf("inforet %v, ok %t\n", infoRet, ok)
			if ok && infoRet != nil {
				this.SetUrlForHash(v.FileHash, infoRet.Url)
			}
			// log.Debugf("progress store file %s, %v, ok %t", v.TaskId, v, ok)
			for node, cnt := range v.Count {
				log.Infof("progress type:%d file:%s, hash:%s, total:%d, peer:%s, uploaded:%d, progress:%f", v.Type, v.FileName, v.FileHash, v.Total, node, cnt, float64(cnt)/float64(v.Total))
			}
		case <-this.closeCh:
			this.Dsp.CloseProgressChannel()
			return
		}
	}
}

func (this *Endpoint) DeleteTransferRecord(taskIds []string) *FileTaskResp {
	resp := &FileTaskResp{
		Tasks: make([]*FileTask, 0, len(taskIds)),
	}
	for _, id := range taskIds {
		taskResp := &FileTask{
			Id:    id,
			State: int(task.TaskStateCancel),
		}
		err := this.DeleteProgress([]string{id})
		if err != nil {
			taskResp.Code = DSP_CANCEL_TASK_FAILED
			taskResp.Error = err.Error()
		}
		resp.Tasks = append(resp.Tasks, taskResp)
	}
	return resp
}

// GetTransferList. get transfer progress list
func (this *Endpoint) GetTransferList(pType TransferType, offset, limit uint64) *TransferlistResp {
	infos := make([]*Transfer, 0)
	off := uint64(0)
	resp := &TransferlistResp{
		IsTransfering: false,
		Transfers:     []*Transfer{},
	}
	allTasksKey, err := this.GetAllProgressKeys()
	if err != nil {
		return resp
	}
	for idx, key := range allTasksKey {
		info, err := this.GetProgressByKey(key)
		if err != nil {
			log.Warnf("get progress failed %d for %s info %v err %s", idx, key, info, err)
			continue
		}
		if len(info.TaskId) == 0 {
			continue
		}
		if pType == transferTypeUploading && info.Type != task.TaskTypeUpload {
			continue
		}
		if pType == transferTypeDownloading && info.Type != task.TaskTypeDownload {
			continue
		}
		pInfo := this.getTransferDetail(pType, info)
		if pInfo == nil {
			continue
		}
		if !resp.IsTransfering {
			resp.IsTransfering = (pType == transferTypeUploading || pType == transferTypeDownloading) && (pInfo.Status != task.TaskStateFailed && pInfo.Status != task.TaskStateDone)
		}

		if off < offset {
			off++
			continue
		}
		infos = append(infos, pInfo)
		off++
		if limit > 0 && uint64(len(infos)) >= limit {
			break
		}
	}
	resp.Transfers = infos
	return resp
}

// GetTransferList. get transfer progress list
func (this *Endpoint) GetTransferDetail(pType TransferType, id string) (*Transfer, *DspErr) {
	if len(id) == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	resp := &Transfer{}
	info, err := this.GetProgress(this.Dsp.WalletAddress(), id)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	pInfo := this.getTransferDetail(pType, info)
	if pInfo == nil {
		return resp, &DspErr{Code: INTERNAL_ERROR, Error: ErrMaps[INTERNAL_ERROR]}
	}
	return pInfo, nil
}

func (this *Endpoint) CalculateUploadFee(filePath string, durationVal, intervalVal, timesVal, copynumVal, whitelistVal, storeType interface{}) (*CalculateResp, *DspErr) {
	currentAccount := this.Dsp.CurrentAccount()
	fssetting, err := this.Dsp.Chain.Native.Fs.GetSetting()
	if err != nil {
		return nil, &DspErr{Code: FS_GET_SETTING_FAILED, Error: err}
	}
	interval, err := OptionStrToFloat64(intervalVal, float64(fssetting.DefaultProvePeriod))
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
	copyNum, err := OptionStrToFloat64(copynumVal, float64(fssetting.DefaultCopyNum))
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

	currentHeight, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	if fs.FileStoreType(sType) == fs.FileStoreTypeNormal {
		userspace, err := this.Dsp.Chain.Native.Fs.GetUserSpace(currentAccount.Address)
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
		log.Debugf("userspace.ExpireHeight %d, current %d, interval:%v", userspace.ExpireHeight, currentHeight, interval)
		fee, err := this.Dsp.CalculateUploadFee(opt)
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
	opt.ExpiredHeight = uint64(currentHeight) + uint64(duration)
	log.Debugf("opt :%v\n", opt)
	fee, err := this.Dsp.CalculateUploadFee(opt)
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
	info := &DownloadFileInfo{}
	var fileLink string
	if strings.HasPrefix(url, dspCom.FILE_URL_CUSTOM_HEADER) {
		fileLink = this.Dsp.GetLinkFromUrl(url)
	} else if strings.HasPrefix(url, dspCom.FILE_LINK_PREFIX) {
		fileLink = url
	} else if strings.HasPrefix(url, dspCom.PROTO_NODE_PREFIX) || strings.HasPrefix(url, dspCom.RAW_NODE_PREFIX) {
		// TODO support get download file info from hash
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	if len(fileLink) == 0 {
		return nil, &DspErr{Code: DSP_GET_FILE_LINK_FAILED, Error: ErrMaps[DSP_GET_FILE_LINK_FAILED]}
	}
	values := this.Dsp.GetLinkValues(fileLink)
	if values == nil {
		return nil, &DspErr{Code: DSP_GET_FILE_LINK_FAILED, Error: ErrMaps[DSP_GET_FILE_LINK_FAILED]}
	}
	info.Hash = values[dspCom.FILE_LINK_HASH_KEY]
	info.Name = values[dspCom.FILE_LINK_NAME_KEY]
	blockNumStr := values[dspCom.FILE_LINK_BLOCKNUM_KEY]
	blockNum, err := strconv.ParseUint(blockNumStr, 10, 64)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	info.Size = blockNum * dspCom.CHUNK_SIZE / 1024
	extParts := strings.Split(info.Name, ".")
	if len(extParts) > 1 {
		info.Ext = extParts[len(extParts)-1]
	}
	info.Fee = blockNum * dspCom.CHUNK_SIZE * common.DSP_DOWNLOAD_UNIT_PRICE
	info.FeeFormat = utils.FormatUsdt(info.Fee)
	info.Path = this.getDownloadFilePath(info.Name)
	info.DownloadDir = this.getDownloadFilePath("")
	return info, nil
}

func (this *Endpoint) EncryptFile(path, password string) *DspErr {
	err := this.Dsp.Fs.AESEncryptFile(path, password, path+".temp")
	if err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	err = os.Rename(path+".temp", path)
	if err != nil {
		return &DspErr{Code: DSP_ENCRYPTED_FILE_FAILED, Error: err}
	}
	return nil
}

func (this *Endpoint) DecryptFile(path, password string) *DspErr {
	sourceFile, err := os.Open(path)
	if err != nil {
		return &DspErr{Code: DSP_FILE_NOT_EXISTS, Error: err}
	}
	defer sourceFile.Close()
	prefix := make([]byte, dspUtils.PREFIX_LEN)
	_, err = sourceFile.Read(prefix)
	if err != nil {
		return &DspErr{Code: DSP_DECRYPTED_FILE_FAILED, Error: err}
	}
	filePrefix := &dspUtils.FilePrefix{}
	filePrefix.Deserialize([]byte(prefix))
	if !dspUtils.VerifyEncryptPassword(password, filePrefix.EncryptSalt, filePrefix.EncryptHash) {
		return &DspErr{Code: DSP_FILE_DECRYPTED_WRONG_PWD, Error: ErrMaps[DSP_FILE_DECRYPTED_WRONG_PWD]}
	}
	err = this.Dsp.Fs.AESDecryptFile(path, string(prefix), password, dspUtils.GetDecryptedFilePath(path))
	if err != nil {
		return &DspErr{Code: DSP_DECRYPTED_FILE_FAILED, Error: err}
	}
	// err = os.Rename(path+".temp", path)
	// if err != nil {
	// 	return &DspErr{Code: DSP_DECRYPTED_FILE_FAILED, Error: err}
	// }
	return nil
}

func (this *Endpoint) GetFileRevene() (uint64, *DspErr) {
	if this.sqliteDB == nil {
		return 0, &DspErr{Code: NO_DB, Error: ErrMaps[NO_DB]}
	}
	sum, err := this.sqliteDB.SumRecordsProfit()
	if err != nil {
		return 0, &DspErr{Code: DB_SUM_SHARE_PROFIT_FAILED, Error: err}
	}
	return uint64(sum), nil
}

func (this *Endpoint) GetFileShareIncome(start, end, offset, limit uint64) (*FileShareIncomeResp, *DspErr) {
	resp := &FileShareIncomeResp{}
	records, err := this.sqliteDB.FineShareRecordsByCreatedAt(int64(start), int64(end), int64(offset), int64(limit))
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
			SharedAt:     uint64(record.CreatedAt.Unix()),
		})
	}
	resp.TotalIncomeFormat = utils.FormatUsdt(resp.TotalIncome)
	return resp, nil
}

func (this *Endpoint) RegisterShareNotificationCh() {
	if this.Dsp == nil {
		log.Errorf("this.Dsp is nil")
		return
	}
	this.Dsp.RegShareNotificationChannel()
	for {
		select {
		case v, ok := <-this.Dsp.ShareNotificationChannel():
			if !ok {
				break
			}
			log.Debugf("share notification taskkey=%s, filehash=%s, walletaddr=%s, state=%d, amount=%d", v.TaskKey, v.FileHash, v.ToWalletAddr, v.State, v.PaymentAmount)
			switch v.State {
			case task.ShareStateBegin:
				// TODO: repace id with a real id, not random timestamp
				id := fmt.Sprintf("%s-%d", v.TaskKey, time.Now().Unix())
				_, err := this.sqliteDB.InsertShareRecord(id, v.FileHash, v.FileName, v.FileOwner, v.ToWalletAddr, v.PaymentAmount)
				log.Debugf("insert share record : %s, %v", id, v)
				if err != nil {
					log.Errorf("insert new share_record failed %s, err %s", id, err)
				}
			case task.ShareStateReceivedPaying, task.ShareStateEnd:
				_, err := this.sqliteDB.IncreaseShareRecordProfit("", v.TaskKey, v.PaymentAmount)
				log.Debugf("insert share record2 : %s, %v", v)
				if err != nil {
					log.Errorf("increase share_record profit failed %s, err %s", v.TaskKey, err)
				}
			default:
				log.Warn("unknown state type")
			}

		case <-this.closeCh:
			this.Dsp.CloseShareNotificationChannel()
			return
		}
	}
}

func (this *Endpoint) GetUploadFiles(fileType DspFileListType, offset, limit uint64) ([]*FileResp, *DspErr) {
	fileList, err := this.Dsp.Chain.Native.Fs.GetFileList(this.Dsp.Account.Address)
	if err != nil {
		return nil, &DspErr{Code: FS_GET_FILE_LIST_FAILED, Error: err}
	}

	now, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}

	files := make([]*FileResp, 0)
	offsetCnt := uint64(0)
	for _, hash := range fileList.List {
		fi, err := this.Dsp.Chain.Native.Fs.GetFileInfo(string(hash.Hash))
		if err != nil {
			log.Errorf("get file info err %s", err)
			continue
		}
		// 0: all, 1. image, 2. document. 3. video, 4. music
		fileName := strings.ToLower(string(fi.FileDesc))
		if !FileNameMatchType(fileType, fileName) {
			continue
		}
		if offsetCnt < offset {
			offsetCnt++
			continue
		}
		offsetCnt++

		expired := fi.ExpiredHeight
		expiredAt := uint64(time.Now().Unix()) + (expired - uint64(now))
		updatedAt := uint64(time.Now().Unix())
		if fi.BlockHeight > uint64(now) {
			updatedAt -= (fi.BlockHeight - uint64(now))
		} else {
			updatedAt -= uint64(now) - fi.BlockHeight
		}
		url, _ := this.GetUrlFromHash(string(hash.Hash))
		downloadedCount, _ := this.sqliteDB.CountRecordByFileHash(string(hash.Hash))
		profit, _ := this.sqliteDB.SumRecordsProfitByFileHash(string(hash.Hash))

		fr := &FileResp{
			Hash:          string(hash.Hash),
			Name:          string(fi.FileDesc),
			Url:           url,
			Size:          fi.FileBlockNum * fi.FileBlockSize,
			DownloadCount: downloadedCount,
			ExpiredAt:     expiredAt,
			// TODO fix by db
			UpdatedAt:     updatedAt,
			Profit:        profit,
			Privilege:     fi.Privilege,
			CurrentHeight: uint64(now),
			ExpiredHeight: expired,
			StoreType:     fs.FileStoreType(fi.StorageType),
			RealFileSize:  fi.RealFileSize,
		}
		files = append(files, fr)
		if limit > 0 && uint64(len(files)) >= limit {
			break
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
}

func (this *Endpoint) GetFileInfo(fileHashStr string) (*fileInfoResp, *DspErr) {
	if len(fileHashStr) == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	info, err := this.Dsp.Chain.Native.Fs.GetFileInfo(fileHashStr)
	if err != nil {
		return nil, &DspErr{Code: DSP_FILE_INFO_NOT_FOUND, Error: ErrMaps[DSP_FILE_INFO_NOT_FOUND]}
	}

	now, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	expiredAt := uint64(time.Now().Unix()) + (info.ExpiredHeight - uint64(now))
	result := &fileInfoResp{
		FileHash:      string(info.FileHash),
		CopyNum:       info.CopyNum,
		Interval:      info.ProveInterval,
		ProveTimes:    info.ProveTimes,
		ExpiredHeight: info.ExpiredHeight,
		Privilege:     info.Privilege,
		OwnerAddress:  info.FileOwner.ToBase58(),
		Whitelist:     []string{},
		ExpiredAt:     expiredAt,
		CurrentHeight: uint64(now),
		Size:          info.FileBlockNum * info.FileBlockSize,
		RealFileSize:  info.RealFileSize,
	}
	block, _ := this.Dsp.Chain.GetBlockByHeight(uint32(info.BlockHeight))
	if block == nil {
		result.CreatedAt = uint64(time.Now().Unix())
	} else {
		result.CreatedAt = uint64(block.Header.Timestamp)
	}
	if info.Privilege != fs.WHITELIST {
		return result, nil
	}

	whitelist, err := this.Dsp.Chain.Native.Fs.GetWhiteList(fileHashStr)
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

func (this *Endpoint) GetDownloadFiles(fileType DspFileListType, offset, limit uint64) ([]*DownloadFilesInfo, *DspErr) {
	fileInfos := make([]*DownloadFilesInfo, 0)
	if this.Dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	infos, _, err := this.Dsp.AllDownloadFiles()
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
		url, err := this.GetUrlFromHash(file)
		if err != nil {
			log.Errorf("get url from hash %s, err %s", file, err)
		}
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
		downloadedCount, _ := this.sqliteDB.CountRecordByFileHash(file)
		profit, _ := this.sqliteDB.SumRecordsProfitByFileHash(file)
		lastSharedAt, _ := this.sqliteDB.FindLastShareTime(file)
		fileInfo, _ := this.Dsp.Chain.Native.Fs.GetFileInfo(file)
		owner := ""
		privilege := uint64(fs.PUBLIC)
		if fileInfo != nil {
			owner = fileInfo.FileOwner.ToBase58()
			privilege = fileInfo.Privilege
		}
		filePrefix := &dspUtils.FilePrefix{}
		filePrefix.Deserialize(info.Prefix)
		fileInfos = append(fileInfos, &DownloadFilesInfo{
			Hash:          file,
			Name:          fileNameFromPath,
			OwnerAddress:  owner,
			Url:           url,
			Size:          info.TotalBlockCount * dspCom.CHUNK_SIZE / 1024,
			DownloadCount: downloadedCount,
			DownloadAt:    info.CreatedAt,
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
	txHash, err := this.Dsp.Chain.Native.Fs.WhiteListOp(fileHash, op, li)
	if err != nil {
		return "", &DspErr{Code: DSP_WHITELIST_OP_FAILED, Error: err}
	}
	return hex.EncodeToString(chainCom.ToArrayReverse(txHash)), nil
}

func (this *Endpoint) GetWhitelist(fileHash string) ([]*WhiteListRule, *DspErr) {
	if len(fileHash) == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	list, err := this.Dsp.Chain.Native.Fs.GetWhiteList(fileHash)
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

func (this *Endpoint) SetUserSpace(walletAddr string, size, sizeOpType, blockCount, countOpType uint64) (string, *DspErr) {
	if sizeOpType == uint64(fs.UserSpaceNone) && countOpType == uint64(fs.UserSpaceNone) {
		return "", &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	if sizeOpType == uint64(fs.UserSpaceNone) {
		size = 0
	}
	if countOpType == uint64(fs.UserSpaceNone) {
		blockCount = 0
	}
	tx, err := this.Dsp.UpdateUserSpace(walletAddr, size, sizeOpType, blockCount, countOpType)
	if err != nil {
		return tx, ParseContractError(err)
	}
	txReHash, err := hex.DecodeString(tx)
	if err != nil {
		log.Errorf("decode tx string failed %s", err)
	}
	_, err = this.Dsp.Chain.PollForTxConfirmed(time.Duration(common.POLL_TX_COMFIRMED_TIMEOUT)*time.Second, chainCom.ToArrayReverse(txReHash))
	if err != nil {
		return "", &DspErr{Code: CHAIN_WAIT_TX_COMFIRMED_TIMEOUT, Error: err}
	}
	event, err := this.Dsp.Chain.GetSmartContractEvent(tx)
	if err != nil || event == nil {
		log.Debugf("get event err %s, event :%v", err, event)
		_, err := this.sqliteDB.InsertUserspaceRecord(tx, walletAddr, size, storage.UserspaceOperation(sizeOpType), blockCount, storage.UserspaceOperation(countOpType), 0, storage.TransferTypeNone)
		if err != nil {
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
		if len(states) != 4 || states[0] != "transfer" {
			continue
		}
		from := states[1].(string)
		to := states[2].(string)
		if to != chainSdkFs.FS_CONTRACT_ADDRESS.ToBase58() && to != this.Dsp.WalletAddress() {
			continue
		}
		hasTransfer = true
		amount := states[3].(uint64)
		transferType := storage.TransferTypeIn
		if to == walletAddr {
			transferType = storage.TransferTypeOut
		}
		_, err := this.sqliteDB.InsertUserspaceRecord(tx, walletAddr, size, storage.UserspaceOperation(sizeOpType), blockCount, storage.UserspaceOperation(countOpType), amount, transferType)
		if err != nil {
			log.Errorf("insert userspace record err %s", err)
		}
		log.Debugf("from %s to %s amount %d", from, to, amount)
	}
	if len(event.Notify) == 0 || !hasTransfer {
		_, err := this.sqliteDB.InsertUserspaceRecord(tx, walletAddr, size, storage.UserspaceOperation(sizeOpType), blockCount, storage.UserspaceOperation(countOpType), 0, storage.TransferTypeNone)
		if err != nil {
			log.Errorf("insert userspace record err %s", err)
			return "", &DspErr{Code: DB_ADD_USER_SPACE_RECORD_FAILED, Error: err}
		}
		return tx, nil
	}
	return tx, nil
}

func (this *Endpoint) GetUserSpaceCost(walletAddr string, size, sizeOpType, blockCount, countOpType uint64) (*UserspaceCostResp, *DspErr) {
	if sizeOpType == uint64(fs.UserSpaceNone) && countOpType == uint64(fs.UserSpaceNone) {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	if sizeOpType == uint64(fs.UserSpaceNone) {
		size = 0
	}
	if countOpType == uint64(fs.UserSpaceNone) {
		blockCount = 0
	}
	cost, err := this.Dsp.GetUpdateUserSpaceCost(walletAddr, size, sizeOpType, blockCount, countOpType)
	if err != nil {
		return nil, ParseContractError(err)
	}
	if cost.From.ToBase58() == this.Dsp.Account.Address.ToBase58() {
		return &UserspaceCostResp{
			Fee:          cost.Value,
			FeeFormat:    utils.FormatUsdt(cost.Value),
			TransferType: storage.TransferTypeIn,
		}, nil
	} else if cost.To.ToBase58() == this.Dsp.Account.Address.ToBase58() {
		return &UserspaceCostResp{
			Refund:       cost.Value,
			RefundFormat: utils.FormatUsdt(cost.Value),
			TransferType: storage.TransferTypeOut,
		}, nil
	}
	return nil, &DspErr{Code: INTERNAL_ERROR, Error: ErrMaps[INTERNAL_ERROR]}
}

func (this *Endpoint) GetUserSpace(addr string) (*Userspace, *DspErr) {
	space, err := this.Dsp.GetUserSpace(addr)
	if err != nil || space == nil {
		return &Userspace{
			Used:      0,
			Remain:    0,
			ExpiredAt: 0,
			Balance:   0,
		}, nil
	}
	currentHeight, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	expiredAt := uint64(0)
	updateHeight := space.UpdateHeight
	now := uint64(time.Now().Unix())
	log.Debugf("space.ExpireHeight %v\n", space.ExpireHeight)
	if space.ExpireHeight > uint64(currentHeight) {
		blk, err := this.Dsp.Chain.GetBlockByHeight(uint32(updateHeight))
		if err != nil {
			return nil, &DspErr{Code: CHAIN_GET_BLK_BY_HEIGHT_FAILED, Error: err}
		}
		expiredAt = uint64(blk.Header.Timestamp) + (space.ExpireHeight - uint64(updateHeight))
		log.Debugf("expiredAt %d height %d, expiredheight %d updatedheight %d", expiredAt, blk.Header.Timestamp, space.ExpireHeight, updateHeight)
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
	if expiredAt <= now {
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
		Used:          space.Used,
		Remain:        space.Remain,
		ExpiredAt:     expiredAt,
		Balance:       space.Balance,
		CurrentHeight: uint64(currentHeight),
		ExpiredHeight: space.ExpireHeight,
	}, nil
}

func (this *Endpoint) GetUserspaceRecords(walletAddr string, offset, limit uint64) ([]*UserspaceRecordResp, *DspErr) {
	records, err := this.sqliteDB.SelectUserspaceRecordByWalletAddr(walletAddr, offset, limit)
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
		if record.TransferType == storage.TransferTypeOut {
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
	info, err := this.Dsp.Chain.Native.Fs.GetNodeList()
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	m := make(map[string]interface{})
	m["Count"] = info.NodeNum
	return m, nil
}

func (this *Endpoint) GetProveDetail(fileHashStr string) (interface{}, *DspErr) {
	details, err := this.Dsp.Chain.Native.Fs.GetFileProveDetails(fileHashStr)
	if err != nil {
		return nil, &DspErr{Code: CONTRACT_ERROR, Error: err}
	}
	return details, nil
}

func (this *Endpoint) getTransferDetail(pType TransferType, info *task.ProgressInfo) *Transfer {
	if info.TaskState != task.TaskStateDone && info.TaskState != task.TaskStateFailed {
		// update state by task cache
		state, err := this.Dsp.GetTaskState(info.TaskId)
		if err == nil {
			info.TaskState = state
		}
	}
	sum := uint64(0)
	npros := make([]*NodeProgress, 0)
	for haddr, cnt := range info.Count {
		sum += cnt
		pros := &NodeProgress{
			HostAddr: haddr,
		}
		if info.Type == task.TaskTypeUpload {
			pros.UploadSize = cnt * dspCom.CHUNK_SIZE / 1024
		} else if info.Type == task.TaskTypeDownload {
			pros.DownloadSize = cnt * dspCom.CHUNK_SIZE / 1024
		}
		npros = append(npros, pros)
	}
	pInfo := &Transfer{
		Id:           info.TaskId,
		FileHash:     info.FileHash,
		FileName:     info.FileName,
		Path:         info.FilePath,
		CopyNum:      info.CopyNum,
		Type:         pType,
		StoreType:    info.StoreType,
		Status:       info.TaskState,
		DetailStatus: info.ProgressState,
		FileSize:     info.Total * dspCom.CHUNK_SIZE / 1024,
		Nodes:        npros,
		CreatedAt:    info.CreatedAt,
		UpdatedAt:    info.UpdatedAt,
	}
	pInfo.IsUploadAction = (info.Type == task.TaskTypeUpload)
	pInfo.Progress = 0
	// log.Debugf("get transfer %s detail total %d sum %d ret %v err %s info.type %d", info.TaskKey, info.Total, sum, info.Result, info.ErrorMsg, info.Type)
	switch pType {
	case transferTypeUploading:
		if info.Total > 0 && sum >= info.Total && info.Result != nil && len(info.ErrorMsg) == 0 {
			return nil
		}

		pInfo.UploadSize = sum * dspCom.CHUNK_SIZE / 1024
		if len(pInfo.Nodes) > 0 && pInfo.FileSize > 0 {
			pInfo.Progress = (float64(pInfo.UploadSize) / float64(pInfo.FileSize)) / float64(len(pInfo.Nodes))
		}
	case transferTypeDownloading:
		if info.Total > 0 && sum >= info.Total {
			return nil
		}

		pInfo.DownloadSize = sum * dspCom.CHUNK_SIZE / 1024
		if pInfo.FileSize > 0 {
			pInfo.Progress = float64(pInfo.DownloadSize) / float64(pInfo.FileSize)
		}
	case transferTypeComplete:
		if sum < info.Total || info.Total == 0 {
			return nil
		}
		if info.TaskState == task.TaskStateFailed {
			return nil
		}
		if info.Type == task.TaskTypeUpload {
			if info.Result == nil {
				return nil
			}
			pInfo.UploadSize = sum * dspCom.CHUNK_SIZE / 1024
			if pInfo.UploadSize == 0 {
				return nil
			}
			if pInfo.Status != task.TaskStateDone && pInfo.FileSize > 0 && pInfo.UploadSize == pInfo.FileSize*uint64(len(pInfo.Nodes)) {
				log.Warnf("task:%s taskstate is %d, status:%d, but it has done", info.TaskId, info.TaskState, pInfo.Status)
				pInfo.Status = task.TaskStateDone
			}
			if len(pInfo.Nodes) > 0 && pInfo.FileSize > 0 {
				pInfo.Progress = (float64(pInfo.UploadSize) / float64(pInfo.FileSize)) / float64(len(pInfo.Nodes))
			}
		} else if info.Type == task.TaskTypeDownload {
			pInfo.DownloadSize = sum * dspCom.CHUNK_SIZE / 1024
			if pInfo.DownloadSize == 0 {
				return nil
			}
			if pInfo.Status != task.TaskStateDone && pInfo.FileSize > 0 && pInfo.DownloadSize == pInfo.FileSize {
				pInfo.Status = task.TaskStateDone
				log.Warnf("task:%s taskstate is %d, but it has done", info.TaskId, info.TaskState)
			}
			if pInfo.FileSize > 0 {
				pInfo.Progress = float64(pInfo.DownloadSize) / float64(pInfo.FileSize)
			}
			pInfo.Encrypted = this.Dsp.IsFileEncrypted(pInfo.Path)
		}
	}
	if info.TaskState == task.TaskStateFailed {
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
