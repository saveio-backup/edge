package dsp

import (
	"errors"
	"strconv"
	"strings"
)

func FileNameMatchType(fileType DspFileListType, fileName string) bool {
	switch fileType {
	case DspFileListTypeAll:
		return true
	case DspFileListTypeImage:
		if strings.Index(fileName, ".jpg") != -1 ||
			strings.Index(fileName, ".gif") != -1 ||
			strings.Index(fileName, ".svg") != -1 ||
			strings.Index(fileName, ".png") != -1 ||
			strings.Index(fileName, ".jpeg") != -1 {
			return true
		}
	case DspFileListTypeDoc:
		if strings.Index(fileName, ".doc") != -1 ||
			strings.Index(fileName, ".txt") != -1 ||
			strings.Index(fileName, ".dat") != -1 ||
			strings.Index(fileName, ".docx") != -1 ||
			strings.Index(fileName, ".md") != -1 ||
			strings.Index(fileName, ".pdf") != -1 ||
			strings.Index(fileName, ".xlx") != -1 {
			return true
		}
	case DspFileListTypeVideo:
		if strings.Index(fileName, ".mp4") != -1 ||
			strings.Index(fileName, ".mov") != -1 ||
			strings.Index(fileName, ".rmvb") != -1 ||
			strings.Index(fileName, ".avi") != -1 ||
			strings.Index(fileName, ".rm") != -1 {
			return true
		}
	case DspFileListTypeMusic:
		if strings.Index(fileName, ".mp3") != -1 {
			return true
		}
	}
	return false
}

func RequiredStrToUint64(in interface{}) (uint64, error) {
	str, ok := in.(string)
	if !ok {
		return 0, errors.New("value is required")
	}
	if len(str) == 0 {
		return 0, errors.New("value is required")
	}
	ret, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func OptionStrToUint64(in interface{}) (uint64, error) {
	str, ok := in.(string)
	if !ok {
		return 0, nil
	}
	if len(str) == 0 {
		return 0, nil
	}
	ret, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

// OptionStrToFloat64. convert a optional string to float64, if the value is not found, use default value.
// return error if parse the string value failed
func OptionStrToFloat64(in interface{}, def float64) (float64, error) {
	str, ok := in.(string)
	if !ok {
		return def, nil
	}
	if len(str) == 0 {
		return def, nil
	}
	ret, err := strconv.ParseFloat(str, 10)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func ParseContractError(err error) *DspErr {
	if strings.Contains(err.Error(), "[FS UserSpace]") {
		if strings.Contains(err.Error(), "FsManageUserSpace can't revoke, there exists files") {
			return &DspErr{Code: FS_CANT_REVOKE_OF_EXISTS_FILE, Error: err}
		}
		if strings.Contains(err.Error(), "FsManageUserSpace  no user space to revoke") {
			return &DspErr{Code: FS_NO_USER_SPACE_TO_REVOKE, Error: err}
		}
		if strings.Contains(err.Error(), "FsManageUserSpace block count too small at first purchase user space") {
			return &DspErr{Code: FS_USER_SPACE_SECOND_TOO_SMALL, Error: err}
		}
		if strings.Contains(err.Error(), "FsManageUserSpace can't revoke other user space") {
			return &DspErr{Code: FS_USER_SPACE_PERMISSION_DENIED, Error: err}
		}
		if strings.Contains(err.Error(), "AppCallTransfer, transfer error!") {
			return &DspErr{Code: INSUFFICIENT_BALANCE, Error: err}
		}
	}
	return &DspErr{Code: CONTRACT_ERROR, Error: err}
}
