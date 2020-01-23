package dsp

import (
	"os"

	"github.com/saveio/edge/common/config"
	"github.com/saveio/themis/common/log"
)

type ConfigResponse struct {
	BaseDir      string
	DownloadPath string
	LogDirName   string
	LogLevel     int
	LogMaxSize   int
	BlockConfirm int
	SeedInterval int
}

func (this *Endpoint) GetConfigs() *ConfigResponse {
	newResp := &ConfigResponse{
		BaseDir:      config.Parameters.BaseConfig.BaseDir,
		DownloadPath: config.FsFileRootPath(),
		LogDirName:   config.Parameters.BaseConfig.LogPath,
		LogLevel:     config.Parameters.BaseConfig.LogLevel,
		LogMaxSize:   int(config.Parameters.BaseConfig.LogMaxSize),
		BlockConfirm: int(config.Parameters.BaseConfig.BlockConfirm),
		SeedInterval: config.Parameters.BaseConfig.SeedInterval,
	}
	return newResp
}

func (this *Endpoint) SetConfigs(fields map[string]interface{}) (*ConfigResponse, *DspErr) {
	newResp := &ConfigResponse{
		BaseDir:      config.Parameters.BaseConfig.BaseDir,
		DownloadPath: config.FsFileRootPath(),
		LogDirName:   config.Parameters.BaseConfig.LogPath,
		LogLevel:     config.Parameters.BaseConfig.LogLevel,
		LogMaxSize:   int(config.Parameters.BaseConfig.LogMaxSize),
		BlockConfirm: int(config.Parameters.BaseConfig.BlockConfirm),
		SeedInterval: config.Parameters.BaseConfig.SeedInterval,
	}
	for key, value := range fields {
		switch key {
		case "BaseDir":
			newPath, ok := value.(string)
			if !ok {
				log.Debugf("unsupport type %s %T", key, value)
				continue
			}
			// TODO: need reboot now, support no reboot future
			config.Parameters.BaseConfig.BaseDir = newPath
			newResp.BaseDir = newPath
		case "LogDirName":
			newPath, ok := value.(string)
			if !ok {
				log.Debugf("unsupport type %s %T", key, value)
				continue
			}
			config.Parameters.BaseConfig.LogPath = newPath
			newResp.LogDirName = newPath
		case "LogLevel":
			newLevel, ok := value.(float64)
			if !ok {
				log.Debugf("unsupport type %s %T", key, value)
				continue
			}
			config.Parameters.BaseConfig.LogLevel = int(newLevel)
			newResp.LogLevel = int(newLevel)
		case "BlockConfirm":
			newConfirm, ok := value.(float64)
			if !ok {
				log.Debugf("unsupport type %s %T", key, value)
				continue
			}
			config.Parameters.BaseConfig.BlockConfirm = uint32(newConfirm)
			newResp.BlockConfirm = int(newConfirm)
		case "SeedInterval":
			newInterval, ok := value.(float64)
			if !ok {
				log.Debugf("unsupport type %s %T", key, value)
				continue
			}
			config.Parameters.BaseConfig.SeedInterval = int(newInterval)
			newResp.SeedInterval = int(newInterval)
		case "DownloadPath":
			newPath, ok := value.(string)
			if !ok {
				log.Debugf("unsupport type %s %T", key, value)
				continue
			}
			if this == nil {
				continue
			}
			dsp := this.getDsp()
			if dsp == nil {
				continue
			}
			oldPath := config.Parameters.FsConfig.FsFileRoot
			config.Parameters.FsConfig.FsFileRoot = newPath
			if err := dsp.UpdateConfig("FsFileRoot", config.FsFileRootPath()); err != nil {
				log.Errorf("update config err %s", err)
				config.Parameters.FsConfig.FsFileRoot = oldPath
				newResp.DownloadPath = oldPath
			} else {
				newResp.DownloadPath = newPath
			}
			newResp.DownloadPath = newPath
		case "LogMaxSize":
			newLogSize, ok := value.(float64)
			if !ok {
				log.Debugf("unsupport type %s %T", key, value)
				continue
			}
			config.Parameters.BaseConfig.LogMaxSize = uint64(newLogSize)
			newResp.LogMaxSize = int(newLogSize)
		}
	}
	err := config.Save()
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	return newResp, nil
}

type DBType int

const (
	DBTypeAll DBType = iota
	DBTypeEdge
	DBTypeDsp
	DBTypeChannel
	DBTypeFs
)

func (this *Endpoint) RemoveDBDir(dbType DBType) *DspErr {
	switch dbType {
	case DBTypeAll:
		if err := os.RemoveAll(config.DspDBPath()); err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
		if err := os.RemoveAll(config.ChannelDBPath()); err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
		if err := os.RemoveAll(config.ClientSqliteDBPath()); err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
		if err := os.RemoveAll(config.FsRepoRootPath()); err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
	case DBTypeEdge:
		if err := os.RemoveAll(config.ClientSqliteDBPath()); err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
	case DBTypeDsp:
		if err := os.RemoveAll(config.DspDBPath()); err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
	case DBTypeChannel:
		if err := os.RemoveAll(config.ChannelDBPath()); err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
	case DBTypeFs:
		if err := os.RemoveAll(config.FsRepoRootPath()); err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
	}
	return nil

}
