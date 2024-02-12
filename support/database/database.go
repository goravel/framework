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

// GetForeignKeyField retrieves the foreign key field name between parentModel and childModel by checking gorm tags and associations.
// If not found, it returns an empty string.
func GetForeignKeyField(parentModel,childModel interface{}) string {
	parentType := reflect.TypeOf(parentModel) //get type
	childType := reflect.TypeOf(childModel)

    for i := 0; i < parentType.NumField(); i++ {
        field := parentType.Field(i)
        if strings.Contains(field.Tag.Get("gorm"), "ForeignKey") {
            return field.Name
        }
    }

    // try to find an association field in the child model
    for i := 0; i < childType.NumField(); i++ {
        field := childType.Field(i)
		fmt.Println(field)
        if field.Type.Kind() == reflect.Ptr && field.Type.Elem() == parentType {
            return ToSnakeCase(field.Name) + "_id"
        }
    }

	return ""
}