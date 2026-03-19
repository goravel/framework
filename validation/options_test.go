package validation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	contractsvalidation "github.com/goravel/framework/contracts/validation"
)

func TestFiltersOption(t *testing.T) {
	filters := map[string]any{"name": "trim|lower", "email": "lower"}

	opts := &contractsvalidation.Options{}
	Filters(filters)(opts)

	assert.Equal(t, filters, opts.Filters)
}

func TestCustomFiltersOption(t *testing.T) {
	f := &mockFilter{signature: "test"}
	customFilters := []contractsvalidation.Filter{f}

	opts := &contractsvalidation.Options{}
	CustomFilters(customFilters)(opts)

	assert.Len(t, opts.CustomFilters, 1)
	assert.Equal(t, "test", opts.CustomFilters[0].Signature())
}

func TestMessagesOption(t *testing.T) {
	messages := map[string]string{"required": "Field is required", "email": "Invalid email"}

	opts := &contractsvalidation.Options{}
	Messages(messages)(opts)

	assert.Equal(t, messages, opts.Messages)
}

func TestAttributesOption(t *testing.T) {
	attributes := map[string]string{"name": "Full Name", "email": "Email Address"}

	opts := &contractsvalidation.Options{}
	Attributes(attributes)(opts)

	assert.Equal(t, attributes, opts.Attributes)
}

func TestPrepareForValidationOption(t *testing.T) {
	prepareFunc := func(ctx context.Context, data contractsvalidation.Data) error {
		return nil
	}

	opts := &contractsvalidation.Options{}
	PrepareForValidation(prepareFunc)(opts)

	assert.NotNil(t, opts.PrepareForValidation)
}

func TestApplyOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []contractsvalidation.Option
		check   func(t *testing.T, opts *contractsvalidation.Options)
	}{
		{
			name: "with multiple options",
			options: []contractsvalidation.Option{
				Filters(map[string]any{"name": "trim"}),
				Messages(map[string]string{"required": "Field is required"}),
				Attributes(map[string]string{"name": "Full Name"}),
			},
			check: func(t *testing.T, opts *contractsvalidation.Options) {
				assert.Equal(t, map[string]any{"name": "trim"}, opts.Filters)
				assert.Equal(t, map[string]string{"required": "Field is required"}, opts.Messages)
				assert.Equal(t, map[string]string{"name": "Full Name"}, opts.Attributes)
			},
		},
		{
			name:    "with no options",
			options: []contractsvalidation.Option{},
			check: func(t *testing.T, opts *contractsvalidation.Options) {
				assert.Nil(t, opts.Filters)
				assert.Nil(t, opts.Messages)
				assert.Nil(t, opts.Attributes)
				assert.Nil(t, opts.PrepareForValidation)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyOptions(tt.options)
			tt.check(t, result)
		})
	}
}
