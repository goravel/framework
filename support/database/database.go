package database

import "reflect"

func GetID(dest any) any {
	if dest == nil {
		return nil
	}

	t := reflect.TypeOf(dest)
	v := reflect.ValueOf(dest)

	if t.Kind() == reflect.Pointer {
		return GetIDByReflect(t.Elem(), v.Elem())
	}

	return GetIDByReflect(t, v)
}

func GetIDByReflect(t reflect.Type, v reflect.Value) any {
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == "Model" && v.Field(i).Type().Kind() == reflect.Struct {
			structField := v.Field(i).Type()
			for j := 0; j < structField.NumField(); j++ {
				if structField.Field(j).Tag.Get("gorm") == "primaryKey" {
					return v.Field(i).Field(j).Interface()
				}
			}
		}
		if t.Field(i).Tag.Get("gorm") == "primaryKey" {
			return v.Field(i).Interface()
		}
	}

	return nil
}
