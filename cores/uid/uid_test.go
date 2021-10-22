package uid

import "testing"

func TestUid(t *testing.T) {
	u1 := NewBaiduUidGenerator(BaiduUidConfig{EpochStr: "2021-10-22", WorkerId: 1})
	u2 := NewSnowflakeGenerator(SnowflakeConfig{MachineID: 1})

	id1, err := u1.NextId()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(id1)
	t.Log(u2.NextId())
}
