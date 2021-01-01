package go_internals

import (
	"bytes"
	"runtime"
	"strconv"
	"unsafe"
)

func GetG() *GStub {
	return (*GStub)(unsafe.Pointer(getg()))
}

func GetGID() int64 {
	return GetG().goid
}

func GetGIDSlowSafe() int64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseInt(string(b), 10, 64)
	return n
}

func AssignGcAssistBytes(credit int64) {
	GetG().gcAssistBytes = credit
}

func GetGcAssistBytes() int64 {
	return GetG().gcAssistBytes
}