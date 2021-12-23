package dsp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/saveio/dsp-go-sdk/task/poc"
	tskUtils "github.com/saveio/dsp-go-sdk/utils/task"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/utils"
	"github.com/saveio/edge/utils/plot"
	"github.com/saveio/themis/common/log"
)

func (this *Endpoint) NewPocTask(cfg *plot.PlotConfig) (string, *DspErr) {

	dsp := this.getDsp()
	if dsp == nil {
		return "", NewDspErr(NO_DSP)
	}

	fileName := tskUtils.GetPlotFileName(cfg.Nonces, cfg.StartNonce, cfg.NumericID)
	fileName = filepath.Join(cfg.Path, fileName)

	existId, _ := dsp.GetPocTaskIdByFileName(fileName)
	if len(existId) > 0 {
		log.Debugf("file name %s, exist taskId %v", fileName, existId)
		return existId, nil
	}

	taskId, err := dsp.NewPocTask("")
	if err != nil {
		return "", NewDspErr(DSP_TASK_POC_ERROR, err)
	}
	return taskId, nil
}

func (this *Endpoint) GenPlotPDPData(taskId string, cfg *plot.PlotConfig) *DspErr {

	dsp := this.getDsp()
	if dsp == nil {
		return NewDspErr(NO_DSP)
	}
	cfgData, _ := json.Marshal(cfg)
	log.Infof("plot config cfg with no size %s", cfgData)
	// var err error
	err := plot.Plot(cfg)
	if err != nil {
		dsp.SetPocTaskFailed(taskId, err.Error())
		log.Errorf("plot task %s with cfg %v err %s", taskId, cfg, err)
		return NewDspErr(DSP_TASK_POC_ERROR)
	}

	err = dsp.GenPlotPDPData(taskId, &poc.PlotConfig{
		Sys:        cfg.Sys,
		NumericID:  cfg.NumericID,
		StartNonce: cfg.StartNonce,
		Nonces:     cfg.Nonces,
		Path:       cfg.Path,
	})
	if err != nil {
		log.Errorf("generate new plot file err %s", err)
		return NewDspErr(DSP_TASK_POC_ERROR, err)
	}

	return nil
}

func (this *Endpoint) AddPlotFile(taskId, fileName string, createSector bool) (interface{}, *DspErr) {
	system := runtime.GOOS
	if strings.Contains(system, plot.SYS_WIN) {
		system = plot.SYS_WIN
	} else if !strings.Contains(system, "darwin") {
		system = plot.SYS_LINUX
	}

	acc, derr := this.GetCurrentAccount()
	if derr != nil {
		return nil, derr
	}
	numericID := fmt.Sprintf("%v", utils.WalletAddressToId([]byte(acc.Address)))

	fileBaseName := filepath.Base(fileName)
	if !strings.HasPrefix(fileBaseName, numericID) {
		return nil, &DspErr{Code: DSP_TASK_POC_ERROR, Error: fmt.Errorf("wrong plot file %s", fileName)}
	}

	startNonce, nonce := poc.GetNonceFromName(fileName)
	if nonce == 0 {
		return nil, &DspErr{Code: DSP_TASK_POC_WRONG_NONCE, Error: fmt.Errorf("wrong start nonce or nonces")}
	}

	cfg := &poc.PlotConfig{
		Sys:        system,
		NumericID:  numericID,
		StartNonce: startNonce,
		Nonces:     nonce,
		Path:       config.PlotPath(),
	}

	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp, err := dsp.AddNewPlotFile(taskId, createSector, cfg)
	if err != nil {
		log.Errorf("add new plot file err %s", err)
		return nil, &DspErr{Code: DSP_TASK_POC_ERROR, Error: err}

	}
	return resp, nil
}

func (this *Endpoint) AddPlotFiles(directory string, createSector bool) (interface{}, *DspErr) {
	system := runtime.GOOS
	if strings.Contains(system, plot.SYS_WIN) {
		system = plot.SYS_WIN
	} else if !strings.Contains(system, "darwin") {
		system = plot.SYS_LINUX
	}

	acc, derr := this.GetCurrentAccount()
	if derr != nil {
		return nil, derr
	}
	numericID := fmt.Sprintf("%v", utils.WalletAddressToId([]byte(acc.Address)))

	fileInfos, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, &DspErr{Code: DSP_TASK_POC_ERROR, Error: err}
	}

	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}

	resps := make([]interface{}, 0)

	type errResp struct {
		FileName string
		Err      *DspErr
	}

	log.Infof("fileInfos %v len %v", fileInfos, len(fileInfos))
	for _, fi := range fileInfos {

		if fi.IsDir() {
			continue
		}

		fileName := fi.Name()

		if !strings.HasPrefix(fileName, numericID) {
			continue
		}

		startNonce, nonce := poc.GetNonceFromName(fileName)
		if nonce == 0 {

			resps = append(resps, &errResp{
				FileName: fileName,
				Err:      &DspErr{Code: DSP_TASK_POC_WRONG_NONCE, Error: fmt.Errorf("wrong start nonce or nonces")},
			})
			continue
		}

		cfg := &poc.PlotConfig{
			Sys:        system,
			NumericID:  numericID,
			StartNonce: startNonce,
			Nonces:     nonce,
			Path:       config.PlotPath(),
		}

		resp, err := dsp.AddNewPlotFile("", createSector, cfg)
		if err != nil {
			log.Errorf("add new plot files err %s", err)
			resps = append(resps, &errResp{
				FileName: fileName,
				Err:      &DspErr{Code: DSP_TASK_POC_ERROR, Error: err},
			})
			continue

		}
		resps = append(resps, resp)
	}

	return resps, nil
}

