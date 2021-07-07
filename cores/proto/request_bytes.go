// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package proto

import "errors"

// RequestBytes 数据流解析
type RequestBytes struct {
	Ver     int8
	Cmd     Command
	SubCmd  Command
	SeqID   uint32
	Content []byte
}

// V 版本
func (req *RequestBytes) V() int8 {
	return req.Ver
}

// Command 主命令号
func (req *RequestBytes) Command() Command {
	return req.Cmd
}

// SubCommand 子命令号
func (req *RequestBytes) SubCommand() Command {
	return req.SubCmd
}

// SID 请求编号
func (req *RequestBytes) SID() uint32 {
	return req.SeqID
}

// Body 请求编号
func (req *RequestBytes) Body() []byte {
	return req.Content
}

// Marshal 封包
func (req *RequestBytes) Marshal() ([]byte, error) {
	if !req.IsValid() {
		return nil, errors.New("Unexpected data")
	}

	var b BytesBuffer
	b.WriteInt8(req.Ver)
	b.WriteUint32(req.SeqID)
	b.WriteInt16(req.Cmd.Int16())
	b.WriteInt16(req.SubCmd.Int16())
	if err := b.WriteBytes(req.Content...); err != nil {
		return nil, err
	}
	return b.Data(), nil
}

// IsValid 检查数据是否合法
func (req *RequestBytes) IsValid() bool {
	if req.SeqID < 1 {
		return false
	}

	if req.Cmd < 1 || req.SubCmd < 1 {
		return false
	}
	return true
}

// Unmarshal 解析rquest 数据
func (req *RequestBytes) Unmarshal(frame []byte) error {
	rd := NewBytesBuffer(frame)
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
	req.Ver = v
	req.SeqID = sid
	req.Cmd = Command(cmd)
	req.SubCmd = Command(subcmd)
	req.Content = rd.Bytes()
	return nil
}
