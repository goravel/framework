package database

import (
	"reflect"
	"strings"

	"github.com/spf13/cast"
)

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


// GetIDField returns the name of the primary key field of the provided struct (dest).
func GetPrimaryField(dest any) string {
	if dest == nil {
		return ""
	}

	t := reflect.TypeOf(dest)
	v := reflect.ValueOf(dest)

	if t.Kind() == reflect.Pointer {
		return GetPrimaryKeyField(t.Elem(), v.Elem())
	}

	return GetPrimaryKeyField(t, v)
}

func GetIDByReflect(t reflect.Type, v reflect.Value) any {
	for i := 0; i < t.NumField(); i++ {
		if !t.Field(i).IsExported() {
			continue
		}
		if t.Field(i).Name == "Model" && v.Field(i).Type().Kind() == reflect.Struct {
			structField := v.Field(i).Type()
			for j := 0; j < structField.NumField(); j++ {
				if !structField.Field(j).IsExported() {
					continue
				}
				if strings.Contains(structField.Field(j).Tag.Get("gorm"), "primaryKey") {
					id := v.Field(i).Field(j).Interface()
					if cast.ToString(id) == "" || cast.ToInt(id) == 0 {
						return nil
					}

					return id
				}
			}
		}
		if strings.Contains(t.Field(i).Tag.Get("gorm"), "primaryKey") {
			id := v.Field(i).Interface()
			if cast.ToString(id) == "" && cast.ToInt(id) == 0 {
				return nil
			}

			return id
		}
	}

	return nil
}

func GetPrimaryKeyField(t reflect.Type, v reflect.Value) string {
    for i := 0; i < t.NumField(); i++ {
        if !t.Field(i).IsExported() {
            continue
        }

        // Check for Model struct and iterate through its fields
        if t.Field(i).Name == "Model" && v.Field(i).Type().Kind() == reflect.Struct {
            structField := v.Field(i).Type()
            for j := 0; j < structField.NumField(); j++ {
                if !structField.Field(j).IsExported() {
                    continue
                }

                // Check if the field is the primary key
                if strings.Contains(structField.Field(j).Tag.Get("gorm"), "primaryKey") {
                    return structField.Field(j).Name
                }
            }
        }

        // Check if the field itself is the primary key
        if strings.Contains(t.Field(i).Tag.Get("gorm"), "primaryKey") {
            return t.Field(i).Name
        }
    }

    return ""
}