func (this *Endpoint) GetAllProvedPlotFile() (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp, err := dsp.GetAllProvedPlotFile()
	if err != nil {
		return nil, &DspErr{Code: DSP_TASK_POC_ERROR, Error: err}

	}
	return resp, nil
}

func (this *Endpoint) GetAllPocTasks() (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	resp, err := dsp.GetAllPocTasks()
	log.Infof("resp+++ %v", resp)
	if err != nil {
		return nil, &DspErr{Code: DSP_TASK_POC_ERROR, Error: err}

	}
	return resp, nil
}

func (this *Endpoint) DeletePlotFile(fileHash string, gasLimit uint64) (*DeleteFileResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}

	fi, err := dsp.GetFileInfo(fileHash)
	if fi == nil && dsp.IsFileInfoDeleted(err) {
		taskId := dsp.GetPlotTaskId(fileHash)
		if len(taskId) == 0 {
			log.Debugf("file info is deleted: %v, %s, taskId is null", fi, err)
			return nil, nil
		}
		fileName := dsp.GetPlotTaskFileName(fileHash)
		if len(fileName) == 0 {
			log.Debugf("file info is deleted: %v, %s, filename is null", fi, err)
			return nil, nil
		}
		baseFileName := filepath.Base(string(fileName))
		fullFileName := filepath.Join(config.PlotPath(), baseFileName)
		log.Infof("fullFileName %v, basefilename %s %s", fullFileName, fileName, baseFileName)

		cleanTaskErr := dsp.DeletePocTask(taskId)
		if cleanTaskErr != nil {
			return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: cleanTaskErr}
		}
		os.Remove(fullFileName)
		log.Debugf("file info is deleted: %v, %s", fi, err)
		return nil, nil
	}
	if fi != nil && err == nil && fi.FileOwner.ToBase58() == dsp.WalletAddress() {
		baseFileName := filepath.Base(string(fi.FileDesc))
		fullFileName := filepath.Join(config.PlotPath(), baseFileName)
		log.Infof("fullFileName %v, basefilename %s %s", fullFileName, baseFileName)

		taskId := dsp.GetPlotTaskId(fileHash)
		tx, _, _ := dsp.DeleteUploadFilesFromChain([]string{fileHash}, gasLimit)
		resp := &DeleteFileResp{}
		resp.Tx = tx
		resp.FileHash = fileHash

		cleanTaskErr := dsp.DeletePocTask(taskId)
		if cleanTaskErr != nil {
			return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: cleanTaskErr}
		}
		os.Remove(fullFileName)

		return resp, nil
	}
	log.Debugf("fi :%v, err :%v", fi, err)
	return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: err}
}

func (this *Endpoint) BatchDeletePlotFiles(taskIds []string, gasLimit uint64) ([]*DeleteFileResp, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}

	deleteResp := make([]*DeleteFileResp, 0)
	for _, taskId := range taskIds {
		resp := &DeleteFileResp{}
		resp.TaskId = taskId
		fileHash := dsp.GetTaskFileHash(taskId)
		if len(fileHash) == 0 {
			fileName := dsp.GetPlotTaskFileNameById(taskId)
			log.Debugf("poc task %s has null filehash its name %s", taskId, fileName)
			if len(fileName) > 0 {
				baseFileName := filepath.Base(string(fileName))
				fullFileName := filepath.Join(config.PlotPath(), baseFileName)
				os.Remove(fullFileName)
			}
			cleanTaskErr := dsp.DeletePocTask(taskId)
			if cleanTaskErr != nil {
				return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: cleanTaskErr}
			}
			deleteResp = append(deleteResp, resp)
			continue
		}
		fileName := dsp.GetPlotTaskFileName(fileHash)
		fi, err := dsp.GetFileInfo(fileHash)
		if fi == nil && dsp.IsFileInfoDeleted(err) {
			if len(fileName) == 0 {
				cleanTaskErr := dsp.DeletePocTask(taskId)
				if cleanTaskErr != nil {
					return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: cleanTaskErr}
				}
				deleteResp = append(deleteResp, resp)
				continue
			}
			baseFileName := filepath.Base(string(fileName))
			fullFileName := filepath.Join(config.PlotPath(), baseFileName)
			log.Infof("fullFileName %v, basefilename %s %s", fullFileName, fileName, baseFileName)
			cleanTaskErr := dsp.DeletePocTask(taskId)
			if cleanTaskErr != nil {
				os.Remove(fullFileName)
				return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: cleanTaskErr}
			}
			os.Remove(fullFileName)
			deleteResp = append(deleteResp, resp)
			continue
		}

		if fi != nil && err == nil && fi.FileOwner.ToBase58() == dsp.WalletAddress() {
			baseFileName := filepath.Base(string(fi.FileDesc))
			fullFileName := filepath.Join(config.PlotPath(), baseFileName)
			log.Infof("fullFileName %v, basefilename %s %s", fullFileName, baseFileName)

			tx, _, _ := dsp.DeleteUploadFilesFromChain([]string{fileHash}, gasLimit)
			resp.Tx = tx
			resp.FileHash = fileHash
			cleanTaskErr := dsp.DeletePocTask(taskId)
			if cleanTaskErr != nil {
				os.Remove(fullFileName)
				return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: cleanTaskErr}
			}
			os.Remove(fullFileName)
			deleteResp = append(deleteResp, resp)
			continue
		}
		cleanTaskErr := dsp.DeletePocTask(taskId)
		if cleanTaskErr != nil {
			return nil, &DspErr{Code: DSP_DELETE_FILE_FAILED, Error: cleanTaskErr}
		}
		deleteResp = append(deleteResp, resp)
	}

	return deleteResp, nil
}
