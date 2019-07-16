package dsp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	dspCom "github.com/saveio/dsp-go-sdk/common"
	"github.com/saveio/dsp-go-sdk/task"
	"github.com/saveio/edge/common"
	clicom "github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp/storage"
	chainSdkFs "github.com/saveio/themis-go-sdk/fs"
	"github.com/saveio/themis-go-sdk/usdt"
	"github.com/saveio/themis/cmd/utils"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/smartcontract/service/native/onifs"
	fs "github.com/saveio/themis/smartcontract/service/native/onifs"
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
	Tx         string
	IsUploaded bool
}

type TransferType int

const (
	transferTypeComplete TransferType = iota
	transferTypeUploading
	transferTypeDownloading
)

type transferStatus int

const (
	transferStatusPause transferStatus = iota
	transferStatusPreparing
	transferStatusDoing
	transferStatusDone
	transferStatusFailed
)

type nodeProgress struct {
	HostAddr     string
	UploadSize   uint64
	DownloadSize uint64
}

type transfer struct {
	Id             string
	FileHash       string
	FileName       string
	Type           TransferType
	Status         transferStatus
	Path           string
	IsUploadAction bool
	UploadSize     uint64
	DownloadSize   uint64
	FileSize       uint64
	Nodes          []*nodeProgress
	Progress       float64
	CreatedAt      uint64
	UpdatedAt      uint64
	Result         interface{} `json:",omitempty"`
	ErrMsg         string      `json:",omitempty"`
}

type transferlistResp struct {
	IsTransfering bool
	Transfers     []*transfer
}

type downloadFileInfo struct {
	Hash      string
	Name      string
	Ext       string
	Size      uint64
	Fee       uint64
	FeeFormat string
}

type fileShareIncome struct {
	Name         string
	OwnerAddress string
	Profit       uint64
	ProfitFormat string
	SharedAt     uint64
}

type FileShareIncomeResp struct {
	TotalIncome       uint64
	TotalIncomeFormat string
	Incomes           []*fileShareIncome
}

type fileResp struct {
	Hash          string
	Name          string
	Url           string
	Size          uint64
	DownloadCount uint64
	ExpiredAt     uint64
	UpdatedAt     uint64
	Profit        uint64
	Privilege     uint64
	StoreType     dspCom.FileStoreType
}

type downloadFilesInfo struct {
	Hash          string
	Name          string
	Url           string
	Size          uint64
	DownloadCount uint64
	DownloadAt    uint64
	LastShareAt   uint64
	Profit        uint64
	ProfitFormat  string
}

type WhiteListRule struct {
	Addr          string
	StartHeight   uint64
	ExpiredHeight uint64
}

type userspace struct {
	Used      uint64
	Remain    uint64
	ExpiredAt uint64
	Balance   uint64
}

type UserspaceRecordResp struct {
	Size       uint64
	ExpiredAt  uint64
	Cost       uint64
	CostFormat string
}

type calculateResp struct {
	Fee       uint64
	FeeFormat string
}

type userspaceCostResp struct {
	Fee          uint64
	FeeFormat    string
	Refund       uint64
	RefundFormat string
	TransferType storage.UserspaceTransferType
}

type FsContractSettingResp struct {
	DefaultCopyNum     uint64
	DefaultProvePeriod uint64
	MinChallengeRate   uint64
	MinVolume          uint64
}

