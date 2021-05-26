package conf

import (
	"reflect"
	"testing"
)

type Cd1 struct {
	X *testM
}

type testM struct {
	X uint
	D map[int]bool
	H []int
}

type Cd struct {
	X uint
	V []map[string]string
}

func TestBindValue(t *testing.T) {
	var v int32
	v = 3
	ko := &testM{}
	v0, err := bindValue(reflect.TypeOf(ko.X), v)
	if err != nil {
		t.Fatal(err)
		return
	}

	x := reflect.ValueOf(ko).Elem().FieldByName("X")
	x.Set(v0)

	n := make(map[int]bool)
	n[1] = true
	n[2] = false

	v1, err := bindValue(reflect.TypeOf(ko.D), n)
	if err != nil {
		t.Fatal(err)
		return
	}

	x1 := reflect.ValueOf(ko).Elem().FieldByName("D")
	x1.Set(v1)

	n0 := make([]int64, 3)
	n0[0] = 1
	n0[1] = 3
	n0[2] = 50

	v2, err := bindValue(reflect.TypeOf(ko.H), n0)
	if err != nil {
		t.Fatal(err)
		return
	}

	x2 := reflect.ValueOf(ko).Elem().FieldByName("H")
	x2.Set(v2)

	n1 := make(map[string]interface{})
	n1["X"] = 1234

	ko1 := testM{}
	v3, err := bindValue(reflect.TypeOf(ko1), n1)
	if err != nil {
		t.Fatal(err)
		return
	}

	kobb := &testM{X: 1990, D: map[int]bool{1: true, 4: false}, H: []int{1323, 455}}
	ko2 := &Cd1{}
	v4, err := bindValue(reflect.TypeOf(ko2.X), kobb)
	if err != nil {
		t.Fatal(err)
		return
	}

	x3 := reflect.ValueOf(ko2)
	x3.Elem().FieldByName("X").Set(v4)
	t.Log(ko, v3, v4, ko2, x3.CanAddr(), x3.Type().Kind(), v4.Type().Kind())

	ko4 := &Cd{}
	v6, err := bindValue(reflect.TypeOf(ko4.V), []map[string]string{map[string]string{"ddd": "ee"}, map[string]string{"333": "cc"}})
	if err != nil {
		t.Fatal(err)
		return
	}

	t.Log(v6)
}

type authorization struct {
	Users []map[string]string `alias:"users"`
	Md    string              `default:"你好"`
}

type TestConfig struct {
	Listen string `alias:"listen"`

	Authorization authorization `alias:"authorization"`

	A  uint8          `alias:"aint8"`
	B  int32          `alias:"bint32"`
	C  []string       `alias:"ddDjjdXjj"`
	D  bool           `alias:"css_ddd_dd"`
	F  []int32        `alias:"varrtint32"`
	H  []float32      `alias:"varrtint323"`
	G  []map[int]bool `alias:"tesv"`
	Hx []int64        `default:"[ 1,2,3,4 ]"`
}

func TestBindWithConfFile(t *testing.T) {
	c2 := &TestConfig{
		Listen: "127.0.0.1:4222",
		Authorization: authorization{
			Users: []map[string]string{
				map[string]string{"password": "$2a$10$UHR6GhotWhpLsKtVP0/i6.Nh9.fuY73cWjLoJjb2sKT8KISBcUW5q", "user": "alice"},
				map[string]string{"password": "$2a$11$dZM98SpGeI7dCFFGSpt.JObQcix8YHml4TBUZoge9R1uxnMIln5ly", "user": "bob"}},

			Md: "你好",
		},

		A: 255,
		B: 2666,
		C: []string{"a.com", "b.com", "c.com"},
		D: true,
		F: []int32{1, 3, 36},
		H: []float32{1.2, 3.5, 4.66},
		G: []map[int]bool{
			map[int]bool{1: true, 2: false},
			map[int]bool{5: true, 6: true},
		},
		Hx: []int64{1, 2, 3, 4},
	}
	c := &TestConfig{}
	if err := BindWithConfFile("./simple.conf", c); err != nil {
		t.Fatal(err)
		return
	}

	if !reflect.DeepEqual(c, c2) {
		t.Fatalf("Not Equal:\nReceived: '%+v'\nExpected: '%+v'\n", c2, c)
		return
	}
}

func TestBindWithConf(t *testing.T) {
	str := `1235`
	mapping, err := Parse(str)
	t.Log(mapping, err)
}
