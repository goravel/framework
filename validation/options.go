package validation

import (
	"context"

	contractsvalidation "github.com/goravel/framework/contracts/validation"
)

func Filters(filters map[string]any) contractsvalidation.Option {
	return func(opts *contractsvalidation.Options) {
		opts.Filters = filters
	}
}

func Messages(messages map[string]string) contractsvalidation.Option {
	return func(opts *contractsvalidation.Options) {
		opts.Messages = messages
	}
}

func Attributes(attributes map[string]string) contractsvalidation.Option {
	return func(opts *contractsvalidation.Options) {
		opts.Attributes = attributes
	}
}

func PrepareForValidation(prepare func(ctx context.Context, data contractsvalidation.Data) error) contractsvalidation.Option {
	return func(opts *contractsvalidation.Options) {
		opts.PrepareForValidation = prepare
	}
}

func CustomFilters(filters []contractsvalidation.Filter) contractsvalidation.Option {
	return func(opts *contractsvalidation.Options) {
		opts.CustomFilters = filters
	}
}

func MaxMultipartMemory(maxMemory int64) contractsvalidation.Option {
	return func(opts *contractsvalidation.Options) {
		opts.MaxMultipartMemory = maxMemory
	}
}

func applyOptions(options []contractsvalidation.Option) *contractsvalidation.Options {
	opts := &contractsvalidation.Options{}
	for _, o := range options {
		o(opts)
	}
	return opts
}
