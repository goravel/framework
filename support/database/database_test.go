package database

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/database/orm"
)

type TestStruct struct {
	ID   int `gorm:"primaryKey"`
	name string
}

type TestStructString struct {
	ID   string `gorm:"primaryKey"`
	name string
}

type TestStructUUID struct {
	ID   uuid.UUID `gorm:"primaryKey"`
	name string
}

type TestStructNoPK struct {
	ID   int
	name string
}

func TestGetID(t *testing.T) {
	tests := []struct {
		description string
		setup       func(description string)
	}{
		{
			description: "return value",
			setup: func(description string) {
				type User struct {
					ID     uint `gorm:"primaryKey"`
					Name   string
					Avatar string
				}
				user := User{}
				user.ID = 1
				assert.Equal(t, uint(1), GetID(&user), description)
			},
		},
		{
			description: "return value with orm.Model",
			setup: func(description string) {
				type User struct {
					orm.Model
					Name   string
					Avatar string
				}
				user := User{}
				user.ID = 1
				assert.Equal(t, uint(1), GetID(&user), description)
			},
		},
		{
			description: "return nil",
			setup: func(description string) {
				type User struct {
					Name   string
					Avatar string
				}
				user := User{}
				assert.Nil(t, GetID(&user), description)
			},
		},
		{
			description: "return value(struct)",
			setup: func(description string) {
				type User struct {
					ID     uint `gorm:"primaryKey"`
					Name   string
					Avatar string
				}
				user := User{}
				user.ID = 1
				assert.Equal(t, uint(1), GetID(user), description)
			},
		},
		{
			description: "return value with orm.Model",
			setup: func(description string) {
				type User struct {
					orm.Model
					Name   string
					Avatar string
				}
				user := User{}
				user.ID = 1
				assert.Equal(t, uint(1), GetID(user), description)
			},
		},
		{
			description: "return nil",
			setup: func(description string) {
				type User struct {
					Name   string
					Avatar string
				}
				user := User{}
				assert.Nil(t, GetID(user), description)
			},
		},
		{
			description: "return nil when model is nil",
			setup: func(description string) {
				type User struct {
					Name   string
					Avatar string
				}
				assert.Nil(t, GetID(&User{}), description)
				assert.Nil(t, GetID(nil), description)
			},
		},
	}
	for _, test := range tests {
		test.setup(test.description)
	}
}

func TestGetIDByReflect(t *testing.T) {
	tests := []struct {
		description string
		setup       func(description string)
	}{
		{
			description: "TestStruct.ID type int",
			setup: func(description string) {
				ts := TestStruct{ID: 1, name: "name"}
				v := reflect.ValueOf(ts)
				tpe := reflect.TypeOf(ts)

				result := GetIDByReflect(tpe, v)

				assert.Equal(t, 1, result)
			},
		},
		{
			description: "TestStruct.ID type string",
			setup: func(description string) {
				ts := TestStructString{ID: "goravel", name: "name"}
				v := reflect.ValueOf(ts)
				tpe := reflect.TypeOf(ts)

				result := GetIDByReflect(tpe, v)

				assert.Equal(t, "goravel", result)
			},
		},
		{
			description: "TestStruct.ID type UUID",
			setup: func(description string) {
				id := uuid.New()
				ts := TestStructUUID{ID: id, name: "name"}
				v := reflect.ValueOf(ts)
				tpe := reflect.TypeOf(ts)

				result := GetIDByReflect(tpe, v)

				assert.Equal(t, id, result)
			},
		},
		{
			description: "TestStruct without primaryKey",
			setup: func(description string) {
				ts := TestStructNoPK{ID: 1, name: "name"}
				v := reflect.ValueOf(ts)
				tpe := reflect.TypeOf(ts)

				result := GetIDByReflect(tpe, v)

				assert.Nil(t, result)
			},
		},
	}
	for _, test := range tests {
		test.setup(test.description)
	}
}