func (this *Endpoint) UploadFile(path, desc string, durationVal, intervalVal, timesVal, privilegeVal, copynumVal, storageTypeVal interface{},
	encryptPwd, url string, whitelist []string, share bool) *DspErr {
	currentAccount := this.Dsp.CurrentAccount()
	fssetting, err := this.Dsp.Chain.Native.Fs.GetSetting()
	if err != nil {
		return &DspErr{Code: FS_GET_SETTING_FAILED, Error: err}
	}
	duration, _ := durationVal.(float64)
	interval, ok := intervalVal.(float64)
	if !ok || interval == 0 {
		interval = float64(fssetting.DefaultProvePeriod)
	}
	times, ok := timesVal.(float64)
	storageType, _ := storageTypeVal.(float64)
	if !ok || times == 0 {
		//TODO
		if dspCom.FileStoreType(storageType) == dspCom.FileStoreTypeNormal {
			userspace, err := this.Dsp.Chain.Native.Fs.GetUserSpace(currentAccount.Address)
			if err != nil {
				return &DspErr{Code: FS_GET_USER_SPACE_FAILED, Error: err}
			}
			currentHeight, err := this.Dsp.Chain.GetCurrentBlockHeight()
			if err != nil {
				return &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
			}
			log.Debugf("storageType %d, userspace.ExpireHeight %d, current: %d", storageType, userspace.ExpireHeight, currentHeight)
			if userspace.ExpireHeight <= uint64(currentHeight) {
				return &DspErr{Code: DSP_USER_SPACE_EXPIRED, Error: ErrMaps[DSP_USER_SPACE_EXPIRED]}
			}
			if duration > 0 && (uint64(currentHeight)+uint64(duration)) > userspace.ExpireHeight {
				return &DspErr{Code: DSP_USER_SPACE_PERIOD_NOT_ENOUGH, Error: err}
			}
			if duration == 0 {
				duration = float64(userspace.ExpireHeight) - float64(currentHeight)
			}
			log.Debugf("userspace.ExpireHeight %d, current %d, duration :%v, times :%v", userspace.ExpireHeight, currentHeight, duration, times)
		}
		times = math.Ceil(duration / float64(interval))
	}
	if times == 0 {
		return &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	privilege, ok := privilegeVal.(float64)
	if !ok {
		privilege = onifs.PUBLIC
	}
	copynum, ok := copynumVal.(float64)
	if !ok {
		copynum = float64(fssetting.DefaultCopyNum)
	}
	if len(url) == 0 {
		// random
		b := make([]byte, clicom.DSP_URL_RAMDOM_NAME_LEN/2)
		_, err := rand.Read(b)
		if err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
		url = this.Dsp.Chain.Native.Dns.GetCustomUrlHeader() + hex.EncodeToString(b)
	}
	find, err := this.Dsp.Chain.Native.Dns.QueryUrl(url, this.Dsp.CurrentAccount().Address)
	if find != nil || err == nil {
		return &DspErr{Code: DSP_UPLOAD_URL_EXIST, Error: fmt.Errorf("url exist err %s", err)}
	}

	opt := &dspCom.UploadOption{
		FileDesc:        desc,
		ProveInterval:   uint64(interval),
		ProveTimes:      uint32(times),
		Privilege:       uint32(privilege),
		CopyNum:         uint32(copynum),
		Encrypt:         len(encryptPwd) > 0,
		EncryptPassword: encryptPwd,
		RegisterDns:     len(url) > 0,
		BindDns:         len(url) > 0,
		DnsUrl:          url,
		WhiteList:       whitelist,
		Share:           share,
		StorageType:     dspCom.FileStoreType(storageType),
	}
	optBuf, _ := json.Marshal(opt)
	log.Debugf("path %s, UploadOption :%s\n", path, optBuf)
	go func() {
		log.Debugf("upload file path %s", path)
		ret, err := this.Dsp.UploadFile(path, opt)
		if err != nil {
			log.Errorf("upload failed err %s", err)
			return
		}
		log.Info(ret)
	}()
	return nil
}

func (this *Endpoint) DeleteFile(fileHash string) (*DeleteFileResp, *DspErr) {
	fi, err := this.Dsp.Chain.Native.Fs.GetFileInfo(fileHash)
	if fi != nil && err == nil && fi.FileOwner.ToBase58() == this.Dsp.WalletAddress() {
		resp, err := this.Dsp.DeleteUploadedFile(fileHash)
		if err != nil {
			return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: err}
		}
		return &DeleteFileResp{Tx: resp.Tx, IsUploaded: false}, nil
	}
	err = this.Dsp.DeleteDownloadedFile(fileHash)
	if err != nil {
		return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: err}
	}
	return &DeleteFileResp{IsUploaded: true}, nil
}

