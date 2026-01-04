package validation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	httpvalidate "github.com/goravel/framework/contracts/validation"
)

func TestOne(t *testing.T) {
	var maker *Validation
	tests := []struct {
		describe  string
		data      any
		rules     map[string]string
		options   []httpvalidate.Option
		expectRes string
	}{
		{
			describe: "errors is empty",
			data:     map[string]any{"a": "aa"},
			rules:    map[string]string{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]string{"a": "trim"}),
			},
		},
		{
			describe: "errors isn't empty",
			data:     map[string]any{"a": ""},
			rules:    map[string]string{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]string{"a": "trim"}),
			},
			expectRes: "a is required to not be empty",
		},
		{
			describe: "errors isn't empty when setting messages option",
			data:     map[string]any{"a": ""},
			rules:    map[string]string{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]string{"a": "trim"}),
				Messages(map[string]string{"a.required": "a can't be empty"}),
			},
			expectRes: "a can't be empty",
		},
		{
			describe: "errors isn't empty when setting attributes option",
			data:     map[string]any{"a": ""},
			rules:    map[string]string{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]string{"a": "trim"}),
				Attributes(map[string]string{"a": "aa"}),
			},
			expectRes: "aa is required to not be empty",
		},
		{
			describe: "errors isn't empty when setting messages and attributes option",
			data:     map[string]any{"a": ""},
			rules:    map[string]string{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]string{"a": "trim"}),
				Messages(map[string]string{"a.required": ":attribute can't be empty"}),
				Attributes(map[string]string{"a": "aa"}),
			},
			expectRes: "aa can't be empty",
		},
	}

	for _, test := range tests {
		maker = NewValidation()
		validator, err := maker.Make(
			context.Background(),
			test.data,
			test.rules,
			test.options...,
		)

		assert.Nil(t, err, test.describe)
		assert.NotNil(t, validator, test.describe)

		if test.expectRes != "" {
			errors := validator.Errors()
			assert.NotNil(t, errors)
			assert.Equal(t, test.expectRes, errors.One(), test.describe)
		}
	}
}

func TestGet(t *testing.T) {
	var maker *Validation
	tests := []struct {
		describe string
		data     any
		rules    map[string]string
		filters  map[string]string
		expectA  map[string]string
		expectB  map[string]string
	}{
		{
			describe: "errors is empty",
			data:     map[string]any{"a": "aa", "b": "bb"},
			rules:    map[string]string{"a": "required", "b": "required"},
			filters:  map[string]string{"a": "trim", "b": "trim"},
		},
		{
			describe: "errors isn't empty",
			data:     map[string]any{"c": "cc"},
			rules:    map[string]string{"a": "required", "b": "required"},
			filters:  map[string]string{"a": "trim", "b": "trim"},
			expectA:  map[string]string{"required": "a is required to not be empty"},
			expectB:  map[string]string{"required": "b is required to not be empty"},
		},
	}

	for _, test := range tests {
		maker = NewValidation()
		validator, err := maker.Make(
			context.Background(),
			test.data,
			test.rules,
			Filters(test.filters),
		)
		assert.Nil(t, err, test.describe)
		if len(test.expectA) > 0 {
			errors := validator.Errors()
			assert.NotNil(t, errors)
			assert.Equal(t, test.expectA, errors.Get("a"), test.describe)
		}
		if len(test.expectB) > 0 {
			errors := validator.Errors()
			assert.NotNil(t, errors)
			assert.Equal(t, test.expectB, errors.Get("b"), test.describe)
		}
	}
}

func TestAll(t *testing.T) {
	var maker *Validation
	tests := []struct {
		describe  string
		data      any
		rules     map[string]string
		filters   map[string]string
		expectRes map[string]map[string]string
	}{
		{
			describe:  "errors is empty",
			data:      map[string]any{"a": "aa", "b": "bb"},
			rules:     map[string]string{"a": "required", "b": "required"},
			filters:   map[string]string{"a": "trim", "b": "trim"},
			expectRes: map[string]map[string]string{},
		},
		{
			describe: "errors isn't empty",
			data:     map[string]any{"c": "cc"},
			rules:    map[string]string{"a": "required", "b": "required"},
			filters:  map[string]string{"a": "trim", "b": "trim"},
			expectRes: map[string]map[string]string{
				"a": {"required": "a is required to not be empty"},
				"b": {"required": "b is required to not be empty"},
			},
		},
	}

	for _, test := range tests {
		maker = NewValidation()
		validator, err := maker.Make(
			context.Background(),
			test.data,
			test.rules,
			Filters(test.filters),
		)
		assert.Nil(t, err, test.describe)
		if len(test.expectRes) > 0 {
			errors := validator.Errors()
			assert.NotNil(t, errors)
			assert.Equal(t, test.expectRes, errors.All(), test.describe)
		}
	}
}

func TestHas(t *testing.T) {
	var maker *Validation
	tests := []struct {
		describe  string
		data      any
		rules     map[string]string
		filters   map[string]string
		expectRes bool
	}{
		{
			describe: "errors is empty",
			data:     map[string]any{"a": "aa", "b": "bb"},
			rules:    map[string]string{"a": "required", "b": "required"},
			filters:  map[string]string{"a": "trim", "b": "trim"},
		},
		{
			describe:  "errors isn't empty",
			data:      map[string]any{"c": "cc"},
			rules:     map[string]string{"a": "required", "b": "required"},
			filters:   map[string]string{"a": "trim", "b": "trim"},
			expectRes: true,
		},
	}

	for _, test := range tests {
		maker = NewValidation()
		validator, err := maker.Make(
			context.Background(),
			test.data,
			test.rules,
			Filters(test.filters),
		)
		assert.Nil(t, err, test.describe)
		if test.expectRes {
			errors := validator.Errors()
			assert.NotNil(t, errors)
			assert.Equal(t, test.expectRes, errors.Has("a"), test.describe)
		}
	}
}
