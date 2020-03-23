package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// FileExisted checks whether filename exists in filesystem
func FileExisted(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func GetJsonObjectFromFile(filePath string, jsonObject interface{}) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	// Remove the UTF-8 Byte Order Mark
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(data, jsonObject)
	if err != nil {
		return fmt.Errorf("json.Unmarshal %s error:%s", data, err)
	}
	return nil
}

// IsAbsPath. check if the file path is a absolute path
func IsAbsPath(filePath string) bool {
	return path.IsAbs(filePath) || path.IsAbs(strings.Replace(filePath, "\\", "/", -1))
}
