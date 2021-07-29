package dsp

import (
	"fmt"
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
	log.Infof("system %s", system)
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
		return nil, &DspErr{Code: DSP_TASK_POC_ERROR, Error: err}

	}
	return resp, nil
}
