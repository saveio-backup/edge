package utils

import "github.com/saveio/themis/common/log"

func GeneratePlotFile(system, ID, path string, start, nonces uint64) ([]byte, error) {
	ret, dErr := sendRpcRequest("generateplotfile", []interface{}{system, ID, start, nonces, path})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func GetAllPlotFile(path string) ([]byte, error) {
	ret, dErr := sendRpcRequest("getallplotfiles", []interface{}{path})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func AddPlotFile(taskId, path string, createSector bool) ([]byte, error) {
	ret, dErr := sendRpcRequest("addplotfile", []interface{}{taskId, path, createSector})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func AddPlotFiles(directory string, createSector bool) ([]byte, error) {
	ret, dErr := sendRpcRequest("addplotfiles", []interface{}{directory, createSector})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func GetAllProvedPlotFile() ([]byte, error) {
	ret, dErr := sendRpcRequest("getallprovedplotfile", []interface{}{})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func GetAllPlotTasks() ([]byte, error) {
	ret, dErr := sendRpcRequest("getallpoctasks", []interface{}{})
	log.Infof("ret +++ %v, dErr %v", ret, dErr)
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
