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
			description: "success",
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
			description: "success with orm.Model",
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
			description: "error",
			setup: func(description string) {
				type User struct {
					Name   string
					Avatar string
				}
				user := User{}
				assert.Nil(t, GetID(&user), description)
			},
		},
	}
	for _, test := range tests {
		test.setup(test.description)
	}
}
