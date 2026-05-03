package validation

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpvalidate "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/http"
)

func TestMake(t *testing.T) {
	type Data struct {
		A string `form:"a"`
	}

	ctx := http.NewContext()
	// nolint:all
	ctx.WithValue("test", "test")

	tests := []struct {
		description        string
		data               any
		rules              map[string]any
		options            []httpvalidate.Option
		expectValidator    bool
		expectErr          error
		expectData         Data
		expectErrors       bool
		expectErrorMessage string
	}{
		{
			description: "success when data is map[string]any",
			data:        map[string]any{"a": " b "},
			rules:       map[string]any{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]any{"a": "trim"}),
			},
			expectValidator: true,
			expectData:      Data{A: "b"},
		},
		{
			description: "success when data is struct",
			data:        &Data{A: "  b"},
			rules:       map[string]any{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]any{"a": "trim"}),
			},
			expectValidator: true,
			expectData:      Data{A: "b"},
		},
		{
			description:        "error when data is empty map",
			data:               map[string]any{},
			rules:              map[string]any{"a": "required"},
			expectValidator:    true,
			expectErrors:       true,
			expectErrorMessage: "The a field is required.",
		},
		{
			description: "error when data isn't map[string]any or map[string][]string or struct",
			data:        "1   ",
			rules:       map[string]any{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]any{"a": "trim"}),
			},
			expectErr: errors.ValidationDataInvalidType,
		},
		{
			description: "error when rule is empty map",
			data:        map[string]any{"a": "b"},
			rules:       map[string]any{},
			expectErr:   errors.ValidationEmptyRules,
		},
		{
			description: "error when PrepareForValidation returns error",
			data:        map[string]any{"a": "   b   "},
			rules:       map[string]any{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]any{"a": "trim"}),
				PrepareForValidation(func(ctx context.Context, data httpvalidate.Data) error {
					return assert.AnError
				}),
			},
			expectErr: assert.AnError,
		},
		{
			description: "success when data is map[string]any and with PrepareForValidation",
			data:        map[string]any{"a": "   b  "},
			rules:       map[string]any{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]any{"a": "trim"}),
				PrepareForValidation(func(ctx context.Context, data httpvalidate.Data) error {
					if _, exist := data.Get("a"); exist {
						return data.Set("a", "c")
					}
					return nil
				}),
			},
			expectValidator: true,
			expectData:      Data{A: "c"},
		},
		{
			description: "success when calling PrepareForValidation with ctx",
			data:        map[string]any{"a": "   b  "},
			rules:       map[string]any{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]any{"a": "trim"}),
				PrepareForValidation(func(ctx context.Context, data httpvalidate.Data) error {
					if _, exist := data.Get("a"); exist {
						return data.Set("a", ctx.Value("test"))
					}

					return nil
				}),
			},
			expectValidator: true,
			expectData:      Data{A: "test"},
		},
		{
			description: "contain errors when data is map[string]any and with Messages, Attributes, PrepareForValidation",
			data:        map[string]any{"a": "aa   "},
			rules:       map[string]any{"a": "required", "b": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]any{"a": "trim", "b": "trim"}),
				Messages(map[string]string{
					"b.required": ":attribute can't be empty",
				}),
				Attributes(map[string]string{
					"b": "B",
				}),
				PrepareForValidation(func(ctx context.Context, data httpvalidate.Data) error {
					if _, exist := data.Get("a"); exist {
						return data.Set("a", "c")
					}
					return nil
				}),
			},
			expectValidator:    true,
			expectData:         Data{A: ""},
			expectErrors:       true,
			expectErrorMessage: "B can't be empty",
		},
		{
			description: "success when data is struct and with PrepareForValidation",
			data:        &Data{A: "b"},
			rules:       map[string]any{"a": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]any{"a": "trim"}),
				PrepareForValidation(func(ctx context.Context, data httpvalidate.Data) error {
					if _, exist := data.Get("a"); exist {
						return data.Set("a", "c")
					}
					return nil
				}),
			},
			expectValidator: true,
			expectData:      Data{A: "c"},
		},
		{
			description: "contain errors when data is struct and with Messages, Attributes, PrepareForValidation",
			data:        &Data{A: "b"},
			rules:       map[string]any{"a": "required", "b": "required"},
			options: []httpvalidate.Option{
				Filters(map[string]any{"a": "trim", "b": "trim"}),
				Messages(map[string]string{
					"b.required": ":attribute can't be empty",
				}),
				Attributes(map[string]string{
					"b": "b",
				}),
				PrepareForValidation(func(ctx context.Context, data httpvalidate.Data) error {
					if _, exist := data.Get("a"); exist {
						return data.Set("a", "c")
					}
					return nil
				}),
			},
			expectValidator:    true,
			expectData:         Data{A: ""},
			expectErrors:       true,
			expectErrorMessage: "b can't be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			validation := NewValidation()
			validator, err := validation.Make(ctx, test.data, test.rules, test.options...)
			assert.Equal(t, test.expectValidator, validator != nil, test.description)
			if test.expectErr != nil {
				assert.ErrorIs(t, err, test.expectErr, test.description)
			}

			if validator != nil {
				var data Data
				err = validator.Bind(&data)
				assert.Nil(t, err, test.description)
				assert.Equal(t, test.expectData, data, test.description)
				if validator.Fails() {
					assert.Equal(t, test.expectErrorMessage, validator.Errors().One(), test.description)
				}
				assert.Equal(t, test.expectErrors, validator.Fails(), test.description)
			}
		})
	}
}

