package snid

import (
	"sync"
	"time"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/kits/snid/dao"
	"github.com/doublemo/baa/kits/snid/errcode"
	"github.com/doublemo/baa/kits/snid/proto/pb"
	grpcproto "github.com/golang/protobuf/proto"
)

type (
	snidConfig struct {
		// MachineID 机器码
		MachineID uint64 `alias:"machineid" default:"1"`

		// TimeBits 时间位
		TimeBits uint `alias:"timebits" default:"41"`

		// SeqBits 信息位
		SeqBits uint `alias:"seqbits" default:"12"`

		// MachineBits 机器码位
		MachineBits uint `alias:"machinebits" default:"10"`

		// StartTime UUID产生的开始时间，为纳秒
		StartTime int64 `alias:"starttime" default:"1551839574000000000"`
	}

	snid struct {
		// timeBits 时间位
		timeBits uint

		// seqBits 消息计数位
		seqBits uint

		// machineBits 机器码位
		machineBits uint

		// timeMask 时间
		timeMask uint64

		// seqMask 消息
		seqMask uint64

		// machineIDMask 机器码
		machineIDMask uint64

		// machineID 机器码ID
		machineID uint64

		// startTimeNano 启始时间纳秒
		startTimeNano int64

		lastts int64
		seq    uint64
		mutex  sync.RWMutex
	}
)

func (sn *snid) Serve(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SNID_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{Command: req.Command}
	if req.Header != nil {
		w.Header = req.Header
	}

	num := frame.N
	if num < 1 {
		num = 1
	} else if num > 100 {
		return errcode.Bad(w, errcode.ErrMaxIDNumber), nil
	}

	resp := &pb.SNID_Reply{
		Values: make([]uint64, num),
	}

	for i := 0; i < int(num); i++ {
		resp.Values[i] = sn.Read()
	}

	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}

func (sn *snid) Read() uint64 {
	sn.mutex.RLock()
	lastts := sn.lastts
	seq := sn.seq
	timeMask := sn.timeMask
	machineBits := sn.machineBits
	seqBits := sn.seqBits
	machineID := sn.machineID
	sn.mutex.RUnlock()

	t := sn.ts()
	if t < lastts {
		t = sn.wait(lastts)
	} else if lastts == t {
		seq = (seq + 1) & sn.seqMask
		if seq == 0 {
			t = sn.wait(lastts)
		}
	} else {
		seq = 0
	}

	sn.mutex.Lock()
	sn.lastts = t
	sn.seq = seq
	sn.mutex.Unlock()

	var id uint64
	id |= (uint64(t) & timeMask) << (machineBits + seqBits)
	id |= machineID
	id |= seq
	return id
}

func (sn *snid) wait(s int64) int64 {
	t := sn.ts()
	for t < s {
		time.Sleep(time.Duration(s-t) * time.Millisecond)
		t = sn.ts()
	}
	return t
}

func (sn *snid) ts() int64 {
	return (time.Now().UnixNano() - sn.startTimeNano) / int64(time.Millisecond)
}

func newSnid(c snidConfig) *snid {
	if c.MachineID < 1 {
		c.MachineID = 1
	}

	machineIDMask := uint64(^(int64(-1) << c.MachineBits))
	return &snid{
		timeBits:      c.TimeBits,
		seqBits:       c.SeqBits,
		machineBits:   c.MachineBits,
		timeMask:      uint64(^(int64(-1) << c.TimeBits)),
		seqMask:       uint64(^(int64(-1) << c.SeqBits)),
		machineIDMask: machineIDMask,
		machineID:     (c.MachineID & machineIDMask) << c.SeqBits,
		startTimeNano: c.StartTime,
	}
}

func autoincrementID(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.SNID_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{Command: req.Command}
	if req.Header != nil {
		w.Header = req.Header
	}

	if frame.K == "" {
		return errcode.Bad(w, errcode.ErrKeyIsEmpty), nil
	}

	num := frame.N
	if num < 1 {
		num = 1
	} else if num > 100 {
		return errcode.Bad(w, errcode.ErrMaxIDNumber), nil
	}

	values, err := dao.AutoincrementID(frame.K, int64(num))
	if err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	resp := &pb.SNID_Reply{
		Values: values,
	}

	b, _ := grpcproto.Marshal(resp)
	w.Payload = &corespb.Response_Content{Content: b}
	return w, nil
}
