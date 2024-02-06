package database

import (
	"fmt"
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
		return extractPrimaryKeyField(t.Elem(), v.Elem())
	}

	return extractPrimaryKeyField(t, v)
}

func GetForeignField(dest any) string {

	t := reflect.TypeOf(dest)

	fmt.Println(t)
	if t.Kind() == reflect.Pointer {
		return extractForeignField(t.Elem())
	}
	return extractForeignField(t)
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

func extractPrimaryKeyField(t reflect.Type, v reflect.Value) string {
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
					fmt.Println(structField.Field(j).Name)
					return structField.Field(j).Name
				}
			}
		}

		// Check if the field itself is the primary key
		if strings.Contains(t.Field(i).Tag.Get("gorm"), "primaryKey") {
			fmt.Println(t.Field(i).Name)
			return t.Field(i).Name
		}
	}

	return ""
}

func extractForeignField(t reflect.Type) string {

	for i := 0; i < t.NumField(); i++ {
		if !t.Field(i).IsExported() {
			continue
		}

		// Check if the field has a foreign key tag (e.g., `gorm:"foreignKey:UserID"`)
		if foreignKey := t.Field(i).Tag.Get("gorm"); strings.Contains(foreignKey, "foreignKey") {
			// Parse the foreign key tag to extract the field name
			parts := strings.Split(foreignKey, ":")
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}
	return ""
}
