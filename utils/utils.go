/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2018-11-27
 */
package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/saveio/themis/common/log"
	"github.com/urfave/cli"
)

//GetFlagName deal with short flag, and return the flag name whether flag name have short name
func GetFlagName(flag cli.Flag) string {
	name := flag.GetName()
	if name == "" {
		return ""
	}

	return strings.TrimSpace(strings.Split(name, ",")[0])
}

func ConvertStructToMap(e reflect.Value) map[string]interface{} {
	m := make(map[string]interface{})
	for i := 0; i < e.NumField(); i++ {
		name := e.Type().Field(i).Name
		vType := e.Type().Field(i).Type
		val := e.Field(i).Interface()
		// fmt.Printf("type %v\n", vType)
		if fmt.Sprintf("%v", vType) == "[]uint8" {
			m[fmt.Sprintf("%v", name)] = fmt.Sprintf("%s", val)
		} else if fmt.Sprintf("%v", vType) == "[][]uint8" {
			valBytes := val.([][]byte)
			newVal := make([]string, 0, len(valBytes))
			for _, v := range valBytes {
				newVal = append(newVal, string(v))
			}
			m[fmt.Sprintf("%v", name)] = newVal
		} else {
			m[fmt.Sprintf("%v", name)] = val
		}
	}
	return m
}

func Sha256HexStr(str string) string {
	pwdBuf := sha256.Sum256([]byte(str))
	pwdHash := hex.EncodeToString(pwdBuf[:])
	return pwdHash
}

func StringToUint64(value interface{}) uint64 {
	str, _ := value.(string)
	if len(str) == 0 {
		return 0
	}
	intVal, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0
	}
	return intVal
}

type FileInfos []os.FileInfo

func (s FileInfos) Len() int {
	return len(s)
}

func (s FileInfos) Less(i, j int) bool {
	return s[i].ModTime().Unix() < s[j].ModTime().Unix()
}

func (s FileInfos) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func CleanOldestLogs(path string, maxSizeInKB uint64) {
	var size uint64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += uint64(info.Size())
		}
		return nil
	})
	log.Debugf("size %v, config size %v", size, maxSizeInKB)

	if size < maxSizeInKB*1024 {
		// return
	}
	nowTimestamp := time.Now().Unix()

	fis := make(FileInfos, 0)
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		log.Debugf("name: %s now %d time: %d", filepath.Join(path, info.Name()), nowTimestamp, info.ModTime().Unix())
		if !info.IsDir() && nowTimestamp > info.ModTime().Unix() {
			fis = append(fis, info)
		}
		return nil
	})
	sort.Sort(fis)
	for _, info := range fis {
		log.Debugf("delete name: %s time: %d", filepath.Join(path, info.Name()), info.ModTime().Unix())
		os.Remove(filepath.Join(path, info.Name()))
		size -= uint64(info.Size())
		if size < maxSizeInKB*1024 {
			break
		}
	}
}
