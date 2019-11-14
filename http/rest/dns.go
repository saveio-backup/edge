package rest

import "github.com/saveio/edge/dsp"

func GetAllDNS(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(dsp.SUCCESS)
	if dsp.DspService == nil {
		return ResponsePackWithErrMsg(dsp.NO_ACCOUNT, dsp.ErrMaps[dsp.NO_ACCOUNT].Error())
	}
	list := make([]map[string]string, 0, len(dsp.DspService.Dsp.GetAllOnlineDNS()))
	for addr, host := range dsp.DspService.Dsp.GetAllOnlineDNS() {
		m := make(map[string]string)
		m["WalletAddr"] = addr
		m["HostAddr"] = host
		list = append(list, m)
	}
	resp["Result"] = list
	return resp
}