// Fix: https://github.com/goravel/goravel/issues/533
func TestBindWithNestedStruct(t *testing.T) {
	type Data struct {
		A map[string][]string `json:"a" form:"a"`
		B map[string][]string `json:"b" form:"b"`
	}
	validation := NewValidation()
	validator, err := validation.Make(context.Background(), map[string]any{
		"a": map[string]any{
			"b": []any{"c", "d"},
		},
		"b": map[string][]string{
			"b": {"c", "d"},
		},
	}, map[string]any{"a": "required|map", "b": "required|map"})

	require.NoError(t, err)
	require.NotNil(t, validator)
	require.False(t, validator.Fails())

	var data Data
	require.NoError(t, validator.Bind(&data))
	require.Equal(t, Data{
		A: map[string][]string{
			"b": {"c", "d"},
		},
		B: map[string][]string{
			"b": {"c", "d"},
		},
	}, data)
}

type Case struct {
	description string
	setup       func(Case)
}

func TestRule_Regex(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success with valid regex match",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"email": "test@example.com",
				}, map[string]any{
					"email": `regex:^\S+@\S+\.\S+$`,
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error with invalid regex match",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"email": "testexample.com",
				}, map[string]any{
					"email": `regex:^\S+@\S+\.\S+$`,
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.True(t, validator.Fails(), c.description)
				assert.Equal(t, map[string]string{
					"regex": "The email field format is invalid.",
				}, validator.Errors().Get("email"))
			},
		},
		{
			description: "success with regex and nested structure",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"user": map[string]any{
						"email": "test@example.com",
					},
				}, map[string]any{
					"user.email": `regex:^\S+@\S+\.\S+$`,
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error with regex and nested structure",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"user": map[string]any{
						"email": "testexample.com",
					},
				}, map[string]any{
					"user.email": `regex:^\S+@\S+\.\S+$`,
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.True(t, validator.Fails(), c.description)
				assert.NotEmpty(t, validator.Errors().Get("user.email"))
			},
		},
		{
			description: "error when regex pattern is empty",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"email": "test@example.com",
				}, map[string]any{
					"email": "regex:",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.True(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error with invalid regex match",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"phone": "18005555555",
				}, map[string]any{
					"phone": "regex:^\\+\\d{1,3}-\\d{3}-\\d{3}-\\d{4}$",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.True(t, validator.Fails(), c.description)
				assert.NotEmpty(t, validator.Errors().Get("phone"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Required(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "goravel",
				}, map[string]any{
					"name": "required",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "success with nested",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": map[string]any{
						"first": "Goravel",
					},
				}, map[string]any{
					"name.first": "required",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when key is empty",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "",
				}, map[string]any{
					"name": "required",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.True(t, validator.Fails(), c.description)
				assert.Equal(t, "The name field is required.", validator.Errors().One("name"))
			},
		},
		{
			description: "error when key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "Goravel",
				}, map[string]any{
					"name":  "required",
					"name1": "required",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.True(t, validator.Fails(), c.description)
				assert.Equal(t, "The name1 field is required.", validator.Errors().One("name1"))
			},
		},
		{
			description: "error when nested",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": map[string]string{
						"first": "",
					},
				}, map[string]any{
					"name.first": "required",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.True(t, validator.Fails(), c.description)
				assert.Equal(t, map[string]string{
					"required": "The name.first field is required.",
				}, validator.Errors().Get("name.first"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_RequiredIf(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success when required_if is true",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
				}, map[string]any{
					"name":  "required",
					"name1": "required_if:name,goravel,goravel1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "success when required_if is false",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "goravel2",
				}, map[string]any{
					"name":  "required",
					"name1": "required_if:name,goravel,goravel1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when required_if is true and key is empty",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name1": "",
				}, map[string]any{
					"name":  "required",
					"name1": "required_if:name,goravel,goravel1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.True(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when required_if is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "goravel",
				}, map[string]any{
					"name":  "required",
					"name1": "required_if:name,goravel,goravel1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.True(t, validator.Fails(), c.description)
				assert.Equal(t, map[string]string{
					"required_if": "The name1 field is required when name is goravel, goravel1.",
				}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_RequiredUnless(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success when required_unless is true",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
				}, map[string]any{
					"name":  "required",
					"name1": "required_unless:name,hello,hello1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "success when required_unless is false",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "goravel",
				}, map[string]any{
					"name":  "required",
					"name1": "required_unless:name,goravel,goravel1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when required_unless is true and key is empty",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name1": "",
				}, map[string]any{
					"name":  "required",
					"name1": "required_unless:name,hello,hello1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_unless": "The name1 field is required unless name is in hello, hello1.",
				}, validator.Errors().Get("name1"))
			},
		},
		{
			description: "error when required_unless is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "goravel",
				}, map[string]any{
					"name":  "required",
					"name1": "required_unless:name,hello,hello1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_unless": "The name1 field is required unless name is in hello, hello1.",
				}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_RequiredWith(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success when required_with is true",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name2": "goravel2",
				}, map[string]any{
					"name":  "required",
					"name2": "required_with:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "success when required_with is false",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "",
				}, map[string]any{
					"name": "required_with:name1,name2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when required_with is true and key is empty",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
					"name2": "",
				}, map[string]any{
					"name":  "required",
					"name1": "required",
					"name2": "required_with:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_with": "The name2 field is required when name, name1 is present.",
				}, validator.Errors().Get("name2"))
			},
		},
		{
			description: "error when required_with is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
				}, map[string]any{
					"name":  "required",
					"name1": "required",
					"name2": "required_with:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_with": "The name2 field is required when name, name1 is present.",
				}, validator.Errors().Get("name2"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_RequiredWithAll(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success when required_with_all is true",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
					"name2": "goravel2",
				}, map[string]any{
					"name":  "required",
					"name1": "required",
					"name2": "required_with_all:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "success when not all fields present",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name1": "",
					"name2": "goravel2",
				}, map[string]any{
					"name":  "required",
					"name2": "required_with_all:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "success when required_with_all is false",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "",
				}, map[string]any{
					"name": "required_with_all:name1,name2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when required_with_all is true and key is empty",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
					"name2": "",
				}, map[string]any{
					"name":  "required",
					"name1": "required",
					"name2": "required_with_all:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_with_all": "The name2 field is required when name, name1 are present.",
				}, validator.Errors().Get("name2"))
			},
		},
		{
			description: "error when required_with_all is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
				}, map[string]any{
					"name":  "required",
					"name1": "required",
					"name2": "required_with_all:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_with_all": "The name2 field is required when name, name1 are present.",
				}, validator.Errors().Get("name2"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_RequiredWithout(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success when required_without is true",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name2": "goravel2",
				}, map[string]any{
					"name":  "required",
					"name2": "required_without:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "success when required_without is false",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
					"name2": "goravel2",
				}, map[string]any{
					"name": "required_without:name1,name2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when required_without is true and key is empty",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "goravel",
					"name2": "",
				}, map[string]any{
					"name":  "required",
					"name2": "required_without:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_without": "The name2 field is required when name, name1 is not present.",
				}, validator.Errors().Get("name2"))
			},
		},
		{
			description: "error when required_without is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "goravel",
				}, map[string]any{
					"name":  "required",
					"name2": "required_without:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_without": "The name2 field is required when name, name1 is not present.",
				}, validator.Errors().Get("name2"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_RequiredWithoutAll(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success when required_without_all is true",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "goravel",
				}, map[string]any{
					"name": "required_without_all:name1,name2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "success when required_without_all is false",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name":  "",
					"name1": "goravel1",
				}, map[string]any{
					"name": "required_without_all:name1,name2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when required_without_all is true and key is empty",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name": "",
				}, map[string]any{
					"name": "required_without_all:name1,name2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_without_all": "The name field is required when none of name1, name2 are present.",
				}, validator.Errors().Get("name"))
			},
		},
		{
			description: "error when required_without_all is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name3": "goravel3",
				}, map[string]any{
					"name": "required_without_all:name1,name2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_without_all": "The name field is required when none of name1, name2 are present.",
				}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestAddRules(t *testing.T) {
	validation := NewValidation()

	t.Run("success", func(t *testing.T) {
		err := validation.AddRules([]httpvalidate.Rule{&CustomUppercase{}})
		assert.Nil(t, err)
	})

	t.Run("duplicate rule", func(t *testing.T) {
		err := validation.AddRules([]httpvalidate.Rule{&Duplicate{}})
		assert.EqualError(t, err, "duplicate rule name: required")
	})
}

