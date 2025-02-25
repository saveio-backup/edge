package rest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
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
	log.Infof("GeneratePlotFile cmd %v", cmd)

	system, ok := cmd["System"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, "invalid param system")
	}

	numericID, _ := cmd["NumericID"].(string)
	if len(numericID) == 0 {
		acc, err := dsp.DspService.GetCurrentAccount()
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, "invalid params numericID")
		}
		numericID = fmt.Sprintf("%v", utils.WalletAddressToId([]byte(acc.Address)))
	}

	startNonce, _ := utils.ToUint64(cmd["StartNonce"])
	nonces, _ := utils.ToUint64(cmd["Nonces"])

	path, ok := cmd["Path"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, "invalid params path")
	}
	if len(path) == 0 || path == config.DEFAULT_PLOT_PATH {
		path = config.PlotPath()
	}
	if path != config.PlotPath() {
		config.Parameters.BaseConfig.PlotPath = path
		log.Infof("save config %s", config.Parameters.BaseConfig.PlotPath)
		if err := config.Save(); err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, fmt.Sprintf("save config err %s", err))
		}
	}

	log.Infof("path %s, plot path %s", path, config.PlotPath())
	size, _ := utils.ToUint64(cmd["Size"])
	num, _ := utils.ToUint64(cmd["Num"])
	ms := make([]interface{}, 0)
	if size == 0 {

		if nonces%8 != 0 || nonces == 0 {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, "invalid params nonces, nonces should be an integer multiple of 2048")
		}

		cfg := &plot.PlotConfig{
			Sys:        system,
			NumericID:  numericID,
			StartNonce: uint64(startNonce),
			Nonces:     uint64(nonces),
			Path:       path,
		}
		cfgData, _ := json.Marshal(cfg)
		log.Infof("plot config cfg with no size %s", cfgData)
		taskId, err := dsp.DspService.NewPocTask(cfg)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.ErrorMsg())
		}

		go func() {
			if err := dsp.DspService.GenPlotPDPData(taskId, cfg); err != nil {
				log.Errorf("generate plot pdp data err %s", err)
			}
		}()

		m := make(map[string]interface{})
		m["TaskId"] = taskId
		m["NumericID"] = numericID
		m["StartNonce"] = startNonce
		m["Nonces"] = nonces
		m["Path"] = path
		m["PlotFileName"] = plot.GetPlotFileName(cfg)

		ms = append(ms, m)
		resp["Result"] = ms
		return resp
	}
	var err error
	startNonce, err = plot.GetMinStartNonce(numericID, path)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
	}
	nonces = size / plot.DEFAULT_PLOT_SIZEKB

	if nonces%8 != 0 || nonces == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, "invalid params size, size should be an integer multiple of 2048")
	}

	for i := uint64(0); i < uint64(num); i++ {

		cfg := &plot.PlotConfig{
			Sys:        system,
			NumericID:  numericID,
			StartNonce: uint64(startNonce),
			Nonces:     uint64(nonces),
			Path:       path,
		}
		cfgData, _ := json.Marshal(cfg)
		log.Infof("plot config cfg %s", cfgData)

		taskId, err := dsp.DspService.NewPocTask(cfg)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.ErrorMsg())
		}
		go func() {
			if err := dsp.DspService.GenPlotPDPData(taskId, cfg); err != nil {
				log.Errorf("generate plot pdp data err %s", err)
			}
		}()
		m := make(map[string]interface{})
		m["TaskId"] = taskId
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

	path := config.PlotPath()
	log.Debugf("GetAllPlotFiles path %s", path)
	pathHex, _ := cmd["Path"].(string)
	if len(pathHex) > 0 {
		pathBuf, err := hex.DecodeString(pathHex)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
		path = string(pathBuf)
	}

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
	taskId, _ := cmd["TaskId"].(string)

	fileName, ok := cmd["FileName"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	createSector, ok := cmd["CreateSector"].(bool)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	result, err := dsp.DspService.AddPlotFile(taskId, fileName, createSector)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error.Error()), "insufficient balance") {
			return ResponsePackWithErrMsg(dsp.INSUFFICIENT_BALANCE, err.Error.Error())
		}
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error.Error())
	}

	resp["Result"] = result
	return resp
}

func AddPlotFolderToMine(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	directory, ok := cmd["Directory"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	createSector, ok := cmd["CreateSector"].(bool)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	result, err := dsp.DspService.AddPlotFiles(directory, createSector)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error.Error()), "insufficient balance") {
			return ResponsePackWithErrMsg(dsp.INSUFFICIENT_BALANCE, err.Error.Error())
		}
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error.Error())
	}

	resp["Result"] = result
	return resp
}

func GetAllProvedPlotFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	result, err := dsp.DspService.GetAllProvedPlotFile()
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error.Error())
	}

	resp["Result"] = result
	return resp
}

func GetAllPocTasks(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)

	result, err := dsp.DspService.GetAllPocTasks()
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error.Error())
	}

	resp["Result"] = result
	return resp
}

func DeletePlotFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	fileHash, ok := cmd["Hash"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if len(fileHash) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	gl, _ := cmd["GasLimit"].(string)
	gasLimit := uint64(0)
	if len(gl) > 0 {
		var err error
		gasLimit, err = strconv.ParseUint(gl, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}
	result, err := dsp.DspService.DeletePlotFile(fileHash, gasLimit)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error.Error())
	}

	resp["Result"] = result
	return resp
}

func DeletePlotFiles(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	taskIds, ok := cmd["TaskIds"].([]interface{})
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	if len(taskIds) == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	allTaskIds := make([]string, 0, len(taskIds))
	for _, str := range taskIds {
		taskId, ok := str.(string)
		if !ok {
			continue
		}
		allTaskIds = append(allTaskIds, taskId)
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}
	gl, _ := cmd["GasLimit"].(string)
	gasLimit := uint64(0)
	if len(gl) > 0 {
		var err error
		gasLimit, err = strconv.ParseUint(gl, 10, 64)
		if err != nil {
			return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, err.Error())
		}
	}
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	if checkErr := dsp.DspService.CheckPassword(password); checkErr != nil {
		return ResponsePackWithErrMsg(checkErr.Code, checkErr.Error.Error())
	}
	result, err := dsp.DspService.BatchDeletePlotFiles(allTaskIds, gasLimit)
	if err != nil {
		return ResponsePackWithErrMsg(dsp.INTERNAL_ERROR, err.Error.Error())
	}

	resp["Result"] = result
	return resp
}
