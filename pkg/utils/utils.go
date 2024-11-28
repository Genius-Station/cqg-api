package utils

import (
	"time"
	"sync"
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