func TestCustomFilters(t *testing.T) {
	validation := NewValidation()
	err := validation.AddFilters([]httpvalidate.Filter{&DefaultFilter{}})
	assert.Nil(t, err)

	filters := validation.Filters()
	defaultFilterFunc := filters[0].Handle(context.Background()).(func(string, ...string) string)
	assert.Equal(t, "default", defaultFilterFunc("", "default"))
	assert.Equal(t, "a", defaultFilterFunc("a"))
}

func TestCustomFiltersIntegration(t *testing.T) {
	mp := map[string]any{
		"name":  "krishan ",
		"age":   " 22 ",
		"empty": "",
	}

	validation := NewValidation()
	err := validation.AddFilters([]httpvalidate.Filter{&DefaultFilter{}})
	assert.Nil(t, err)

	validator, err := validation.Make(context.Background(), mp, map[string]any{
		"name":  "required",
		"age":   "required",
		"empty": "required",
	}, Filters(map[string]any{
		"empty": "default:emptyDefault",
		"name":  "trim|upper",
		"age":   "trim|to_int",
	}))

	assert.Nil(t, err)
	var newMp map[string]any
	assert.Nil(t, validator.Bind(&newMp))

	assert.Equal(t, "KRISHAN", newMp["name"])
	assert.Equal(t, 22, newMp["age"])
	assert.Equal(t, "emptyDefault", newMp["empty"])
}

