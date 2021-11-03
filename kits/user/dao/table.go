package dao

import (
	"fmt"
	"hash/crc32"
	"reflect"
	"strconv"

	"gorm.io/gorm"
)

// Table 表接口
type Table interface {
	TableName() string
}

func useTable(v interface{}, table Table, maxRecord, maxTable uint32) func(tx *gorm.DB) *gorm.DB {
	var c32 uint32
	switch v := reflect.ValueOf(v); v.Kind() {
	case reflect.String:
		c32 = makeTablenameFromString(v.String(), maxRecord, maxTable)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		c32 = makeTablenameFromUint64(uint64(v.Int()), maxRecord, maxTable)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		c32 = makeTablenameFromUint64(v.Uint(), maxRecord, maxTable)
	default:
		panic(fmt.Sprintf("unhandled kind %s", v.Kind()))
	}

	return func(tx *gorm.DB) *gorm.DB {
		tablename := DBNamer(table.TableName(), strconv.FormatUint(uint64(c32), 10))
		_, ok := tableCacher.Get(tablename)
		if !ok {
			if !tx.Migrator().HasTable(tablename) {
				if err := tx.Table(tablename).AutoMigrate(table); err != nil {
					tx.AddError(err)
				} else {
					tableCacher.Set(tablename, true, 0)
				}
			} else {
				tableCacher.Set(tablename, true, 0)
			}
		}
		return tx.Table(tablename)
	}
}

// 计算表名
func makeTablenameFromUint64(id uint64, maxRecord, maxTable uint32) uint32 {
	c32 := crc32.ChecksumIEEE([]byte(strconv.FormatUint(id, 10)))
	return (c32 - (c32 / maxRecord * maxRecord)) % maxTable
}

func makeTablenameFromString(s string, maxRecord, maxTable uint32) uint32 {
	c32 := crc32.ChecksumIEEE([]byte(s))
	return (c32 - (c32 / maxRecord * maxRecord)) % maxTable
}
