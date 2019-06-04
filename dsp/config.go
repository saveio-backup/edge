package dsp

import (
	"github.com/saveio/edge/common/config"
)

func (this *Endpoint) SetConfig(key string, value interface{}) *DspErr {
	switch key {
	case "DownloadPath":
		newPath, ok := value.(string)
		if !ok {
			return &DspErr{Code: INVALID_PARAMS, Error: ErrMaps[INVALID_PARAMS]}
		}
		config.Parameters.FsConfig.FsFileRoot = newPath
		err := this.Dsp.UpdateConfig("FsFileRoot", config.FsFileRootPath())
		if err != nil {
			return &DspErr{Code: DSP_UPDATE_CONFIG_FAILED, Error: err}
		}
	}
	err := config.Save()
	if err != nil {
		return &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	return nil
}
