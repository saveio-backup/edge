package dsp

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	dspCom "github.com/saveio/dsp-go-sdk/common"
	"github.com/saveio/dsp-go-sdk/task"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	http "github.com/saveio/edge/http/common"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/log"
)

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
	Result         interface{} `json:",omitempty"`
	ErrMsg         string      `json:",omitempty"`
}

type transferlistResp struct {
	IsTransfering bool
	Transfers     []*transfer
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
				log.Infof("+++++++= file:%s, hash:%s, total:%d, peer:%s, uploaded:%d, progress:%f", v.FileName, v.FileHash, v.Total, node, cnt, float64(cnt)/float64(v.Total))
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
	log.Debugf("allTasksKey:%v", allTasksKey)
	for _, key := range allTasksKey {
		info, err := this.GetProgress(key)
		if err != nil {
			continue
		}
		log.Debugf("==== range key %v", key)
		if pType == transferTypeUploading && info.Type != task.TaskTypeUpload {
			continue
		}
		if pType == transferTypeDownloading && info.Type != task.TaskTypeDownload {
			continue
		}

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
			FileHash: info.FileHash,
			FileName: info.FileName,
			Type:     pType,
			Status:   transferStatusDoing,
			FileSize: info.Total * dspCom.CHUNK_SIZE / 1024,
			Nodes:    npros,
		}
		pInfo.IsUploadAction = (info.Type == task.TaskTypeUpload)
		pInfo.Progress = 0
		switch pType {
		case transferTypeUploading:
			if info.Total > 0 && sum >= info.Total && (info.Result != nil || info.Error != nil) {
				continue
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
				continue
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
				continue
			}
			if info.Type == task.TaskTypeUpload {
				if info.Result == nil {
					continue
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
				pInfo.Path = config.Parameters.FsConfig.FsFileRoot + "/" + pInfo.FileName
			}
		}
		if info.Error != nil {
			pInfo.ErrMsg = info.Error.Error()
			pInfo.Status = transferStatusFailed
		}
		if info.Result != nil {
			pInfo.Result = info.Result
		}
		if !resp.IsTransfering {
			resp.IsTransfering = (pType == transferTypeUploading || pType == transferTypeDownloading) && (pInfo.Status != transferStatusFailed && pInfo.Status != transferStatusDone)
		}

		if off < offset {
			off++
			continue
		}
		log.Debugf("set transfer info.FileHash %v", pInfo.FileHash)
		log.Debugf("set transfer info.FileName %v", pInfo.FileName)
		log.Debugf("set transfer info.Progress %v", pInfo.Progress)
		log.Debugf("set transfer info.Result %v", pInfo.Result)
		log.Debugf("set transfer info.ErrMsg %v", pInfo.ErrMsg)
		infos = append(infos, pInfo)
		off++
		if limit > 0 && uint64(len(infos)) >= limit {
			break
		}
	}
	resp.Transfers = infos
	return resp
}

func (this *Endpoint) CalculateUploadFee(filePath string, proveInterval uint64, proveTimes, copyNum uint32, whitelistCnt uint64) (uint64, error) {
	return this.Dsp.CalculateUploadFee(filePath, &dspCom.UploadOption{
		ProveInterval: proveInterval,
		ProveTimes:    proveTimes,
		CopyNum:       copyNum,
	}, uint64(whitelistCnt))
}

type downloadFileInfo struct {
	Hash      string
	Name      string
	Ext       string
	Size      uint64
	Fee       uint64
	FeeFormat string
}

