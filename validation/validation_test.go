package validation

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	httpvalidate "github.com/goravel/framework/contracts/validation"
)

func TestMake(t *testing.T) {
	type Data struct {
		A string
	}

	tests := []struct {
		description        string
		data               any
		rules              map[string]string
		options            []httpvalidate.Option
		expectValidator    bool
		expectErr          error
		expectData         Data
		expectErrors       bool
		expectErrorMessage string
	}{
		{
			description:     "success when data is map[string]any",
			data:            map[string]any{"a": "b"},
			rules:           map[string]string{"a": "required"},
			expectValidator: true,
			expectData:      Data{A: "b"},
		},
		{
			description:     "success when data is struct",
			data:            &Data{A: "b"},
			rules:           map[string]string{"A": "required"},
			expectValidator: true,
			expectData:      Data{A: "b"},
		},
		{
			description: "error when data isn't map[string]any or struct",
			data:        "1",
			rules:       map[string]string{"a": "required"},
			expectErr:   errors.New("data must be map[string]any or struct"),
		},
		{
			description: "error when data is empty map",
			data:        map[string]any{},
			rules:       map[string]string{"a": "required"},
			expectErr:   errors.New("data can't be empty"),
		},
		{
			description: "error when rule is empty map",
			data:        map[string]any{"a": "b"},
			rules:       map[string]string{},
			expectErr:   errors.New("rules can't be empty"),
		},
		{
			description: "error when PrepareForValidation returns error",
			data:        map[string]any{"a": "b"},
			rules:       map[string]string{"a": "required"},
			options: []httpvalidate.Option{
				PrepareForValidation(func(data httpvalidate.Data) error {
					return errors.New("error")
				}),
			},
			expectErr: errors.New("error"),
		},
		{
			description: "success when data is map[string]any and with PrepareForValidation",
			data:        map[string]any{"a": "b"},
			rules:       map[string]string{"a": "required"},
			options: []httpvalidate.Option{
				PrepareForValidation(func(data httpvalidate.Data) error {
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
			description: "contain errors when data is map[string]any and with Messages, Attributes, PrepareForValidation",
			data:        map[string]any{"a": "aa"},
			rules:       map[string]string{"a": "required", "b": "required"},
			options: []httpvalidate.Option{
				Messages(map[string]string{
					"b.required": ":attribute can't be empty",
				}),
				Attributes(map[string]string{
					"b": "B",
				}),
				PrepareForValidation(func(data httpvalidate.Data) error {
					if _, exist := data.Get("a"); exist {
						return data.Set("a", "c")
					}

					return nil
				}),
			},
			expectValidator:    true,
			expectData:         Data{A: "c"},
			expectErrors:       true,
			expectErrorMessage: "B can't be empty",
		},
		{
			description: "success when data is struct and with PrepareForValidation",
			data:        &Data{A: "b"},
			rules:       map[string]string{"A": "required"},
			options: []httpvalidate.Option{
				PrepareForValidation(func(data httpvalidate.Data) error {
					if _, exist := data.Get("A"); exist {
						return data.Set("A", "c")
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
			rules:       map[string]string{"A": "required", "B": "required"},
			options: []httpvalidate.Option{
				Messages(map[string]string{
					"B.required": ":attribute can't be empty",
				}),
				Attributes(map[string]string{
					"B": "b",
				}),
				PrepareForValidation(func(data httpvalidate.Data) error {
					if _, exist := data.Get("a"); exist {
						return data.Set("a", "c")
					}

					return nil
				}),
			},
			expectValidator:    true,
			expectData:         Data{A: "c"},
			expectErrors:       true,
			expectErrorMessage: "b can't be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			validation := NewValidation()
			validator, err := validation.Make(test.data, test.rules, test.options...)
			assert.Equal(t, test.expectValidator, validator != nil, test.description)
			assert.Equal(t, test.expectErr, err, test.description)

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

type Case struct {
	description string
	setup       func(Case)
}

func TestRule_Required(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "goravel",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name": map[string]string{
						"first": "Goravel",
					},
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name": "",
				}, map[string]string{
					"name": "required",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required": "name is required to not be empty",
				}, validator.Errors().Get("name"))
			},
		},
		{
			description: "error when key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "Goravel",
				}, map[string]string{
					"name":  "required",
					"name1": "required",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required": "name1 is required to not be empty",
				}, validator.Errors().Get("name1"))
			},
		},
		{
			description: "error when nested",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": map[string]string{
						"first": "",
					},
				}, map[string]string{
					"name.first": "required",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required": "name.first is required to not be empty",
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
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name": "goravel2",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name1": "",
				}, map[string]string{
					"name":  "required",
					"name1": "required_if:name,goravel,goravel1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_if": "name1 is required when name is [goravel,goravel1]",
				}, validator.Errors().Get("name1"))
			},
		},
		{
			description: "error when required_if is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "goravel",
				}, map[string]string{
					"name":  "required",
					"name1": "required_if:name,goravel,goravel1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_if": "name1 is required when name is [goravel,goravel1]",
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
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name": "goravel",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name1": "",
				}, map[string]string{
					"name":  "required",
					"name1": "required_unless:name,hello,hello1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_unless": "name1 field is required unless name is in [hello,hello1]",
				}, validator.Errors().Get("name1"))
			},
		},
		{
			description: "error when required_unless is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "goravel",
				}, map[string]string{
					"name":  "required",
					"name1": "required_unless:name,hello,hello1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_unless": "name1 field is required unless name is in [hello,hello1]",
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
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name2": "goravel2",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name": "",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
					"name2": "",
				}, map[string]string{
					"name":  "required",
					"name1": "required",
					"name2": "required_with:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_with": "name2 field is required when [name,name1] is present",
				}, validator.Errors().Get("name2"))
			},
		},
		{
			description: "error when required_with is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
				}, map[string]string{
					"name":  "required",
					"name1": "required",
					"name2": "required_with:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_with": "name2 field is required when [name,name1] is present",
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
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
					"name2": "goravel2",
				}, map[string]string{
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
			description: "success when required_with_all is true",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name1": "",
					"name2": "goravel2",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name": "",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
					"name2": "",
				}, map[string]string{
					"name":  "required",
					"name1": "required",
					"name2": "required_with_all:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_with_all": "name2 field is required when [name,name1] is present",
				}, validator.Errors().Get("name2"))
			},
		},
		{
			description: "error when required_with is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name1": "goravel1",
				}, map[string]string{
					"name":  "required",
					"name1": "required",
					"name2": "required_with_all:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_with_all": "name2 field is required when [name,name1] is present",
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
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name2": "goravel2",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name":  "",
					"name1": "",
					"name2": "",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name":  "goravel",
					"name2": "",
				}, map[string]string{
					"name":  "required",
					"name2": "required_without:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_without": "name2 field is required when [name,name1] is not present",
				}, validator.Errors().Get("name2"))
			},
		},
		{
			description: "error when required_without is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "goravel",
				}, map[string]string{
					"name":  "required",
					"name2": "required_without:name,name1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_without": "name2 field is required when [name,name1] is not present",
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
				validator, err := validation.Make(map[string]any{
					"name": "goravel",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name":  "",
					"name1": "",
					"name2": "",
				}, map[string]string{
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
				validator, err := validation.Make(map[string]any{
					"name": "",
				}, map[string]string{
					"name": "required_without_all:name1,name2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_without_all": "name field is required when none of [name1,name2] are present",
				}, validator.Errors().Get("name"))
			},
		},
		{
			description: "error when required_without_all is true and key isn't exist",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name3": "goravel3",
				}, map[string]string{
					"name": "required_without_all:name1,name2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"required_without_all": "name field is required when none of [name1,name2] are present",
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

func TestRule_Int(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 1,
				}, map[string]string{
					"name": "required|int",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "success with range",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 3,
				}, map[string]string{
					"name": "required|int:2,4",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when type error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "1",
				}, map[string]string{
					"name": "required|int",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"int": "name value must be an integer",
				}, validator.Errors().Get("name"))
			},
		},
		{
			description: "error when value doesn't in the right range",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 1,
				}, map[string]string{
					"name": "required|int:2,4",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"int": "name value must be an integer and in the range 2 - 4",
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

func TestRule_Uint(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 1,
				}, map[string]string{
					"name": "required|uint",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when type error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "s",
				}, map[string]string{
					"name": "required|uint",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"uint": "name value must be an unsigned integer(>= 0)",
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

func TestRule_Bool(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name1":  "on",
					"name2":  "off",
					"name3":  "yes",
					"name4":  "no",
					"name5":  true,
					"name6":  false,
					"name7":  "true",
					"name8":  "false",
					"name9":  "1",
					"name10": "0",
				}, map[string]string{
					"name1":  "bool",
					"name2":  "bool",
					"name3":  "bool",
					"name4":  "bool",
					"name5":  "bool",
					"name6":  "bool",
					"name7":  "bool",
					"name8":  "bool",
					"name9":  "bool",
					"name10": "bool",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when type error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name1": 1,
					"name2": 0,
					"name3": "a",
				}, map[string]string{
					"name1": "bool",
					"name2": "bool",
					"name3": "bool",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"bool": "name1 value must be a bool"}, validator.Errors().Get("name1"))
				assert.Nil(t, validator.Errors().Get("name2"))
				assert.Equal(t, map[string]string{"bool": "name3 value must be a bool"}, validator.Errors().Get("name3"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_String(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "1",
				}, map[string]string{
					"name": "required|string",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "success with range",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "abc",
				}, map[string]string{
					"name": "required|string:2,4",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when type error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 1,
				}, map[string]string{
					"name": "required|string",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"string": "name value must be a string",
				}, validator.Errors().Get("name"))
			},
		},
		{
			description: "error when value doesn't in the right range",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "a",
				}, map[string]string{
					"name": "required|string:2,4",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"string": "name value must be a string",
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

func TestRule_Float(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 1.1,
				}, map[string]string{
					"name": "required|float",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when type error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "a",
				}, map[string]string{
					"name": "required|float",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{
					"float": "name value must be a float",
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

func TestRule_Slice(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name1": []int{1, 2},
					"name2": []uint{1, 2},
					"name3": []string{"a", "b"},
				}, map[string]string{
					"name1": "required|slice",
					"name2": "required|slice",
					"name3": "required|slice",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error when type error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name1": 1,
					"name2": "a",
					"name3": true,
				}, map[string]string{
					"name1": "required|slice",
					"name2": "required|slice",
					"name3": "required|slice",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"slice": "name1 value must be a slice"}, validator.Errors().Get("name1"))
				assert.Equal(t, map[string]string{"slice": "name2 value must be a slice"}, validator.Errors().Get("name2"))
				assert.Equal(t, map[string]string{"slice": "name3 value must be a slice"}, validator.Errors().Get("name3"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_In(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name1": 1,
					"name2": "a",
				}, map[string]string{
					"name1": "required|in:1,2",
					"name2": "required|in:a,b",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name1": 3,
					"name2": "c",
				}, map[string]string{
					"name1": "required|in:1,2",
					"name2": "required|in:a,b",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"in": "name1 value must be in the enum [1 2]"}, validator.Errors().Get("name1"))
				assert.Equal(t, map[string]string{"in": "name2 value must be in the enum [a b]"}, validator.Errors().Get("name2"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_NotIn(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name1": 3,
					"name2": "c",
				}, map[string]string{
					"name1": "required|not_in:1,2",
					"name2": "required|not_in:a,b",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name1": 1,
					"name2": "a",
				}, map[string]string{
					"name1": "required|not_in:1,2",
					"name2": "required|not_in:a,b",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"not_in": "name1 value must not be in the given enum list [%!d(string=1) %!d(string=2)]"}, validator.Errors().Get("name1"))
				assert.Equal(t, map[string]string{"not_in": "name2 value must not be in the given enum list [%!d(string=a) %!d(string=b)]"}, validator.Errors().Get("name2"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_StartsWith(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "abc",
				}, map[string]string{
					"name": "required|starts_with:ab",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "a",
				}, map[string]string{
					"name": "required|starts_with:ab",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"starts_with": "name value does not start with ab"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_EndsWith(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "cab",
				}, map[string]string{
					"name": "required|ends_with:ab",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "a",
				}, map[string]string{
					"name": "required|ends_with:ab",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"ends_with": "name value does not end with ab"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Between(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 2,
				}, map[string]string{
					"name": "required|between:1,3",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 1,
				}, map[string]string{
					"name": "required|between:2,4",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"between": "name field did not pass validation"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Max(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 2,
				}, map[string]string{
					"name": "required|max:3",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 4,
				}, map[string]string{
					"name": "required|max:3",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"max": "name max value is 3"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Min(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 3,
				}, map[string]string{
					"name": "required|min:3",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 2,
				}, map[string]string{
					"name": "required|min:3",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"min": "name min value is 3"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Eq(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "a",
				}, map[string]string{
					"name": "required|eq:a",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "b",
				}, map[string]string{
					"name": "required|eq:a",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"eq": "name field did not pass validation"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Ne(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "b",
				}, map[string]string{
					"name": "required|ne:a",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "a",
				}, map[string]string{
					"name": "required|ne:a",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"ne": "name field did not pass validation"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Lt(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 1,
				}, map[string]string{
					"name": "required|lt:2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 2,
				}, map[string]string{
					"name": "required|lt:1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"lt": "name value should be less than 1"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Gt(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 2,
				}, map[string]string{
					"name": "required|gt:1",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 1,
				}, map[string]string{
					"name": "required|gt:2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"gt": "name value should be greater than 2"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Len(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "abc",
					"name1": [3]string{"a", "b", "c"},
					"name2": []string{"a", "b", "c"},
					"name3": map[string]string{
						"a": "a1",
						"b": "b1",
						"c": "c1",
					},
				}, map[string]string{
					"name":  "required|len:3",
					"name1": "required|len:3",
					"name2": "required|len:3",
					"name3": "required|len:3",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "abc",
					"name1": [3]string{"a", "b", "c"},
					"name2": []string{"a", "b", "c"},
					"name3": map[string]string{
						"a": "a1",
						"b": "b1",
						"c": "c1",
					},
				}, map[string]string{
					"name":  "required|len:2",
					"name1": "required|len:2",
					"name2": "required|len:2",
					"name3": "required|len:2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"len": "name field did not pass validation"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"len": "name1 field did not pass validation"}, validator.Errors().Get("name1"))
				assert.Equal(t, map[string]string{"len": "name2 field did not pass validation"}, validator.Errors().Get("name2"))
				assert.Equal(t, map[string]string{"len": "name3 field did not pass validation"}, validator.Errors().Get("name3"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_MinLen(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "abc",
					"name1": [3]string{"a", "b", "c"},
					"name2": []string{"a", "b", "c"},
					"name3": map[string]string{
						"a": "a1",
						"b": "b1",
						"c": "c1",
					},
				}, map[string]string{
					"name":  "required|min_len:2",
					"name1": "required|min_len:2",
					"name2": "required|min_len:2",
					"name3": "required|min_len:2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "abc",
					"name1": [3]string{"a", "b", "c"},
					"name2": []string{"a", "b", "c"},
					"name3": map[string]string{
						"a": "a1",
						"b": "b1",
						"c": "c1",
					},
				}, map[string]string{
					"name":  "required|min_len:4",
					"name1": "required|min_len:4",
					"name2": "required|min_len:4",
					"name3": "required|min_len:4",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"min_len": "name min length is 4"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"min_len": "name1 min length is 4"}, validator.Errors().Get("name1"))
				assert.Equal(t, map[string]string{"min_len": "name2 min length is 4"}, validator.Errors().Get("name2"))
				assert.Equal(t, map[string]string{"min_len": "name3 min length is 4"}, validator.Errors().Get("name3"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_MaxLen(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "abc",
					"name1": [3]string{"a", "b", "c"},
					"name2": []string{"a", "b", "c"},
					"name3": map[string]string{
						"a": "a1",
						"b": "b1",
						"c": "c1",
					},
				}, map[string]string{
					"name":  "required|max_len:4",
					"name1": "required|max_len:4",
					"name2": "required|max_len:4",
					"name3": "required|max_len:4",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "abc",
					"name1": [3]string{"a", "b", "c"},
					"name2": []string{"a", "b", "c"},
					"name3": map[string]string{
						"a": "a1",
						"b": "b1",
						"c": "c1",
					},
				}, map[string]string{
					"name":  "required|max_len:2",
					"name1": "required|max_len:2",
					"name2": "required|max_len:2",
					"name3": "required|max_len:2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"max_len": "name max length is 2"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"max_len": "name1 max length is 2"}, validator.Errors().Get("name1"))
				assert.Equal(t, map[string]string{"max_len": "name2 max length is 2"}, validator.Errors().Get("name2"))
				assert.Equal(t, map[string]string{"max_len": "name3 max length is 2"}, validator.Errors().Get("name3"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Email(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "hello@goravel.com",
				}, map[string]string{
					"name": "required|email",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "abc",
				}, map[string]string{
					"name": "required|email",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"email": "name value is an invalid email address"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Array(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  [2]string{"a", "b"},
					"name1": []string{"a", "b"},
				}, map[string]string{
					"name":  "required|array",
					"name1": "required|array",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "a",
					"name1": 1,
					"name2": true,
				}, map[string]string{
					"name":  "required|array",
					"name1": "required|array",
					"name2": "required|array",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"array": "name value must be an array"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"array": "name1 value must be an array"}, validator.Errors().Get("name1"))
				assert.Equal(t, map[string]string{"array": "name2 value must be an array"}, validator.Errors().Get("name2"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Map(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": map[string]string{"a": "a1"},
				}, map[string]string{
					"name": "required|map",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "a",
					"name1": 1,
					"name2": true,
					"name3": []string{"a"},
				}, map[string]string{
					"name":  "required|map",
					"name1": "required|map",
					"name2": "required|map",
					"name3": "required|map",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"map": "name value must be a map"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"map": "name1 value must be a map"}, validator.Errors().Get("name1"))
				assert.Equal(t, map[string]string{"map": "name2 value must be a map"}, validator.Errors().Get("name2"))
				assert.Equal(t, map[string]string{"map": "name3 value must be a map"}, validator.Errors().Get("name3"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_EqField(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "a",
					"name1": "a",
				}, map[string]string{
					"name":  "required",
					"name1": "required|eq_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "a",
					"name1": "b",
				}, map[string]string{
					"name":  "required",
					"name1": "required|eq_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"eq_field": "name1 value must be equal the field name"}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_NeField(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "a",
					"name1": "b",
				}, map[string]string{
					"name":  "required",
					"name1": "required|ne_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "a",
					"name1": "a",
				}, map[string]string{
					"name":  "required",
					"name1": "required|ne_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"ne_field": "name1 value cannot be equal to the field name"}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_GtField(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  1,
					"name1": 2,
				}, map[string]string{
					"name":  "required",
					"name1": "required|gt_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  2,
					"name1": 1,
				}, map[string]string{
					"name":  "required",
					"name1": "required|gt_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"gt_field": "name1 value must be greater than the field name"}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_GteField(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  1,
					"name1": 2,
					"name2": 1,
				}, map[string]string{
					"name":  "required",
					"name1": "required|gte_field:name",
					"name2": "required|gte_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  2,
					"name1": 1,
				}, map[string]string{
					"name":  "required",
					"name1": "required|gte_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"gte_field": "name1 value should be greater or equal to the field name"}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_LtField(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  2,
					"name1": 1,
				}, map[string]string{
					"name":  "required",
					"name1": "required|lt_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  1,
					"name1": 2,
				}, map[string]string{
					"name":  "required",
					"name1": "required|lt_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"lt_field": "name1 value should be less than the field name"}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_LteField(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  2,
					"name1": 2,
					"name2": 1,
				}, map[string]string{
					"name":  "required",
					"name1": "required|lte_field:name",
					"name2": "required|lte_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  1,
					"name1": 2,
				}, map[string]string{
					"name":  "required",
					"name1": "required|lte_field:name",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"lte_field": "name1 value should be less than or equal to the field name"}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Date(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "2022-12-25",
					"name1": "2022/12/25",
					"name2": "",
				}, map[string]string{
					"name":  "required|date",
					"name1": "required|date",
					"name2": "date",
					"name3": "date",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "2022.12.25",
					"name1": "a",
				}, map[string]string{
					"name":  "required|date",
					"name1": "required|date",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"date": "name value should be a date string"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"date": "name1 value should be a date string"}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_GtDate(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "2022-12-25",
				}, map[string]string{
					"name": "required|gt_date:2022-12-24",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "2022-12-25",
				}, map[string]string{
					"name": "required|gt_date:2022-12-26",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"gt_date": "name field did not pass validation"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_LtDate(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "2022-12-25",
				}, map[string]string{
					"name": "required|lt_date:2022-12-26",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "2022-12-25",
				}, map[string]string{
					"name": "required|lt_date:2022-12-24",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"lt_date": "name field did not pass validation"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_GteDate(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "2022-12-25",
					"name1": "2022-12-25",
				}, map[string]string{
					"name":  "required|gte_date:2022-12-25",
					"name1": "required|gte_date:2022-12-24",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "2022-12-25",
				}, map[string]string{
					"name": "required|gte_date:2022-12-26",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"gte_date": "name field did not pass validation"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_lteDate(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "2022-12-25",
					"name1": "2022-12-25",
				}, map[string]string{
					"name":  "required|lte_date:2022-12-25",
					"name1": "required|lte_date:2022-12-26",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "2022-12-25",
				}, map[string]string{
					"name": "required|lte_date:2022-12-24",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"lte_date": "name field did not pass validation"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Alpha(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "abcABC",
				}, map[string]string{
					"name": "required|alpha",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "abcABC123",
					"name1": "abc.",
				}, map[string]string{
					"name":  "required|alpha",
					"name1": "required|alpha",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"alpha": "name value contains only alpha char"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"alpha": "name1 value contains only alpha char"}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_AlphaNum(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "abcABC123",
				}, map[string]string{
					"name": "required|alpha_num",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "abcABC123.",
				}, map[string]string{
					"name": "required|alpha_num",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"alpha_num": "name field did not pass validation"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_AlphaDash(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "abcABC123-_",
				}, map[string]string{
					"name": "required|alpha_dash",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "abcABC123-_.",
				}, map[string]string{
					"name": "required|alpha_dash",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"alpha_dash": "name field did not pass validation"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Json(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "{\"a\":1}",
				}, map[string]string{
					"name": "required|json",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "a",
				}, map[string]string{
					"name": "required|json",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"json": "name value should be a json string"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Number(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": 1,
				}, map[string]string{
					"name": "required|number",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "a",
				}, map[string]string{
					"name": "required|number",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"number": "name field did not pass validation"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_FullUrl(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "https://www.goravel.dev",
					"name1": "http://www.goravel.dev",
				}, map[string]string{
					"name":  "required|full_url",
					"name1": "required|full_url",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "a",
				}, map[string]string{
					"name": "required|full_url",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"full_url": "name must be a valid full URL address"}, validator.Errors().Get("name"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Ip(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "192.168.1.1",
					"name1": "FE80:CD00:0000:0CDE:1257:0000:211E:729C",
				}, map[string]string{
					"name":  "required|ip",
					"name1": "required|ip",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "a",
					"name1": "192.168.1.300",
				}, map[string]string{
					"name":  "required|ip",
					"name1": "required|ip",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"ip": "name value should be an IP (v4 or v6) string"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"ip": "name1 value should be an IP (v4 or v6) string"}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Ipv4(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "192.168.1.1",
				}, map[string]string{
					"name": "required|ipv4",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "a",
					"name1": "FE80:CD00:0000:0CDE:1257:0000:211E:729C",
					"name2": "192.168.1.300",
				}, map[string]string{
					"name":  "required|ipv4",
					"name1": "required|ipv4",
					"name2": "required|ipv4",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"ipv4": "name value should be an IPv4 string"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"ipv4": "name1 value should be an IPv4 string"}, validator.Errors().Get("name1"))
				assert.Equal(t, map[string]string{"ipv4": "name2 value should be an IPv4 string"}, validator.Errors().Get("name2"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestRule_Ipv6(t *testing.T) {
	validation := NewValidation()
	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name": "FE80:CD00:0000:0CDE:1257:0000:211E:729C",
				}, map[string]string{
					"name": "required|ipv6",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":  "a",
					"name1": "192.168.1.300",
				}, map[string]string{
					"name":  "required|ipv6",
					"name1": "required|ipv6",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"ipv6": "name value should be an IPv6 string"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"ipv6": "name1 value should be an IPv6 string"}, validator.Errors().Get("name1"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

func TestAddRule(t *testing.T) {
	validation := NewValidation()
	err := validation.AddRules([]httpvalidate.Rule{&Uppercase{}})
	assert.Nil(t, err)

	err = validation.AddRules([]httpvalidate.Rule{&Duplicate{}})
	assert.EqualError(t, err, "duplicate rule name: required")
}

func TestCustomRule(t *testing.T) {
	validation := NewValidation()
	err := validation.AddRules([]httpvalidate.Rule{&Uppercase{}, &Lowercase{}})
	assert.Nil(t, err)

	tests := []Case{
		{
			description: "success",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":    "ABC",
					"address": "de",
				}, map[string]string{
					"name":    "required|uppercase:3",
					"address": "required|lowercase:2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.False(t, validator.Fails(), c.description)
			},
		},
		{
			description: "error",
			setup: func(c Case) {
				validator, err := validation.Make(map[string]any{
					"name":    "abc",
					"address": "DE",
				}, map[string]string{
					"name":    "required|uppercase:3",
					"address": "required|lowercase:2",
				})
				assert.Nil(t, err, c.description)
				assert.NotNil(t, validator, c.description)
				assert.Equal(t, map[string]string{"uppercase": "name must be upper"}, validator.Errors().Get("name"))
				assert.Equal(t, map[string]string{"lowercase": "address must be lower"}, validator.Errors().Get("address"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setup(test)
		})
	}
}

type Uppercase struct {
}

//Signature The name of the rule.
func (receiver *Uppercase) Signature() string {
	return "uppercase"
}

//Passes Determine if the validation rule passes.
func (receiver *Uppercase) Passes(data httpvalidate.Data, val any, options ...any) bool {
	name, exist := data.Get("name")

	return strings.ToUpper(val.(string)) == val.(string) && len(val.(string)) == cast.ToInt(options[0]) && name == val && exist
}

//Message Get the validation error message.
func (receiver *Uppercase) Message() string {
	return ":attribute must be upper"
}

type Lowercase struct {
}

//Signature The name of the rule.
func (receiver *Lowercase) Signature() string {
	return "lowercase"
}

//Passes Determine if the validation rule passes.
func (receiver *Lowercase) Passes(data httpvalidate.Data, val any, options ...any) bool {
	address, exist := data.Get("address")

	return strings.ToLower(val.(string)) == val.(string) && len(val.(string)) == cast.ToInt(options[0]) && address == val && exist
}

//Message Get the validation error message.
func (receiver *Lowercase) Message() string {
	return ":attribute must be lower"
}

type Duplicate struct {
}

//Signature The name of the rule.
func (receiver *Duplicate) Signature() string {
	return "required"
}

//Passes Determine if the validation rule passes.
func (receiver *Duplicate) Passes(data httpvalidate.Data, val any, options ...any) bool {
	return true
}

//Message Get the validation error message.
func (receiver *Duplicate) Message() string {
	return ""
}