func (this *Endpoint) GetFsConfig() (*FsContractSettingResp, *DspErr) {
	set, err := this.Dsp.Chain.Native.Fs.GetSetting()
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}

	return &FsContractSettingResp{
		DefaultCopyNum:     set.DefaultCopyNum,
		DefaultProvePeriod: set.DefaultProvePeriod,
		MinChallengeRate:   set.MinChallengeRate,
		MinVolume:          set.MinVolume,
	}, nil
}

func (this *Endpoint) DownloadFile(fileHash, url, link, password string, max uint64, setFileName bool) *DspErr {
	if len(fileHash) > 0 {
		go func() {
			// TODO: get file name
			err := this.Dsp.DownloadFile(fileHash, "", dspCom.ASSET_USDT, true, password, false, setFileName, int(max))
			if err != nil {
				log.Errorf("Downloadfile from url failed %s", err)
			}
		}()
		return nil
	}
	if len(url) > 0 {
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
		go func() {
			err := this.Dsp.DownloadFileByLink(fileHash, dspCom.ASSET_USDT, true, password, false, setFileName, int(max))
			if err != nil {
				log.Errorf("Downloadfile from url failed %s", err)
			}
		}()
		return nil
	}
	return nil
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
			log.Debugf("progress store file %s, %v, ok %t", v.TaskKey, v, ok)
			for node, cnt := range v.Count {
				log.Infof("prorgess type:%d file:%s, hash:%s, total:%d, peer:%s, uploaded:%d, progress:%f", v.Type, v.FileName, v.FileHash, v.Total, node, cnt, float64(cnt)/float64(v.Total))
			}
		case <-this.closhCh:
			this.Dsp.CloseProgressChannel()
			return
		}
	}
}

