package proto

import (
	"reflect"
	"testing"
)

func TestRequestBytes(t *testing.T) {
	rq := &RequestBytes{}
	rq.Ver = 1
	rq.Cmd = 10001
	rq.SubCmd = 12
	rq.SeqID = 154
	rq.Content = make([]byte, 10)

	bytes0, err := rq.Marshal()
	if err != nil {
		t.Fatal(err)
		return
	}

	rd2 := &RequestBytes{}
	if err := rd2.Unmarshal(bytes0); err != nil {
		t.Fatal(err)
		return
	}

	b0, _ := rd2.Marshal()
	if !reflect.DeepEqual(bytes0, b0) {
		t.Fatalf("Not Equal:\nReceived: '%+v'\nExpected: '%+v'\n", bytes0, b0)
		return
	}
}
