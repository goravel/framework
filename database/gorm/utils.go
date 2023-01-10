package gorm

import "reflect"

func copyStruct(dest any) reflect.Value {
	t := reflect.TypeOf(dest).Elem()
	v := reflect.ValueOf(dest).Elem()
	destFields := make([]reflect.StructField, 0)
	for i := 0; i < t.NumField(); i++ {
		destFields = append(destFields, t.Field(i))
	}
	copyDestStruct := reflect.StructOf(destFields)

	return v.Convert(copyDestStruct)
}
