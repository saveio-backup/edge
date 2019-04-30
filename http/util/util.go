package util

import (
	"errors"
	"strconv"
)

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
