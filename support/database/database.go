package database

import (
	"github.com/jinzhu/inflection"
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

func GetForeignKeyField(model any, relation string) string {
	modelType := reflect.TypeOf(model) //get type
	return GetForeignKeyFieldByReflect(modelType, relation)
}

func GetForeignKeyFieldByReflect(t reflect.Type, relation string) string {
	field, ok := t.FieldByName(relation)
	if !ok {
		return ""
	}

	gormTag := field.Tag.Get("gorm")
	if strings.Contains(gormTag, "foreignKey") {
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "foreignKey") {
				return strings.TrimPrefix(part, "foreignKey:")
			}
		}
	}

	return inflection.Singular(relation) + "ID"
}

func GetPivotTableByReflect(t reflect.Type, relation string) string {
	field, ok := t.FieldByName(relation)
	if !ok {
		return ""
	}

	gormTag := field.Tag.Get("gorm")
	if strings.Contains(gormTag, "many2many") {
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "many2may") {
				return strings.TrimPrefix(part, "many2many:")
			}
		}
	}

	return ""
}

func IsMany2ManyByReflect(t reflect.Type, relation string) bool {
	field, ok := t.FieldByName(relation)
	if !ok {
		return false
	}

	gormTag := field.Tag.Get("gorm")
	if !strings.Contains(gormTag, "many2many") {
		return false
	}

	return true
}
