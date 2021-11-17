package proto

import (
	"strconv"

	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
)

// NewResponseBytes 创建Response Bytes
func NewResponseBytes(cmd coresproto.Command, resp *corespb.Response) *coresproto.ResponseBytes {
	w := &coresproto.ResponseBytes{}
	if resp == nil {
		return w
	}

	w.Cmd = cmd
	w.SubCmd = coresproto.Command(resp.Command)
	w.SID = 1
	w.Code = 0
	w.Ver = 1

	if seqno, ok := resp.Header["seqno"]; ok {
		if i, err := strconv.ParseUint(seqno, 10, 32); err == nil {
			w.SID = uint32(i)
		}
	}

	switch payload := resp.Payload.(type) {
	case *corespb.Response_Content:
		w.Content = payload.Content
	case *corespb.Response_Error:
		w.Code = int32(payload.Error.Code)
		w.Content = []byte(payload.Error.Message)
	}
	return w
}
