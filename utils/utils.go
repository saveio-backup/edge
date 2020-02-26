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
	"reflect"
	"strconv"
	"strings"

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
