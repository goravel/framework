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
		name     string
		column   string
		expected bool
	}{
		{name: "column", column: "group", expected: true},
		{name: "column with underscore", column: "user_id", expected: true},
		{name: "leading underscore", column: "_id", expected: true},
		{name: "mixed case", column: "Name", expected: true},
		{name: "table qualified column", column: "users.group", expected: true},
		{name: "schema qualified column", column: "public.users.group", expected: true},
		{name: "empty string", column: "", expected: false},
		{name: "leading digit", column: "1id", expected: false},
		{name: "trailing dot", column: "users.", expected: false},
		{name: "leading dot", column: ".group", expected: false},
		{name: "function", column: "LOWER(name)", expected: false},
		{name: "json selector", column: "data->name", expected: false},
		{name: "double quoted", column: `"group"`, expected: false},
		{name: "backtick quoted", column: "`group`", expected: false},
		{name: "with direction", column: "id desc", expected: false},
		{name: "star", column: "*", expected: false},
		{name: "table star", column: "users.*", expected: false},
		{name: "injection", column: "id; DROP TABLE users", expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, isPlainIdentifier(test.column))
		})
	}
}
