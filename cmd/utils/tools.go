package utils

import "math"

func ParserByteSizeToKB(value interface{}) float64 {
	valueF, ok := value.(float64)
	if !ok {
		return 0
	}
	return valueF / 1024
}

func ParseAssets(value interface{}) float64 {
	valueF, ok := value.(float64)
	if !ok {
		return 0
	}
	return valueF / math.Pow10(9)
}
