// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>

package proto

import (
	"fmt"
	"reflect"
)

// FastPack 快速封包
type FastPack interface {
	// Pack 封包
	Pack(w *BytesBuffer) error
}

// Pack 封包
func Pack(w *BytesBuffer, tbl interface{}) ([]byte, error) {
	if tbl == nil {
		return w.Data(), nil
	}

	// 如果传的对象自有封包方法
	// 那么就直接调用它
	if fast, ok := tbl.(FastPack); ok {
		if err := fast.Pack(w); err != nil {
			return nil, err
		}

		return w.Data(), nil
	}

	if err := pack(reflect.ValueOf(tbl), w); err != nil {
		return nil, err
	}
	return w.Data(), nil
}

func pack(v reflect.Value, w *BytesBuffer) (err error) {
	switch v.Kind() {
	case reflect.Bool:
		err = w.WriteBool(v.Bool())
	case reflect.Uint8:
		err = w.WriteUint8(uint8(v.Uint()))
	case reflect.Uint16:
		err = w.WriteUint16(uint16(v.Uint()))
	case reflect.Uint32:
		err = w.WriteUint32(uint32(v.Uint()))
	case reflect.Uint64:
		err = w.WriteUint64(uint64(v.Uint()))
	case reflect.Int8:
		err = w.WriteInt8(int8(v.Int()))
	case reflect.Int16:
		err = w.WriteInt16(int16(v.Int()))
	case reflect.Int32:
		err = w.WriteInt32(int32(v.Int()))
	case reflect.Int64:
		err = w.WriteInt64(int64(v.Int()))
	case reflect.Float32:
		err = w.WriteFloat32(float32(v.Float()))
	case reflect.Float64:
		err = w.WriteFloat64(v.Float())
	case reflect.String:
		err = w.WriteString(v.String())
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return nil
		}

		err = pack(v.Elem(), w)

	case reflect.Slice:
		if bs, ok := v.Interface().([]byte); ok {
			err = w.WriteUint16Bytes(bs)
		} else {
			size := v.Len()
			if err := w.WriteUint16(uint16(size)); err != nil {
				return err
			}

			for i := 0; i < size; i++ {
				if err := pack(v.Index(i), w); err != nil {
					return err
				}
			}
		}

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := pack(v.Field(i), w); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("Unexpected type: %v", v.Kind())
	}

	return
}
