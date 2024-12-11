package utils

import (
	"time"
	"sync"
	"strconv"
	"fmt"
)

var counter uint32 = 0
var mu sync.Mutex

func GenerateMsgID() *uint32 {
	mu.Lock()
    defer mu.Unlock()

    now := uint32(time.Now().Unix())
    counter++
    uniqueID := now + counter 
    return &uniqueID
}

func StringPtr(s string) *string {
	return &s
}

func Uint32Ptr(u int32) *uint32 {
	value := uint32(u)
	return &value
}


func ParseInt(value string) (int64, error) {
	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing string to int: %v", err)
	}
	return parsedValue, nil
}

func ParseFloat(value string) (float64, error) {
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing string to float: %v", err)
	}
	return parsedValue, nil
}