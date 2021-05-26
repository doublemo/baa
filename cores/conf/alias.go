// Copyright (c) 2021 The baa Authors <https://github.com/doublemo/baa>

package conf

import (
	"reflect"
)

// BindWithConfFile 将配置文件绑定到对象
// 需要conf库的支持
func BindWithConfFile(fp string, o interface{}) error {
	mapping, err := ParseFile(fp)
	if err != nil {
		return err
	}

	return Bind(mapping, o)
}

// BindWithConf 将配置文件内容绑定到对象
// 需要conf库的支持
func BindWithConf(data string, o interface{}) error {
	mapping, err := Parse(data)
	if err != nil {
		return err
	}
	return Bind(mapping, o)
}

// Bind 将内容绑定到对象
func Bind(mapping map[string]interface{}, o interface{}) error {
	defaultVal := make(map[string]interface{})
	value := reflect.ValueOf(o)
	typElem := reflect.TypeOf(o).Elem()
	for i := 0; i < typElem.NumField(); i++ {
		field := typElem.Field(i)
		name := field.Name
		if alias, ok := field.Tag.Lookup("alias"); ok {
			if alias == "-" {
				name = ""
			} else {
				name = alias
			}
		}

		if name == "" {
			continue
		}

		if val, ok := field.Tag.Lookup("default"); ok && val != "" {
			mapping, err := Parse(name + " = " + val)
			if err != nil {
				return err
			}
			defaultVal = mapping
		}

		m, ok := mapping[name]
		if !ok {
			m, ok = defaultVal[name]
			if !ok {
				continue
			}
		}

		v, err := bindValue(field.Type, m)
		if err != nil {
			return err
		}

		if field.Type.Kind() != reflect.Ptr && v.Type().Kind() == reflect.Ptr {
			value.Elem().FieldByName(field.Name).Set(v.Elem())
		} else {
			value.Elem().FieldByName(field.Name).Set(v)
		}
	}

	return nil
}
