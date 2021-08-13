package dsp

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/saveio/dsp-go-sdk/task/poc"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/utils"
	"github.com/saveio/edge/utils/plot"
	"github.com/saveio/themis/common/log"
)

func (this *Endpoint) AddPlotFile(fileName string, createSector bool) (interface{}, *DspErr) {
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
	resp, err := dsp.AddNewPlotFile("", createSector, cfg)
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
