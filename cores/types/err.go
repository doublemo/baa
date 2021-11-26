// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package types

import (
	"fmt"
)

// ErrCode 错误码
type ErrCode struct {
	// code 错误码
	code int32

	// message 错误信息
	message string
}

// Code 错误码
func (e ErrCode) Code() int32 {
	return e.code
}

// Error 错误信息
func (e ErrCode) Error() string {
	return e.message
}

// Bytes 获取错误信息
func (e ErrCode) Bytes() []byte {
	return []byte(e.message)
}

// ToError 转换为error类型
func (e ErrCode) ToError() error {
	return fmt.Errorf("<%d> %s", e.code, e.Error())
}

// NewErrCode 创建错误码
func NewErrCode(code int32, message string) *ErrCode {
	return &ErrCode{code, message}
}
