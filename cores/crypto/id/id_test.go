package id

import "testing"

func TestID(t *testing.T) {
	var uid uint64
	uid = 4
	data, err := Encrypt(uid, []byte("7581BDD8E8DA3839"))
	if err != nil {
		t.Fatal(err)
	}

	v, err := Decrypt(data, []byte("7581BDD8E8DA3839"))
	if err != nil {
		t.Fatal(err)
	}

	if v != uid {
		t.Fatal("不相等")
	}

	t.Log("Code", data)
}
