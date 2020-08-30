package rest

import (
	"github.com/saveio/edge/dsp"
)

func CreateSector(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	sectorId, ok := cmd["SectorId"].(float64)
	if !ok || sectorId == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	proveLevel, ok := cmd["ProveLevel"].(float64)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	size, ok := cmd["Size"].(float64)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	tx, derr := dsp.DspService.CreateSector(uint64(sectorId), uint64(proveLevel), uint64(size))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}

	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp

}

func DeleteSector(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	sectorId, ok := cmd["SectorId"].(float64)
	if !ok || sectorId == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	tx, derr := dsp.DspService.DeleteSector(uint64(sectorId))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}

	m := make(map[string]interface{}, 0)
	m["Tx"] = tx
	resp["Result"] = m
	return resp
}
func GetSectorInfo(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	sectorId, ok := cmd["SectorId"].(float64)
	if !ok || sectorId == 0 {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	sectorInfo, derr := dsp.DspService.GetSectorInfo(uint64(sectorId))
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	// TODO: may need to update
	resp["Result"] = sectorInfo
	return resp
}
func GetSectorInfosForNode(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	addr, ok := cmd["Addr"].(string)
	if !ok {
		return ResponsePackWithErrMsg(dsp.INVALID_PARAMS, dsp.ErrMaps[dsp.INVALID_PARAMS].Error())
	}

	sectorInfos, derr := dsp.DspService.GetSectorInfosForNode(addr)
	if derr != nil {
		return ResponsePackWithErrMsg(derr.Code, derr.Error.Error())
	}
	resp["Result"] = sectorInfos
	return resp
}
