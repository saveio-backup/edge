package dsp

import (
	"strings"

	http "github.com/saveio/edge/http/common"
)

func FileNameMatchType(fileType http.DSP_FILE_LIST_TYPE, fileName string) bool {
	switch fileType {
	case http.DSP_FILE_LIST_TYPE_ALL:
		return true
	case http.DSP_FILE_LIST_TYPE_IMAGE:
		if strings.Index(fileName, "jpg") == -1 &&
			strings.Index(fileName, "gif") == -1 &&
			strings.Index(fileName, "svg") == -1 &&
			strings.Index(fileName, "png") == -1 &&
			strings.Index(fileName, "jpeg") == -1 {
			return false
		}
	case http.DSP_FILE_LIST_TYPE_DOC:
		if strings.Index(fileName, "doc") == -1 &&
			strings.Index(fileName, "txt") == -1 &&
			strings.Index(fileName, "dat") == -1 &&
			strings.Index(fileName, "docx") == -1 &&
			strings.Index(fileName, "md") == -1 &&
			strings.Index(fileName, "pdf") == -1 &&
			strings.Index(fileName, "xlx") == -1 {
			return false
		}
	case http.DSP_FILE_LIST_TYPE_VIDEO:
		if strings.Index(fileName, "mp4") == -1 &&
			strings.Index(fileName, "mov") == -1 &&
			strings.Index(fileName, "rmvb") == -1 &&
			strings.Index(fileName, "avi") == -1 &&
			strings.Index(fileName, "rm") == -1 {
			return false
		}
	case http.DSP_FILE_LIST_TYPE_MUSIC:
		if strings.Index(fileName, "mp3") == -1 {
			return false
		}
	}
	return false
}