// GetTransferList. get transfer progress list
func (this *Endpoint) GetTransferList(pType TransferType, offset, limit uint64) *transferlistResp {
	infos := make([]*transfer, 0)
	off := uint64(0)
	resp := &transferlistResp{
		IsTransfering: false,
		Transfers:     []*transfer{},
	}
	allTasksKey, err := this.GetAllProgressKeys()
	if err != nil {
		log.Errorf("get all task keys err %s", err)
		return resp
	}
	// log.Debugf("allTasksKey :%v", allTasksKey)
	for idx, key := range allTasksKey {
		info, err := this.GetProgress(key)
		if err != nil {
			log.Errorf("get progress failed %d for %s info %v err %s", idx, key, info, err)
			continue
		}
		if pType == transferTypeUploading && info.Type != task.TaskTypeUpload {
			continue
		}
		if pType == transferTypeDownloading && info.Type != task.TaskTypeDownload {
			continue
		}
		pInfo := this.getTransferDetail(pType, info)
		// log.Debugf("get pinfo  %v of %v", pInfo, key)
		if pInfo == nil {
			continue
		}
		if !resp.IsTransfering {
			resp.IsTransfering = (pType == transferTypeUploading || pType == transferTypeDownloading) && (pInfo.Status != transferStatusFailed && pInfo.Status != transferStatusDone)
		}

		if off < offset {
			off++
			continue
		}
		// log.Debugf("#%d set transfer type %d info.FileHash %v, fileName %s, progress:%v, result %v, err %s, status %d", idx, pType, pInfo.FileHash, pInfo.FileName, pInfo.Progress, pInfo.Result, pInfo.ErrMsg, pInfo.Status)
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
func (this *Endpoint) GetTransferDetail(tType task.TaskType, fileHash, url string) (*transfer, *DspErr) {
	if len(url) == 0 && len(fileHash) == 0 {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
	}
	resp := &transfer{}
	allTasksKey, err := this.GetAllProgressKeys()
	if err != nil {
		return resp, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	hash := fileHash
	if len(hash) == 0 {
		hash = this.Dsp.GetFileHashFromUrl(url)
	}
	log.Debugf("get hash from url %s %s", url, hash)
	for _, key := range allTasksKey {
		info, err := this.GetProgress(key)
		if err != nil {
			continue
		}
		if info.Type != task.TaskTypeDownload {
			continue
		}
		if info.FileHash != hash {
			continue
		}
		// pInfo := this.getTransferDetail(pType, info)
		// if pInfo == nil {
		// 	return resp, &DspErr{Code: INTERNAL_ERROR, Error: ErrMaps[INTERNAL_ERROR]}
		// }
		// resp = pInfo
		break
	}
	return resp, nil
}

func (this *Endpoint) CalculateUploadFee(filePath string, durationVal, intervalVal, timesVal, copynumVal, whitelistVal, storeType interface{}) (*calculateResp, *DspErr) {
	currentAccount := this.Dsp.CurrentAccount()
	fssetting, err := this.Dsp.Chain.Native.Fs.GetSetting()
	if err != nil {
		return nil, &DspErr{Code: FS_GET_SETTING_FAILED, Error: err}
	}
	duration, err := OptionStrToFloat64(durationVal, 0)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	interval, err := OptionStrToFloat64(intervalVal, float64(fssetting.DefaultProvePeriod))
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	times, err := OptionStrToFloat64(timesVal, 0)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	if times == 0 {
		userspace, err := this.Dsp.Chain.Native.Fs.GetUserSpace(currentAccount.Address)
		if err != nil {
			return nil, &DspErr{Code: FS_GET_USER_SPACE_FAILED, Error: err}
		}
		if userspace == nil {
			return nil, &DspErr{Code: FS_GET_USER_SPACE_FAILED, Error: err}
		}
		currentHeight, err := this.Dsp.Chain.GetCurrentBlockHeight()
		if err != nil {
			return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
		}
		if userspace.ExpireHeight <= uint64(currentHeight) {
			return nil, &DspErr{Code: DSP_USER_SPACE_EXPIRED, Error: ErrMaps[DSP_USER_SPACE_EXPIRED]}
		}
		if duration > 0 && (uint64(currentHeight)+uint64(duration)) > userspace.ExpireHeight {
			return nil, &DspErr{Code: DSP_USER_SPACE_NOT_ENOUGH, Error: ErrMaps[DSP_USER_SPACE_NOT_ENOUGH]}
		}
		if duration == 0 {
			duration = float64(userspace.ExpireHeight) - float64(currentHeight)
		}
		times = math.Ceil(duration / float64(interval))
		log.Debugf("userspace.ExpireHeight %d, current %d, duration :%v, times :%v", userspace.ExpireHeight, currentHeight, duration, times)
	}
	copynum, err := OptionStrToFloat64(copynumVal, float64(fssetting.DefaultCopyNum))
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	wh, err := OptionStrToFloat64(whitelistVal, 0)
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	sType, _ := OptionStrToFloat64(storeType, 0)
	fee, err := this.Dsp.CalculateUploadFee(filePath, &dspCom.UploadOption{
		ProveInterval: uint64(interval),
		ProveTimes:    uint32(times),
		CopyNum:       uint32(copynum),
		StorageType:   dspCom.FileStoreType(sType),
	}, uint64(wh))
	if err != nil {
		return nil, &DspErr{Code: DSP_CALC_UPLOAD_FEE_FAILED, Error: err}
	}
	feeFormat := utils.FormatUsdt(fee)
	return &calculateResp{
		Fee:       fee,
		FeeFormat: feeFormat,
	}, nil
}

func (this *Endpoint) GetDownloadFileInfo(url string) (*downloadFileInfo, *DspErr) {
	info := &downloadFileInfo{}
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
	err := this.Dsp.Fs.AESDecryptFile(path, password, path+".temp")
	if err != nil {
		return &DspErr{Code: DSP_DECRYPTED_FILE_FAILED, Error: err}
	}
	err = os.Rename(path+".temp", path)
	if err != nil {
		return &DspErr{Code: DSP_DECRYPTED_FILE_FAILED, Error: err}
	}
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
	resp.Incomes = make([]*fileShareIncome, 0, len(records))
	for _, record := range records {
		if record.Profit == 0 {
			continue
		}
		resp.TotalIncome += record.Profit
		fileName := ""
		info, _ := this.Dsp.DownloadedFileInfo(record.FileHash)
		if info != nil {
			fileName = info.FileName
		}
		fileInfo, _ := this.Dsp.Chain.Native.Fs.GetFileInfo(record.FileHash)
		owner := ""
		if fileInfo != nil {
			owner = fileInfo.FileOwner.ToBase58()
		}
		resp.Incomes = append(resp.Incomes, &fileShareIncome{
			Name:         fileName,
			OwnerAddress: owner,
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
				_, err := this.sqliteDB.InsertShareRecord(id, v.FileHash, v.ToWalletAddr, v.PaymentAmount)
				if err != nil {
					log.Errorf("insert new share_record failed %s, err %s", id, err)
				}
			case task.ShareStateReceivedPaying, task.ShareStateEnd:
				_, err := this.sqliteDB.IncreaseShareRecordProfit("", v.TaskKey, v.PaymentAmount)
				if err != nil {
					log.Errorf("increase share_record profit failed %s, err %s", v.TaskKey, err)
				}
			default:
				log.Warn("unknown state type")
			}

		case <-this.closhCh:
			this.Dsp.CloseShareNotificationChannel()
			return
		}
	}
}

func (this *Endpoint) GetUploadFiles(fileType DspFileListType, offset, limit uint64) ([]*fileResp, *DspErr) {
	fileList, err := this.Dsp.Chain.Native.Fs.GetFileList()
	if err != nil {
		return nil, &DspErr{Code: FS_GET_FILE_LIST_FAILED, Error: err}
	}
	// setting, err := this.Dsp.Chain.Native.Fs.GetSetting()
	// if err != nil {
	// 	return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	// }
	now, err := this.Dsp.Chain.GetCurrentBlockHeight()
	if err != nil {
		return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	}
	// space, err := this.Dsp.GetUserSpace(this.Dsp.WalletAddress())
	// if err != nil {
	// 	return nil, &DspErr{Code: CHAIN_GET_HEIGHT_FAILED, Error: err}
	// }

	files := make([]*fileResp, 0)
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

		expired := fi.BlockHeight + (fi.ChallengeRate * fi.ChallengeTimes)
		expiredAt := uint64(time.Now().Unix()) + (expired - uint64(now))
		updatedAt := uint64(time.Now().Unix()) + (fi.BlockHeight - uint64(now))
		url, _ := this.GetUrlFromHash(string(hash.Hash))
		downloadedCount, _ := this.sqliteDB.CountRecordByFileHash(string(hash.Hash))
		profit, _ := this.sqliteDB.SumRecordsProfitByFileHash(string(hash.Hash))

		fr := &fileResp{
			Hash:          string(hash.Hash),
			Name:          string(fi.FileDesc),
			Url:           url,
			Size:          fi.FileBlockNum * fi.FileBlockSize,
			DownloadCount: downloadedCount,
			ExpiredAt:     expiredAt,
			// TODO fix by db
			UpdatedAt: updatedAt,
			Profit:    profit,
			Privilege: fi.Privilege,
			StoreType: dspCom.FileStoreType(fi.StorageType),
		}
		files = append(files, fr)
		if limit > 0 && uint64(len(files)) >= limit {
			break
		}
	}
	return files, nil
}

type fileInfoResp struct {
	FileHash   string
	CreatedAt  uint64
	CopyNum    uint64
	Interval   uint64
	ProveTimes uint64
	Priviledge uint64
	Whitelist  []string
	ExpiredAt  uint64
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
	expired := info.BlockHeight + (info.ChallengeRate * info.ChallengeTimes)
	expiredAt := uint64(time.Now().Unix()) + (expired - uint64(now))
	result := &fileInfoResp{
		FileHash:   string(info.FileHash),
		CopyNum:    info.CopyNum,
		Interval:   info.ChallengeRate,
		ProveTimes: info.ChallengeTimes,
		Priviledge: info.Privilege,
		Whitelist:  []string{},
		ExpiredAt:  expiredAt,
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

func (this *Endpoint) GetDownloadFiles(fileType DspFileListType, offset, limit uint64) ([]*downloadFilesInfo, *DspErr) {
	fileInfos := make([]*downloadFilesInfo, 0)
	if this.Dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	files := this.Dsp.AllDownloadFiles()
	offsetCnt := uint64(0)
	for _, file := range files {
		info, err := this.Dsp.DownloadedFileInfo(file)
		if err != nil || info == nil {
			continue
		}
		url, err := this.GetUrlFromHash(file)
		if err != nil {
			log.Errorf("get url from hash %s, err %s", file, err)
		}
		// 0: all, 1. image, 2. document. 3. video, 4. music
		fileName := strings.ToLower(info.FileName)
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
		fileInfos = append(fileInfos, &downloadFilesInfo{
			Hash:          file,
			Name:          info.FileName,
			Url:           url,
			Size:          uint64(len(info.BlockHashes)) * dspCom.CHUNK_SIZE / 1024,
			DownloadCount: downloadedCount,
			DownloadAt:    info.CreatedAt,
			LastShareAt:   lastSharedAt,
			Profit:        profit,
			ProfitFormat:  utils.FormatUsdt(profit),
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
	log.Debugf("get event err %s, event :%v", err, event)
	if err != nil || event == nil {
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
		if n.ContractAddress != usdt.USDT_CONTRACT_ADDRESS.ToHexString() || to != chainSdkFs.FS_CONTRACT_ADDRESS.ToBase58() {
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

func (this *Endpoint) GetUserSpaceCost(walletAddr string, size, sizeOpType, blockCount, countOpType uint64) (*userspaceCostResp, *DspErr) {
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
		return &userspaceCostResp{
			Fee:          cost.Value,
			FeeFormat:    utils.FormatUsdt(cost.Value),
			TransferType: storage.TransferTypeIn,
		}, nil
	} else if cost.To.ToBase58() == this.Dsp.Account.Address.ToBase58() {
		return &userspaceCostResp{
			Refund:       cost.Value,
			RefundFormat: utils.FormatUsdt(cost.Value),
			TransferType: storage.TransferTypeOut,
		}, nil
	}
	return nil, &DspErr{Code: INTERNAL_ERROR, Error: ErrMaps[INTERNAL_ERROR]}
}

func (this *Endpoint) GetUserSpace(addr string) (*userspace, *DspErr) {
	space, err := this.Dsp.GetUserSpace(addr)
	if err != nil || space == nil {
		log.Errorf("get user space err %s, space %v", err, space)
		return &userspace{
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
	if space.ExpireHeight > uint64(currentHeight) {
		blk, err := this.Dsp.Chain.GetBlockByHeight(uint32(updateHeight))
		if err != nil {
			return nil, &DspErr{Code: CHAIN_GET_BLK_BY_HEIGHT_FAILED, Error: err}
		}
		expiredAt = uint64(blk.Header.Timestamp) + (space.ExpireHeight - uint64(updateHeight))
		log.Debugf("expiredAt %d height %d, expiredheight %d updatedheight %d", expiredAt, blk.Header.Timestamp, space.ExpireHeight, updateHeight)
	} else {
		space, err := this.GetUserspaceRecords(addr, 0, 1)
		if err != nil || len(space) == 0 {
			expiredAt = now
			log.Debugf("no space expiredAt %d ", expiredAt)
		} else {
			expiredAt = space[0].ExpiredAt
			log.Debugf(" space[0] expiredAt %d ", expiredAt)
		}
	}
	log.Debugf("expiredAt %d, now %d ", expiredAt, now)
	if expiredAt <= now {
		return &userspace{
			Used:      0,
			Remain:    0,
			ExpiredAt: expiredAt,
			Balance:   space.Balance,
		}, nil
	}
	return &userspace{
		Used:      space.Used,
		Remain:    space.Remain,
		ExpiredAt: expiredAt,
		Balance:   space.Balance,
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
		resp = append(resp, &UserspaceRecordResp{
			Size:       record.TotalSize,
			ExpiredAt:  record.ExpiredAt,
			Cost:       record.Amount,
			CostFormat: utils.FormatUsdt(record.Amount),
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

func (this *Endpoint) getTransferDetail(pType TransferType, info *task.ProgressInfo) *transfer {
	sum := uint64(0)
	npros := make([]*nodeProgress, 0)
	for haddr, cnt := range info.Count {
		sum += cnt
		pros := &nodeProgress{
			HostAddr: haddr,
		}
		if info.Type == task.TaskTypeUpload {
			pros.UploadSize = cnt * dspCom.CHUNK_SIZE / 1024
		} else if info.Type == task.TaskTypeDownload {
			pros.DownloadSize = cnt * dspCom.CHUNK_SIZE / 1024
		}
		npros = append(npros, pros)
	}
	pInfo := &transfer{
		Id:        info.TaskKey,
		FileHash:  info.FileHash,
		FileName:  info.FileName,
		Type:      pType,
		Status:    transferStatusDoing,
		FileSize:  info.Total * dspCom.CHUNK_SIZE / 1024,
		Nodes:     npros,
		CreatedAt: info.CreatedAt,
		UpdatedAt: info.UpdatedAt,
	}
	pInfo.IsUploadAction = (info.Type == task.TaskTypeUpload)
	pInfo.Progress = 0
	// log.Debugf("get transfer %s detail total %d sum %d ret %v err %s info.type %d", info.TaskKey, info.Total, sum, info.Result, info.ErrorMsg, info.Type)
	switch pType {
	case transferTypeUploading:
		if info.Total > 0 && sum >= info.Total && info.Result != nil && len(info.ErrorMsg) == 0 {
			return nil
		}
		if info.Total == 0 {
			pInfo.Status = transferStatusPreparing
		}
		pInfo.UploadSize = sum * dspCom.CHUNK_SIZE / 1024
		if len(pInfo.Nodes) > 0 && pInfo.FileSize > 0 {
			pInfo.Progress = (float64(pInfo.UploadSize) / float64(pInfo.FileSize)) / float64(len(pInfo.Nodes))
		}
	case transferTypeDownloading:
		if info.Total > 0 && sum >= info.Total {
			return nil
		}
		if info.Total == 0 {
			pInfo.Status = transferStatusPreparing
		}
		pInfo.DownloadSize = sum * dspCom.CHUNK_SIZE / 1024
		if pInfo.FileSize > 0 {
			pInfo.Progress = float64(pInfo.DownloadSize) / float64(pInfo.FileSize)
		}
	case transferTypeComplete:
		if sum < info.Total || info.Total == 0 {
			return nil
		}
		if info.Type == task.TaskTypeUpload {
			if info.Result == nil {
				return nil
			}
			pInfo.UploadSize = sum * dspCom.CHUNK_SIZE / 1024
			if pInfo.FileSize > 0 && pInfo.UploadSize == pInfo.FileSize*uint64(len(pInfo.Nodes)) {
				pInfo.Status = transferStatusDone
			}
			if len(pInfo.Nodes) > 0 && pInfo.FileSize > 0 {
				pInfo.Progress = (float64(pInfo.UploadSize) / float64(pInfo.FileSize)) / float64(len(pInfo.Nodes))
			}
		} else if info.Type == task.TaskTypeDownload {
			pInfo.DownloadSize = sum * dspCom.CHUNK_SIZE / 1024
			if pInfo.FileSize > 0 && pInfo.DownloadSize == pInfo.FileSize {
				pInfo.Status = transferStatusDone
			}
			if pInfo.FileSize > 0 {
				pInfo.Progress = float64(pInfo.DownloadSize) / float64(pInfo.FileSize)
			}
			//TODO: use from progress
			pInfo.Path = config.FsFileRootPath() + "/" + pInfo.FileName
		}
	}
	if len(info.ErrorMsg) != 0 {
		pInfo.ErrMsg = info.ErrorMsg
		pInfo.Status = transferStatusFailed
	}
	if info.Result != nil {
		pInfo.Result = info.Result
	}
	return pInfo
}
