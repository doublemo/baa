package uid

import (
	"sync"
	"time"
)

const (
	defaultSnowflakeTimeBits    uint = 41
	defaultSnowflakeSeqBits     uint = 12
	defaultSnowflakeMachineBits uint = 10
)

type (
	// SnowflakeConfig 配置文件
	SnowflakeConfig struct {
		// MachineID 机器码
		MachineID uint64 `alias:"machineid" default:"1"`

		// TimeBits 时间位
		TimeBits uint `alias:"timebits" default:"41"`

		// SeqBits 信息位
		SeqBits uint `alias:"seqbits" default:"12"`

		// MachineBits 机器码位
		MachineBits uint `alias:"machinebits" default:"10"`

		// StartTime UUID产生的开始时间，为纳秒 (2^41-1)/(1000 * 60 * 60 * 24 * 365) = 69
		StartTime int64 `alias:"starttime" default:"1634140800000000000"`
	}

	// Snowflake 原始雪花算法
	Snowflake struct {
		timeBits      uint
		seqBits       uint
		machineBits   uint
		timeMask      uint64
		seqMask       uint64
		machineIDMask uint64
		machineID     uint64
		startTimeNano int64
		lastts        int64
		seq           uint64
		mutex         sync.RWMutex
	}
)

// NextId 获取ID
func (uid *Snowflake) NextId() uint64 {
	uid.mutex.RLock()
	lastts := uid.lastts
	seq := uid.seq
	timeMask := uid.timeMask
	machineBits := uid.machineBits
	seqBits := uid.seqBits
	machineID := uid.machineID
	uid.mutex.RUnlock()

	t := uid.ts()
	if t < lastts {
		t = uid.wait(lastts)
	}

	if lastts == t {
		seq = (seq + 1) & uid.seqMask
		if seq == 0 {
			t = uid.wait(lastts)
		}
	} else {
		seq = 0
	}

	uid.mutex.Lock()
	uid.lastts = t
	uid.seq = seq
	uid.mutex.Unlock()

	var id uint64
	id |= (uint64(t) & timeMask) << (machineBits + seqBits)
	id |= machineID
	id |= seq
	return id
}

func (uid *Snowflake) wait(s int64) int64 {
	t := uid.ts()
	for t < s {
		time.Sleep(time.Duration(s-t) * time.Millisecond)
		t = uid.ts()
	}
	return t
}

func (uid *Snowflake) ts() int64 {
	return (time.Now().UnixNano() - uid.startTimeNano) / int64(time.Millisecond)
}

func NewSnowflakeGenerator(c SnowflakeConfig) *Snowflake {
	if c.MachineID < 1 {
		c.MachineID = 1
	}

	if c.TimeBits < 1 {
		c.TimeBits = defaultSnowflakeTimeBits
	}

	if c.MachineBits < 1 {
		c.MachineBits = defaultSnowflakeMachineBits
	}

	if c.SeqBits < 1 {
		c.SeqBits = defaultSnowflakeSeqBits
	}

	if c.StartTime < 1 {
		c.StartTime = 1634140800000000000
	}

	machineIDMask := uint64(^(int64(-1) << c.MachineBits))
	return &Snowflake{
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
