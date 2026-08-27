package vrpc

import (
	"sync"
)

var (
	_errorCodes  sync.Map
	_successCode sync.Map
)

func SetErrorCodes(serviceName string, errorCodes map[string]int) {
	_errorCodes.Store(serviceName, errorCodes)
}

func SetSuccessCodes(serviceName string, code int) {
	_successCode.Store(serviceName, code)
}

func GetErrorCodes(serviceName string) map[string]int {
	val, ok := _errorCodes.Load(serviceName)
	if !ok {
		return nil
	}
	return val.(map[string]int)
}

func GetSuccessCode(serviceName string) int {
	val, ok := _successCode.Load(serviceName)
	if !ok {
		return 0
	}
	return val.(int)
}

func MapErrorCode(serviceName string, err error) int {
	errorCodes := GetErrorCodes(serviceName)
	if errorCodes == nil {
		return 500
	}
	val, ok := errorCodes[err.Error()]
	if !ok {
		return 500
	}
	return val
}

func MapSuccessCode(serviceName string) int {
	successCode := GetSuccessCode(serviceName)
	if successCode == 0 {
		return 200
	}
	return successCode
}