func (this *Endpoint) GetDownloadFileInfo(url string) *downloadFileInfo {
	info := &downloadFileInfo{}
	var fileLink string
	if strings.HasPrefix(url, dspCom.FILE_URL_CUSTOM_HEADER) {
		fileLink = this.Dsp.GetLinkFromUrl(url)
	} else if strings.HasPrefix(url, dspCom.FILE_LINK_PREFIX) {
		fileLink = url
	} else if strings.HasPrefix(url, dspCom.PROTO_NODE_PREFIX) || strings.HasPrefix(url, dspCom.RAW_NODE_PREFIX) {
		// TODO support get download file info from hash
		return nil
	}
	if len(fileLink) == 0 {
		return nil
	}
	values := this.Dsp.GetLinkValues(fileLink)
	if values == nil {
		return nil
	}
	info.Hash = values[dspCom.FILE_LINK_HASH_KEY]
	info.Name = values[dspCom.FILE_LINK_NAME_KEY]
	blockNumStr := values[dspCom.FILE_LINK_BLOCKNUM_KEY]
	blockNum, err := strconv.ParseUint(blockNumStr, 10, 64)
	if err != nil {
		return nil
	}
	info.Size = blockNum * dspCom.CHUNK_SIZE / 1024
	extParts := strings.Split(info.Name, ".")
	if len(extParts) > 1 {
		info.Ext = extParts[len(extParts)-1]
	}
	info.Fee = blockNum * dspCom.CHUNK_SIZE * common.DSP_DOWNLOAD_UNIT_PRICE
	info.FeeFormat = utils.FormatUsdt(info.Fee)
	return info
}

func (this *Endpoint) EncryptFile(path, password string) error {
	err := this.Dsp.Fs.AESEncryptFile(path, password, path+".temp")
	if err != nil {
		return err
	}
	return os.Rename(path+".temp", path)
}

func (this *Endpoint) DecryptFile(path, password string) error {
	err := this.Dsp.Fs.AESDecryptFile(path, password, path+".temp")
	if err != nil {
		return err
	}
	return os.Rename(path+".temp", path)
}

type fileShareIncome struct {
	Name         string
	Profit       uint64
	ProfitFormat string
	SharedAt     uint64
}

type FileShareIncomeResp struct {
	TotalIncome       uint64
	TotalIncomeFormat string
	Incomes           []*fileShareIncome
}

func (this *Endpoint) GetFileRevene() (uint64, error) {
	if this.sqliteDB == nil {
		return 0, errors.New("sqlite db is nil")
	}
	return this.sqliteDB.SumRecordsProfit()
}

func (this *Endpoint) GetFileShareIncome(start, end, offset, limit uint64) (*FileShareIncomeResp, error) {
	resp := &FileShareIncomeResp{}
	records, err := this.sqliteDB.FineShareRecordsByCreatedAt(int64(start), int64(end), int64(offset), int64(limit))
	if err != nil {
		return nil, err
	}
	resp.Incomes = make([]*fileShareIncome, 0, len(records))
	for _, record := range records {
		resp.TotalIncome += record.Profit
		fileName := ""
		info, _ := this.Dsp.DownloadedFileInfo(record.FileHash)
		if info != nil {
			fileName = info.FileName
		}
		resp.Incomes = append(resp.Incomes, &fileShareIncome{
			Name:         fileName,
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
}

func (this *Endpoint) GetUploadFiles(fileType http.DSP_FILE_LIST_TYPE, offset, limit uint64) ([]*fileResp, error) {
	fileList, err := this.Chain.Native.Fs.GetFileList()
	if err != nil {
		return nil, err
	}
	now, err := this.Chain.GetCurrentBlockHeight()
	if err != nil {
		return nil, err
	}
	files := make([]*fileResp, 0)
	offsetCnt := uint64(0)
	for _, hash := range fileList.List {
		fi, err := this.Chain.Native.Fs.GetFileInfo(string(hash.Hash))
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
		}
		files = append(files, fr)
		if limit > 0 && uint64(len(files)) >= limit {
			break
		}
	}
	return files, nil
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
}

func (this *Endpoint) GetDownloadFiles(fileType http.DSP_FILE_LIST_TYPE, offset, limit uint64) ([]*downloadFilesInfo, error) {
	fileInfos := make([]*downloadFilesInfo, 0)
	if this.Dsp == nil {
		return nil, errors.New("dsp is nil")
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
		})
		if uint64(len(fileInfos)) > limit {
			break
		}
	}
	return fileInfos, nil
}
