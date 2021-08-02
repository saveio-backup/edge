package utils

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

func AddPlotFile(path string, createSector bool) ([]byte, error) {
	ret, dErr := sendRpcRequest("addplotfile", []interface{}{path, createSector})
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
