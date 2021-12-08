package dao

import (
	"reflect"

	"github.com/doublemo/baa/cores/conf"
)

const defaultChatMessageKey = "inbox"

// Messages 聊天信息
type Messages struct {
	ID          uint64 `gorm:"<-:create;primaryKey"`
	SeqId       uint64
	TSeqId      uint64
	FSeqId      uint64
	To          uint64
	From        uint64
	Content     string
	Group       int32
	ContentType string
	Topic       uint64
	Status      int32
	Origin      int32
	CreatedAt   int64
}

// ToMap 转为Map
func (m *Messages) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	elem := reflect.TypeOf(m).Elem()
	v := reflect.ValueOf(m).Elem()
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		tag, ok := f.Tag.Lookup("json")
		if ok {
			if tag == "-" {
				continue
			}
		} else {
			tag = f.Name
		}
		data[tag] = v.Field(i).Interface()
	}
	return data
}

// FromMap 从Map中绑定数据
func (m *Messages) FromMap(data map[string]interface{}) error {
	return conf.Bind(data, m)
}