func TestCustomRule(t *testing.T) {
	validation := NewValidation()
	err := validation.AddRules([]httpvalidate.Rule{&CustomUppercase{}, &CustomLowercase{}})
	assert.Nil(t, err)

	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name1":   "on",
					"name2":   "off",
					"name3":   "yes",
					"name4":   "no",
					"name5":   true,
					"name6":   false,
					"name7":   "true",
					"name8":   "false",
					"name9":   "1",
					"name10":  "0",
					"name":    "ABC",
					"address": "de",
				}, map[string]any{
					"name1":   "bool",
					"name2":   "bool",
					"name3":   "bool",
					"name4":   "bool",
					"name5":   "bool",
					"name6":   "bool",
					"name7":   "bool",
					"name8":   "bool",
					"name9":   "bool",
					"name10":  "bool",
					"name":    "required|custom_uppercase:3",
					"address": "required|custom_lowercase:2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(context.Background(), map[string]any{
					"name1":   1,
					"name2":   0,
					"name3":   "a",
					"name":    "abc",
					"address": "DE",
				}, map[string]any{
					"name1":   "bool",
					"name2":   "bool",
					"name3":   "bool",
					"name":    "required|custom_uppercase:3",
					"address": "required|custom_lowercase:2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"custom_uppercase": "name must be upper"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"custom_lowercase": "address must be lower"}, validator.Errors().Get("address"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestValidated(t *testing.T) {
	validation := NewValidation()
	validator, err := validation.Make(context.Background(), map[string]any{
		"name":  "goravel",
		"email": "test@example.com",
		"extra": "not in rules",
	}, map[string]any{
		"name":  "required",
		"email": "required|email",
	})
	assert.Nil(t, err)
	assert.False(t, validator.Fails())

	validated := validator.Validated()
	assert.Equal(t, "goravel", validated["name"])
	assert.Equal(t, "test@example.com", validated["email"])
	// "extra" should not be in validated data
	_, exists := validated["extra"]
	assert.False(t, exists)
}

