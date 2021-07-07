// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package proto

// Command  定义命令类型
type Command int16

// Int16 转换
func (c Command) Int16() int16 {
	return int16(c)
}

// Int32 转换
func (c Command) Int32() int32 {
	return int32(c)
}
