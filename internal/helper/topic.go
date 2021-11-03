package helper

import (
	"hash/crc64"
	"sort"
	"strconv"
)

// GenerateTopic 创建topic
func GenerateTopic(id ...uint64) uint64 {
	sort.Slice(id, func(i, j int) bool { return id[i] < id[j] })
	values := ""
	for _, v := range id {
		values += strconv.FormatUint(v, 10)
	}
	return crc64.Checksum([]byte(values), crc64.MakeTable(crc64.ECMA))
}