func TestMapRules(t *testing.T) {
	validation := NewValidation()

	t.Run("map rule fails when nested key does not exist", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"users": map[string]any{},
		}, map[string]any{
			"users.name": "required",
		})
		assert.Nil(t, err)
		assert.True(t, validator.Fails())
		assert.Equal(t, map[string]string{
			"required": "The users.name field is required.",
		}, validator.Errors().Get("users.name"))
	})

	t.Run("map rule fails when nested key is empty", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"users": map[string]any{
				"name": "",
			},
		}, map[string]any{
			"users.name": "required",
		})
		assert.Nil(t, err)
		assert.True(t, validator.Fails())
		assert.Equal(t, map[string]string{
			"required": "The users.name field is required.",
		}, validator.Errors().Get("users.name"))
	})

	t.Run("map rule succeeds when nested key is present", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"users": map[string]any{
				"name": "Alice",
			},
		}, map[string]any{
			"users.name": "required",
		})
		assert.Nil(t, err)
		assert.False(t, validator.Fails())
	})
}

func TestWildcardRules(t *testing.T) {
	validation := NewValidation()

	t.Run("validates wildcard fields", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"users": []any{
				map[string]any{"name": "Alice"},
				map[string]any{"name": ""},
			},
		}, map[string]any{
			"users.*.name": "required",
		})
		assert.Nil(t, err)
		assert.True(t, validator.Fails())
	})

	t.Run("success with all valid", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"users": []any{
				map[string]any{"name": "Alice"},
				map[string]any{"name": "Bob"},
			},
		}, map[string]any{
			"users.*.name": "required|string",
		})
		assert.Nil(t, err)
		assert.False(t, validator.Fails())
	})

	t.Run("validates wildcard fields with typed string slice", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"scores": []string{"a", "b"},
		}, map[string]any{
			"scores.*": "required|string",
		})
		assert.Nil(t, err)
		assert.False(t, validator.Fails())
	})

	t.Run("validates wildcard fields with typed int slice", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"scores": []int{1, 2},
		}, map[string]any{
			"scores.*": "required|int",
		})
		assert.Nil(t, err)
		assert.False(t, validator.Fails())
	})

	t.Run("validates wildcard fields with []any primitive array", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"scores": []any{float64(1), float64(2)},
		}, map[string]any{
			"scores.*": "required|int",
		})
		assert.Nil(t, err)
		assert.False(t, validator.Fails())
	})
}

