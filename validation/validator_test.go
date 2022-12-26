package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBind(t *testing.T) {
	var maker *Validation
	type Data struct {
		A string
		B int
		C string
	}

	tests := []struct {
		describe   string
		data       any
		rules      map[string]string
		expectData Data
		expectErr  error
	}{
		{
			describe: "success when data is map and key is lowercase",
			data:     map[string]any{"a": "aa"},
			rules:    map[string]string{"a": "required"},
			expectData: Data{
				A: "aa",
			},
		},
		{
			describe: "success when data is map, key is lowercase and has errors",
			data:     map[string]any{"a": "aa", "c": "cc"},
			rules:    map[string]string{"a": "required", "b": "required"},
			expectData: Data{
				A: "aa",
				C: "cc",
			},
		},
		{
			describe: "success when data is map and key is uppercase",
			data:     map[string]any{"A": "aa"},
			rules:    map[string]string{"A": "required"},
			expectData: Data{
				A: "aa",
			},
		},
		{
			describe: "success when data is struct and key is uppercase",
			data: struct {
				A string
			}{
				A: "aa",
			},
			rules: map[string]string{"A": "required"},
			expectData: Data{
				A: "aa",
			},
		},
		{
			describe: "empty when data is struct and key is lowercase",
			data: struct {
				a string
			}{
				a: "aa",
			},
			rules:      map[string]string{"a": "required"},
			expectData: Data{},
		},
	}

	for _, test := range tests {
		maker = NewValidation()
		validator, err := maker.Make(
			test.data,
			test.rules,
		)
		assert.Nil(t, err)
		var data Data
		err = validator.Bind(&data)
		assert.Nil(t, test.expectErr, err, test.describe)
		assert.Equal(t, test.expectData, data, test.describe)
	}
}

func TestFails(t *testing.T) {
	var maker *Validation
	tests := []struct {
		describe  string
		data      any
		rules     map[string]string
		expectRes bool
	}{
		{
			describe: "false",
			data:     map[string]any{"a": "aa"},
			rules:    map[string]string{"a": "required"},
		},
		{
			describe:  "true",
			data:      map[string]any{"b": "bb"},
			rules:     map[string]string{"a": "required"},
			expectRes: true,
		},
	}

	for _, test := range tests {
		maker = NewValidation()
		validator, err := maker.Make(
			test.data,
			test.rules,
		)
		assert.Nil(t, err)
		assert.Equal(t, test.expectRes, validator.Fails(), test.describe)
	}
}
