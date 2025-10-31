package gorm

import (
	"reflect"

	gormio "gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func copyStruct(dest any) reflect.Value {
	t := reflect.TypeOf(dest)
	v := reflect.ValueOf(dest)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	}

	destFields := make([]reflect.StructField, 0)
	for i := 0; i < t.NumField(); i++ {
		destFields = append(destFields, t.Field(i))
	}
	copyDestStruct := reflect.StructOf(destFields)

	return v.Convert(copyDestStruct)
}

func getZeroValueFromReflectType(t reflect.Type) any {
	if t.Kind() == reflect.Pointer {
		return reflect.New(t.Elem()).Interface()
	}
	return reflect.New(t.Elem()).Interface()
}

func getModelSchema(model any, db *gormio.DB) (*schema.Schema, error) {
	stmt := gormio.Statement{DB: db}
	err := stmt.Parse(model)
	if err != nil {
		return nil, err
	}
	return stmt.Schema, nil
}