func TestSliceRuleSyntax(t *testing.T) {
	validation := NewValidation()

	t.Run("regex with pipe in pattern using slice syntax", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"code": "foo",
		}, map[string]any{
			"code": []string{"required", "regex:^(foo|bar)$", "string"},
		})
		assert.Nil(t, err)
		assert.NotNil(t, validator)
		assert.False(t, validator.Fails())
	})

	t.Run("regex with pipe fails validation using slice syntax", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"code": "baz",
		}, map[string]any{
			"code": []string{"required", "regex:^(foo|bar)$"},
		})
		assert.Nil(t, err)
		assert.NotNil(t, validator)
		assert.True(t, validator.Fails())
	})

	t.Run("mixed string and slice rules", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"name": "goravel",
			"code": "foo",
		}, map[string]any{
			"name": "required|string",
			"code": []string{"required", "regex:^(foo|bar)$"},
		})
		assert.Nil(t, err)
		assert.NotNil(t, validator)
		assert.False(t, validator.Fails())
	})

	t.Run("invalid rule type returns error", func(t *testing.T) {
		_, err := validation.Make(context.Background(), map[string]any{
			"name": "goravel",
		}, map[string]any{
			"name": 123,
		})
		assert.ErrorIs(t, err, errors.ValidationInvalidRuleType)
	})

	t.Run("slice syntax with filters", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"name": "  Goravel  ",
		}, map[string]any{
			"name": []string{"required", "string"},
		}, Filters(map[string]any{
			"name": []string{"trim", "lower"},
		}))
		assert.Nil(t, err)
		assert.NotNil(t, validator)
		assert.False(t, validator.Fails())

		val := validator.Validated()
		assert.Equal(t, "goravel", val["name"])
	})
}

type CustomUppercase struct {
}

func (receiver *CustomUppercase) Signature() string {
	return "custom_uppercase"
}

func (receiver *CustomUppercase) Passes(ctx context.Context, data httpvalidate.Data, val any, options ...any) bool {
	name, exist := data.Get("name")

	if len(options) > 0 {
		return strings.ToUpper(val.(string)) == val.(string) && len(val.(string)) == cast.ToInt(options[0]) && name == val && exist
	}

	return false
}

func (receiver *CustomUppercase) Message(ctx context.Context) string {
	return ":attribute must be upper"
}

type CustomLowercase struct {
}

func (receiver *CustomLowercase) Signature() string {
	return "custom_lowercase"
}

func (receiver *CustomLowercase) Passes(ctx context.Context, data httpvalidate.Data, val any, options ...any) bool {
	address, exist := data.Get("address")

	if len(options) > 0 {
		return strings.ToLower(val.(string)) == val.(string) && len(val.(string)) == cast.ToInt(options[0]) && address == val && exist
	}

	return false
}

func (receiver *CustomLowercase) Message(ctx context.Context) string {
	return ":attribute must be lower"
}

type Duplicate struct {
}

func (receiver *Duplicate) Signature() string {
	return "required"
}

func (receiver *Duplicate) Passes(ctx context.Context, data httpvalidate.Data, val any, options ...any) bool {
	return true
}

func (receiver *Duplicate) Message(ctx context.Context) string {
	return ""
}

type DefaultFilter struct {
}

func (receiver *DefaultFilter) Signature() string {
	return "default"
}

func (receiver *DefaultFilter) Handle(ctx context.Context) any {
	return func(val string, def ...string) string {
		if val == "" {
			if len(def) > 0 {
				return def[0]
			}
		}

		return val
	}
}
