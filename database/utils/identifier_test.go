package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestNormalizeOrderDirection(t *testing.T) {
	tests := []struct {
		name               string
		direction          string
		expectedNormalized string
		expectedOK         bool
	}{
		{name: "empty means ascending", direction: "", expectedNormalized: "ASC", expectedOK: true},
		{name: "lower case ascending", direction: "asc", expectedNormalized: "ASC", expectedOK: true},
		{name: "descending with surrounding spaces", direction: "  Desc  ", expectedNormalized: "DESC", expectedOK: true},
		{name: "nulls last", direction: "desc nulls last", expectedOK: false},
		{name: "nulls first", direction: "ASC NULLS FIRST", expectedOK: false},
		{name: "two directions", direction: "ASC DESC", expectedOK: false},
		{name: "injection", direction: "DESC; DROP TABLE users", expectedOK: false},
		{name: "comma separated", direction: "asc, desc", expectedOK: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			normalized, ok := NormalizeOrderDirection(test.direction)

			assert.Equal(t, test.expectedOK, ok)
			assert.Equal(t, test.expectedNormalized, normalized)
		})
	}
}
