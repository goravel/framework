package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
)

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
			assert.Equal(t, test.expected, IsPlainIdentifier(test.column))
		})
	}
}

func TestIsJsonSelector(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		expected bool
	}{
		{name: "json selector", column: "data->name", expected: true},
		{name: "table qualified json selector", column: "users.data->name", expected: true},
		{name: "nested json selector", column: "data->nested->name", expected: true},
		{name: "plain column", column: "data", expected: false},
		{name: "trailing arrow", column: "data->", expected: false},
		{name: "single dash", column: "data-name", expected: false},
		{name: "function", column: "LOWER(data->name)", expected: false},
		{name: "with direction", column: "data->name desc", expected: false},
		{name: "injection", column: "data->name; DROP TABLE users", expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, IsJsonSelector(test.column))
		})
	}
}

func TestNormalizeOrderDirection(t *testing.T) {
	tests := []struct {
		name               string
		direction          string
		expectedNormalized string
		expectedError      error
	}{
		{name: "empty means ascending", direction: "", expectedNormalized: "ASC"},
		{name: "lower case ascending", direction: "asc", expectedNormalized: "ASC"},
		{name: "descending with surrounding spaces", direction: "  Desc  ", expectedNormalized: "DESC"},
		{name: "nulls last", direction: "desc nulls last", expectedError: errors.DatabaseInvalidOrderByDirection.Args("desc nulls last")},
		{name: "nulls first", direction: "ASC NULLS FIRST", expectedError: errors.DatabaseInvalidOrderByDirection.Args("ASC NULLS FIRST")},
		{name: "two directions", direction: "ASC DESC", expectedError: errors.DatabaseInvalidOrderByDirection.Args("ASC DESC")},
		{name: "injection", direction: "DESC; DROP TABLE users", expectedError: errors.DatabaseInvalidOrderByDirection.Args("DESC; DROP TABLE users")},
		{name: "comma separated", direction: "asc, desc", expectedError: errors.DatabaseInvalidOrderByDirection.Args("asc, desc")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			normalized, err := NormalizeOrderDirection(test.direction)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
				assert.Empty(t, normalized)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expectedNormalized, normalized)
		})
	}
}
