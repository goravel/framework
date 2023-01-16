package database

import (
	"testing"

	"github.com/goravel/framework/database/orm"

	"github.com/stretchr/testify/assert"
)

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
