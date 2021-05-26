// Copyright (c) 2021 The baa Authors <https://github.com/doublemo/baa>

package conf

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

// bindValue 根据指定类型解析结果
func bindValue(typ reflect.Type, v interface{}) (value reflect.Value, err error) {
	value = reflect.New(typ).Elem()
	valueKind := reflect.TypeOf(v).Kind()

	// value is string
	if valueKind == reflect.String {
		value, err = bindValueFromString(typ, v.(string))
		if err == nil {
			return
		}
	}

	switch typ.Kind() {
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		ret, err := assetValueToUint(v)
		if err == nil {
			value.SetUint(ret)
		}

	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int8, reflect.Int64:
		ret, err := assetValueToUint(v)
		if err == nil {
			value.SetInt(int64(ret))
		}

	case reflect.Float32, reflect.Float64:
		ret, err := assetValueToFloat(v)
		if err == nil {
			value.SetFloat(ret)
		}

	case reflect.String:
		ret, err := assetValueToString(v)
		if err == nil {
			value.SetString(ret)
		}

	case reflect.Bool:
		ret, err := assetValueToBool(v)
		if err == nil {
			value.SetBool(ret)
		}

	case reflect.Slice:
		value, err = bindSlice(typ, v)

	case reflect.Map:
		value, err = bindMap(typ, v)

	case reflect.Struct:
		value, err = bindStruct(typ, v)

	case reflect.Ptr:
		value, err = bindValue(typ.Elem(), v)

	default:
		err = fmt.Errorf("Unexpected type: %v", typ.Kind())
	}
	return
}

// bindSlice 处理切片
func bindSlice(typ reflect.Type, v interface{}) (reflect.Value, error) {
	value := reflect.ValueOf(v)
	if value.Type().Kind() != reflect.Slice {
		return reflect.Zero(typ), fmt.Errorf("Unsupported slice values:%v", v)
	}

	length := value.Len()
	result := reflect.MakeSlice(typ, length, value.Cap())
	for i := 0; i < length; i++ {
		m, err := bindValue(typ.Elem(), value.Index(i).Interface())
		if err != nil {
			return result, err
		}
		result.Index(i).Set(m)
	}
	return result, nil
}

// bindMap 处理map
func bindMap(typ reflect.Type, v interface{}) (reflect.Value, error) {
	value := reflect.ValueOf(v)
	if value.Type().Kind() != reflect.Map {
		return reflect.Zero(typ), fmt.Errorf("Unsupported map values:%v %v", value.Type().Kind(), v)
	}
	result := reflect.MakeMap(typ)
	iter := value.MapRange()
	for iter.Next() {
		k, err := bindValue(typ.Key(), iter.Key().Interface())
		if err != nil {
			return result, err
		}

		m, err := bindValue(typ.Elem(), iter.Value().Interface())
		if err != nil {
			return result, err
		}

		result.SetMapIndex(k, m)
	}
	return result, nil
}

// bindStruct 处理struct
func bindStruct(typ reflect.Type, v interface{}) (reflect.Value, error) {
	value := reflect.ValueOf(v)
	valueKind := value.Type().Kind()

	if valueKind == reflect.Ptr {
		valueKind = value.Elem().Type().Kind()
		v = value.Elem().Interface()
	}

	if valueKind == reflect.Map {
		return bindStructMap(typ, v)
	} else if valueKind == reflect.Struct {
		return bindStructStruct(typ, v)
	} else if valueKind == reflect.Ptr {
		log.Println(value.Elem().Type().Kind())
	}

	return reflect.Zero(typ), fmt.Errorf("Unsupported struct values:%v", v)
}

