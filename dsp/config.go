package dsp

import (
	"errors"

	"github.com/saveio/edge/common/config"
)

func (this *Endpoint) SetConfig(key string, value interface{}) error {
	switch key {
	case "DownloadPath":
		newPath, ok := value.(string)
		if !ok {
			return errors.New("set config invalid value type")
		}
		config.Parameters.FsConfig.FsFileRoot = newPath
		err := this.Dsp.UpdateConfig("FsFileRoot", config.FsFileRootPath())
		if err != nil {
			return err
		}
	}
	return config.Save()
}
