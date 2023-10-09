package utils

import (
	"fmt"
	"strconv"
)

func StringToFloat32(value string) (float32, error) {
	floatValue, err := strconv.ParseFloat(value, 32)
	if err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}
	return float32(floatValue), nil
}