// bindStructMap 处理map to struct
func bindStructMap(typ reflect.Type, v interface{}) (reflect.Value, error) {
	value := reflect.ValueOf(v)
	valueTyp := value.Type()
	if valueTyp.Kind() != reflect.Map && valueTyp.Key().Kind() != reflect.String || value.IsNil() {
		return reflect.Zero(typ), fmt.Errorf("Unsupported map values:%v", v)
	}

	result := reflect.New(typ)
	resultElem := result.Type().Elem()
	defaultVal := make(map[string]interface{})
	for i := 0; i < resultElem.NumField(); i++ {
		field := resultElem.Field(i)
		name := field.Name
		if alias, ok := field.Tag.Lookup("alias"); ok {
			if alias == "-" {
				continue
			}

			name = alias
		}

		if val, ok := field.Tag.Lookup("default"); ok && val != "" {
			mapping, err := Parse(name + " = " + val)
			if err != nil {
				return value, err
			}

			defaultVal = mapping
		}

		mapValue := value.MapIndex(reflect.ValueOf(name))
		if !mapValue.IsValid() {
			if m, ok := defaultVal[name]; ok {
				mapValue = reflect.ValueOf(m)
			} else {
				continue
			}
		}

		m, err := bindValue(field.Type, mapValue.Interface())
		if err != nil {
			return value, err
		}
		result.Elem().FieldByName(field.Name).Set(m)
	}
	return result, nil
}

// bindStructStruct 处理struct to struct
func bindStructStruct(typ reflect.Type, v interface{}) (reflect.Value, error) {
	defaultVal := make(map[string]interface{})
	value := reflect.ValueOf(v)
	valueTyp := value.Type()
	if valueTyp.Kind() != reflect.Struct {
		return reflect.Zero(typ), fmt.Errorf("Unsupported struct values:%v", v)
	}

	result := reflect.New(typ)
	resultElem := result.Type().Elem()
	mapValue := reflect.Zero(valueTyp)

	for i := 0; i < resultElem.NumField(); i++ {
		field := resultElem.Field(i)
		name := field.Name
		if alias, ok := field.Tag.Lookup("alias"); ok {
			if alias == "-" {
				continue
			}

			name = alias
		}

		if val, ok := field.Tag.Lookup("default"); ok && val != "" {
			mapping, err := Parse(name + " = " + val)
			if err != nil {
				return value, err
			}

			defaultVal = mapping
		}

		if vname := findStructSomename(name, value); vname != "" {
			mapValue = value.FieldByName(vname)
			if !mapValue.IsValid() {
				if m, ok := defaultVal[name]; ok {
					mapValue = reflect.ValueOf(m)
				} else {
					continue
				}
			}
		} else {
			if m, ok := defaultVal[name]; ok {
				mapValue = reflect.ValueOf(m)
			}
		}

		if !mapValue.IsValid() || mapValue.IsZero() {
			continue
		}

		m, err := bindValue(field.Type, mapValue.Interface())
		if err != nil {
			return value, err
		}

		result.Elem().FieldByName(field.Name).Set(m)
	}

	return result, nil
}

func findStructSomename(name string, value reflect.Value) string {
	valueName := ""
	for j := 0; j < value.NumField(); j++ {
		vfield := value.Type().Field(j)
		if alias, ok := vfield.Tag.Lookup("alias"); ok {
			if alias == name {
				valueName = vfield.Name
				break
			}
		}

		if vfield.Name == name {
			valueName = vfield.Name
			break
		}
	}
	return valueName
}

// bindValueFromString 处理内容为字符串的信息
func bindValueFromString(typ reflect.Type, v string) (reflect.Value, error) {
	value := reflect.New(typ).Elem()
	switch typ.Kind() {
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		m, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return value, err
		}

		value.SetUint(m)

	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		m, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return value, err
		}

		value.SetInt(m)

	case reflect.Float32, reflect.Float64:
		m, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return value, err
		}

		value.SetFloat(m)

	case reflect.Bool:
		m, err := strconv.ParseBool(v)
		if err != nil {
			return value, err
		}

		value.SetBool(m)

	case reflect.String:
		value.SetString(v)

	default:
		return value, fmt.Errorf("[bindValueFromString] Unexpected type: %v", typ.Kind())
	}

	return value, nil
}

