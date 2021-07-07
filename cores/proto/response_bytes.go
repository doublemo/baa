// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package proto

import (
	"errors"
)

// ResponseBytes 数据流解析
type ResponseBytes struct {
	Code    int16
	Ver     int8
	Cmd     Command
	SubCmd  Command
	SID     uint32
	Content []byte
}

// StatusCode 版本
func (resp *ResponseBytes) StatusCode() int16 {
	return resp.Code
}

// V 版本
func (resp *ResponseBytes) V() int8 {
	return resp.Ver
}

// Command 主命令号
func (resp *ResponseBytes) Command() Command {
	return resp.Cmd
}

// SubCommand 子命令号
func (resp *ResponseBytes) SubCommand() Command {
	return resp.SubCmd
}

// SeqID 请求编号
func (resp *ResponseBytes) SeqID() uint32 {
	return resp.SID
}

// Body 请求编号
func (resp *ResponseBytes) Body() []byte {
	return resp.Content
}

// Marshal 封包
func (resp *ResponseBytes) Marshal() ([]byte, error) {
	if !resp.IsValid() {
		return nil, errors.New("Unexpected data")
	}

	var b BytesBuffer
	b.WriteInt16(resp.Code)
	b.WriteInt8(resp.Ver)
	b.WriteUint32(resp.SID)
	b.WriteInt16(resp.Cmd.Int16())
	b.WriteInt16(resp.SubCmd.Int16())
	if err := b.WriteBytes(resp.Content...); err != nil {
		return nil, err
	}
	return b.Data(), nil
}

// IsValid 检查数据是否合法
func (resp *ResponseBytes) IsValid() bool {
	if resp.SID < 1 {
		return false
	}
	return true
}

// Unmarshal 解析rquest 数据
func (resp *ResponseBytes) Unmarshal(frame []byte) error {
	rd := NewBytesBuffer(frame)
	code, err := rd.ReadInt16()
	if err != nil {
		return err
	}

	v, err := rd.ReadInt8()
	if err != nil {
		return err
	}

	sid, err := rd.ReadUint32()
	if err != nil {
		return err
	}

	cmd, err := rd.ReadInt16()
	if err != nil {
		return err
	}

	subcmd, err := rd.ReadInt16()
	if err != nil {
		return err
	}

	resp.Code = code
	resp.Ver = v
	resp.SID = sid
	resp.Cmd = Command(cmd)
	resp.SubCmd = Command(subcmd)
	resp.Content = rd.Bytes()
	return nil
}

