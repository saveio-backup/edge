package rest

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/saveio/dsp-go-sdk/task/poc"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/utils"
	"github.com/saveio/edge/utils/plot"
	"github.com/saveio/themis/common/log"
)

func GeneratePlotFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	system, ok := cmd["System"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	numericID, _ := cmd["NumericID"].(string)
	if len(numericID) == 0 {
		acc, err := dsp.DspService.GetCurrentAccount()
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error.Error())
		}
		numericID = fmt.Sprintf("%v", utils.WalletAddressToId([]byte(acc.Address)))
	}

	startNonce, _ := utils.ToUint64(cmd["StartNonce"])
	nonces, _ := utils.ToUint64(cmd["Nonces"])

	path, ok := cmd["Path"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if len(path) > 0 {
		config.Parameters.BaseConfig.PlotPath = path
		if err := config.Save(); err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
		}
	}

	size, _ := utils.ToUint64(cmd["Size"])
	num, _ := utils.ToUint64(cmd["Num"])

	if size == 0 {
		cfg := &plot.PlotConfig{
			Sys:        system,
			NumericID:  numericID,
			StartNonce: uint64(startNonce),
			Nonces:     uint64(nonces),
			Path:       path,
		}
		err := plot.Plot(cfg)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error())
		}

		m := make(map[string]interface{})
		m["NumericID"] = numericID
		m["StartNonce"] = startNonce
		m["Nonces"] = nonces
		m["Path"] = path
		m["PlotFileName"] = plot.GetPlotFileName(cfg)

	}
	var err error
	startNonce, err = plot.GetMinStartNonce(numericID, path)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
	}
	nonces = size / plot.DEFAULT_PLOT_SIZEKB

	ms := make([]interface{}, 0)
	for i := uint64(0); i < uint64(num); i++ {

		cfg := &plot.PlotConfig{
			Sys:        system,
			NumericID:  numericID,
			StartNonce: uint64(startNonce),
			Nonces:     uint64(nonces),
			Path:       path,
		}

		go func() {
			err := plot.Plot(cfg)
			if err != nil {
				log.Errorf("create plot err %s", err)
			}
		}()

		m := make(map[string]interface{})
		m["NumericID"] = numericID
		m["StartNonce"] = startNonce
		m["Nonces"] = nonces
		m["Path"] = path
		m["PlotFileName"] = plot.GetPlotFileName(cfg)

		ms = append(ms, m)

		startNonce += nonces
	}
	resp["Result"] = ms
	return resp

}

type PlotFile struct {
	Name       string
	Path       string
	NumericID  string
	Nonce      uint64
	StartNonce uint64
	Size       uint64
	SplitSize  uint64
}

func GetAllPlotFiles(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	acc, derr := dsp.DspService.GetCurrentAccount()
	if derr != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, derr.Error.Error())
	}
	numericID := fmt.Sprintf("%v", utils.WalletAddressToId([]byte(acc.Address)))

	pathHex, ok := cmd["Path"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	pathBuf, err := hex.DecodeString(pathHex)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
	}
	path := string(pathBuf)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
	}

	list := make([]PlotFile, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasPrefix(file.Name(), numericID) {
			continue
		}

		file := PlotFile{
			Name:      file.Name(),
			Path:      path,
			NumericID: numericID,
			SplitSize: uint64(file.Size()),
		}

		startNonce, nonce := poc.GetNonceFromName(file.Name)
		if nonce == 0 {
			continue
		}

		file.Nonce = nonce
		file.StartNonce = startNonce
		file.Size = nonce * plot.DEFAULT_PLOT_SIZEKB

		list = append(list, file)
	}

	resp["Result"] = list
	return resp
}

func AddPlotFileToMine(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	fileName, ok := cmd["FileName"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	createSector, ok := cmd["CreateSector"].(bool)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	result, err := dsp.DspService.AddPlotFile(fileName, createSector)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error.Error())
	}

	resp["Result"] = result
	return resp
}