// assetValueToUint 尝试转换为uint64
func assetValueToUint(v interface{}) (uint64, error) {
	var value uint64
	kind := reflect.TypeOf(v).Kind()
	switch kind {
	case reflect.Uint:
		value = uint64(v.(uint))
	case reflect.Uint16:
		value = uint64(v.(uint16))
	case reflect.Uint32:
		value = uint64(v.(uint32))
	case reflect.Uint64:
		value = v.(uint64)
	case reflect.Uint8:
		value = uint64(v.(uint8))
	case reflect.Int:
		value = uint64(v.(int))
	case reflect.Int16:
		value = uint64(v.(int16))
	case reflect.Int32:
		value = uint64(v.(int32))
	case reflect.Int64:
		value = uint64(v.(int64))
	case reflect.Int8:
		value = uint64(v.(int8))
	case reflect.Float32:
		value = uint64(v.(float32))
	case reflect.Float64:
		value = uint64(v.(float64))
	case reflect.Bool:
		if v.(bool) {
			value = 1
		} else {
			value = 0
		}
	case reflect.String:
		m, err := strconv.ParseUint(v.(string), 10, 64)
		if err != nil {
			return 0, err
		}
		value = m

	default:
		return 0, fmt.Errorf("Unexpected type: %v", kind)
	}

	return value, nil
}

// assetValueToFloat 尝试转换为float64
func assetValueToFloat(v interface{}) (float64, error) {
	var value float64
	kind := reflect.TypeOf(v).Kind()
	switch kind {
	case reflect.Uint:
		value = float64(v.(uint))
	case reflect.Uint16:
		value = float64(v.(uint16))
	case reflect.Uint32:
		value = float64(v.(uint32))
	case reflect.Uint64:
		value = float64(v.(uint64))
	case reflect.Uint8:
		value = float64(v.(uint8))
	case reflect.Int:
		value = float64(v.(int))
	case reflect.Int16:
		value = float64(v.(int16))
	case reflect.Int32:
		value = float64(v.(int32))
	case reflect.Int64:
		value = float64(v.(int64))
	case reflect.Int8:
		value = float64(v.(int8))
	case reflect.Float32:
		value = float64(v.(float32))
	case reflect.Float64:
		value = v.(float64)
	case reflect.Bool:
		if v.(bool) {
			value = 1
		} else {
			value = 0
		}
	case reflect.String:
		m, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return 0, err
		}
		value = m

	default:
		return 0, fmt.Errorf("Unexpected type: %v", kind)
	}

	return value, nil
}

// assetValueToString 尝试转换为string
func assetValueToString(v interface{}) (string, error) {
	kind := reflect.TypeOf(v).Kind()
	if kind == reflect.String {
		return v.(string), nil
	} else if kind >= reflect.Int && kind <= reflect.Uint64 {
		m, err := assetValueToUint(v)
		if err != nil {
			return "", err
		}

		return strconv.FormatUint(m, 64), nil
	} else if kind >= reflect.Float32 && kind <= reflect.Float64 {
		m, err := assetValueToFloat(v)
		if err != nil {
			return "", err
		}

		return strconv.FormatFloat(m, 'f', -1, 64), nil
	} else if reflect.Bool == kind {
		if v.(bool) {
			return "true", nil
		}

		return "false", nil
	}

	return "", fmt.Errorf("Unexpected string: %v", kind)
}

// assetValueToBool 尝试转换为bool
func assetValueToBool(v interface{}) (bool, error) {
	var value bool
	kind := reflect.TypeOf(v).Kind()
	switch kind {
	case reflect.Uint:
		value = v.(uint) > 0
	case reflect.Uint16:
		value = v.(uint16) > 0
	case reflect.Uint32:
		value = v.(uint32) > 0
	case reflect.Uint64:
		value = v.(uint64) > 0
	case reflect.Uint8:
		value = v.(uint8) > 0
	case reflect.Int:
		value = v.(int) > 0
	case reflect.Int16:
		value = v.(int16) > 0
	case reflect.Int32:
		value = v.(int32) > 0
	case reflect.Int64:
		value = v.(int64) > 0
	case reflect.Int8:
		value = v.(int8) > 0
	case reflect.Float32:
		value = v.(float32) > 0
	case reflect.Float64:
		value = v.(float64) > 0
	case reflect.Bool:
		value = v.(bool)
	case reflect.String:
		m, err := strconv.ParseBool(v.(string))
		if err != nil {
			return false, err
		}

		value = m

	default:
		return false, fmt.Errorf("Unexpected type: %v", kind)
	}

	return value, nil
}
