/**
 * doc https://github.com/baidu/uid-generator/blob/master/README.zh_cn.md
 * Represents an implementation of {@link UidGenerator}
 *
 * The unique id has 64bits (long), default allocated as blow:<br>
 * <li>sign: The highest bit is 0
 * <li>delta seconds: The next 28 bits, represents delta seconds since a customer epoch(2016-05-20 00:00:00.000).
 *                    Supports about 8.7 years until to 2024-11-20 21:24:16
 * <li>worker id: The next 22 bits, represents the worker's id which assigns based on database, max id is about 420W
 * <li>sequence: The next 13 bits, represents a sequence within the same second, max for 8192/s<br><br>
 *
 * The {@link DefaultUidGenerator#parseUID(long)} is a tool method to parse the bits
 *
 * <pre>{@code
 * +------+----------------------+----------------+-----------+
 * | sign |     delta seconds    | worker node id | sequence  |
 * +------+----------------------+----------------+-----------+
 *   1bit          28bits              22bits         13bits
 * }</pre>
 *
 * You can also specified the bits by Spring property setting.
 * <li>timeBits: default as 28
 * <li>workerBits: default as 22
 * <li>seqBits: default as 13
 * <li>epochStr: Epoch date string format 'yyyy-MM-dd'. Default as '2016-05-20'<p>
 *
 * <b>Note that:</b> The total bits must be 64 -1
 **/

package uid

import (
	"fmt"
	"sync"
	"time"
)

const (
	defaultBaiduTimeBits   = 28
	defaultBaiduWorkerBits = 22
	defaultBaiduSeqBits    = 13
)

type (
	BaiduUidConfig struct {
		TimeBits   int    `alias:"timeBits" default:"28"`
		WorkerBits int    `alias:"workerBits" default:"22"`
		SeqBits    int    `alias:"seqBits" default:"13"`
		EpochStr   string `alias:"epochStr" default:"2021-10-22"`
		WorkerId   uint64 `alias:"workerId" default:"1"`
	}

	BaiduUidGenerator struct {
		timeBits        int
		workerBits      int
		seqBits         int
		epochStr        string
		epochSeconds    uint64
		workerId        uint64
		sequence        uint64
		lastSecond      uint64
		maxDeltaSeconds uint64
		maxWorkerId     uint64
		maxSequence     uint64
		timestampShift  int
		workerIdShift   int
		sync.RWMutex
	}
)

func (uid *BaiduUidGenerator) NextId() (uint64, error) {
	currentSecond, err := uid.getCurrentSecond()
	if err != nil {
		return 0, err
	}

	uid.RLock()
	lastSecond := uid.lastSecond
	sequence := uid.sequence
	maxSequence := uid.maxSequence
	epochSeconds := uid.epochSeconds
	timestampShift := uid.timestampShift
	workerId := uid.workerId
	workerIdShift := uid.workerIdShift
	uid.RUnlock()

	if currentSecond < lastSecond {
		return 0, fmt.Errorf("Timestamp bits is exhausted. Refusing UID generate. Now: %d", lastSecond-currentSecond)
	}

	if currentSecond == lastSecond {
		sequence = (sequence + 1) & maxSequence
		if sequence == 0 {
			currentSecond, err = uid.getNextSecond(lastSecond)
			if err != nil {
				return 0, err
			}
		}
	} else {
		sequence = 0
	}

	uid.Lock()
	uid.lastSecond = currentSecond
	uid.sequence = sequence
	uid.Unlock()

	deltaSeconds := currentSecond - epochSeconds
	return (deltaSeconds << uint(timestampShift)) | (workerId << uint(workerIdShift)) | sequence, nil
}

func (uid *BaiduUidGenerator) getCurrentSecond() (uint64, error) {
	uid.RLock()
	epochSeconds := uid.epochSeconds
	maxDeltaSeconds := uid.maxDeltaSeconds
	uid.RUnlock()

	currentSecond := uint64(time.Now().Unix())
	if currentSecond-epochSeconds > maxDeltaSeconds {
		return 0, fmt.Errorf("Timestamp bits is exhausted. Refusing UID generate. Now: %d %d", currentSecond, maxDeltaSeconds)
	}

	return currentSecond, nil
}

func (uid *BaiduUidGenerator) getNextSecond(lastTimestamp uint64) (uint64, error) {
	timestamp, err := uid.getCurrentSecond()
	if err != nil {
		return 0, err
	}

	for timestamp <= lastTimestamp {
		time.Sleep(time.Duration(lastTimestamp-timestamp) * time.Millisecond)
		timestamp, err = uid.getCurrentSecond()
		if err != nil {
			return 0, err
		}
	}

	return timestamp, nil
}

func (uid *BaiduUidGenerator) Parse(id uint64) string {
	uid.RLock()
	timestampBits := uid.timeBits
	workerIdBits := uid.workerBits
	sequenceBits := uid.seqBits
	epochSeconds := uid.epochSeconds
	uid.RUnlock()

	totalBits := 1 + timestampBits + workerIdBits + sequenceBits
	sequence := (id << uint(totalBits-sequenceBits)) >> uint(totalBits-sequenceBits)
	workerId := (id << uint(timestampBits+1)) >> uint(totalBits-workerIdBits)
	deltaSeconds := id >> uint(workerIdBits+sequenceBits)
	thatTime := time.Unix(int64(epochSeconds+deltaSeconds), 0)
	return fmt.Sprintf("{\"UID\":\"%d\",\"timestamp\":\"%s\",\"workerId\":\"%d\",\"sequence\":\"%d\"}", id, thatTime.Format("2006-01-02 00:00:00"), workerId, sequence)
}

func NewBaiduUidGenerator(c BaiduUidConfig) *BaiduUidGenerator {
	if c.TimeBits < 1 {
		c.TimeBits = defaultBaiduTimeBits
	}

	if c.WorkerBits < 1 {
		c.WorkerBits = defaultBaiduWorkerBits
	}

	if c.SeqBits < 1 {
		c.SeqBits = defaultBaiduSeqBits
	}

	maxDeltaSeconds := ^(-1 << c.TimeBits)
	maxWorkerId := ^(-1 << c.WorkerBits)
	maxSequence := ^(-1 << c.SeqBits)
	uidg := BaiduUidGenerator{
		timeBits:        c.TimeBits,
		workerBits:      c.WorkerBits,
		seqBits:         c.SeqBits,
		epochStr:        c.EpochStr,
		workerId:        c.WorkerId,
		maxDeltaSeconds: uint64(maxDeltaSeconds),
		maxWorkerId:     uint64(maxWorkerId),
		maxSequence:     uint64(maxSequence),
		timestampShift:  c.WorkerBits + c.SeqBits,
		workerIdShift:   c.SeqBits,
	}

	es, _ := time.Parse("2006-01-02", uidg.epochStr)
	uidg.epochSeconds = uint64(es.Unix())
	return &uidg
}
