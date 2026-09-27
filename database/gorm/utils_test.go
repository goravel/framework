package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyStruct(t *testing.T) {
	type Data struct {
		Name string
		age  int
	}

	data := copyStruct(Data{Name: "name", age: 18})

	assert.Equal(t, "name", data.Field(0).Interface().(string))
	assert.Panics(t, func() {
		data.Field(1).Interface()
	})
}

func TestIsPlainIdentifier(t *testing.T) {
	tests := []struct {
		column   string
		expected bool
	}{
		{column: "group", expected: true},
		{column: "user_id", expected: true},
		{column: "_id", expected: true},
		{column: "Name", expected: true},
		{column: "users.group", expected: true},
		{column: "public.users.group", expected: true},
		{column: "", expected: false},
		{column: "1id", expected: false},
		{column: "users.", expected: false},
		{column: ".group", expected: false},
		{column: "LOWER(name)", expected: false},
		{column: "data->name", expected: false},
		{column: `"group"`, expected: false},
		{column: "`group`", expected: false},
		{column: "id desc", expected: false},
		{column: "*", expected: false},
		{column: "users.*", expected: false},
		{column: "id; DROP TABLE users", expected: false},
	}

	for _, test := range tests {
		t.Run(test.column, func(t *testing.T) {
			assert.Equal(t, test.expected, isPlainIdentifier(test.column))
		})
	}
}
