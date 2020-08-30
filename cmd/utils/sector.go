package utils

func CreateSector(sectorId uint64, proveLevel uint64, size uint64) ([]byte, error) {
	ret, dErr := sendRpcRequest("createsector", []interface{}{sectorId, proveLevel, size})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func DeleteSector(sectorId uint64) ([]byte, error) {
	ret, dErr := sendRpcRequest("deletesector", []interface{}{sectorId})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func GetSectorInfo(sectorId uint64) ([]byte, error) {
	ret, dErr := sendRpcRequest("getsectorinfo", []interface{}{sectorId})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}

func GetSectorInfosForNode(addr string) ([]byte, error) {
	ret, dErr := sendRpcRequest("getsectorinfosfornode", []interface{}{addr})
	if dErr != nil {
		return nil, dErr.Error
	}
	return ret, nil
}
