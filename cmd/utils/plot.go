package utils

func GeneratePlotFile(system, ID, path string, start, nonces uint64) ([]byte, error) {
	ret, dErr := sendRpcRequest("generateplotfile", []interface{}{system, ID, start, nonces, path})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
