package console

type Stubs struct {
}

func (r Stubs) Rule() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/validation"
)

type DummyRule struct {
}

// Signature The name of the rule.
func (receiver *DummyRule) Signature() string {
	return "DummyName"
}

// Passes Determine if the validation rule passes.
func (receiver *DummyRule) Passes(data validation.Data, val any, options ...any) bool {
	return true
}

// Message Get the validation error message.
func (receiver *DummyRule) Message() string {
	return ""
}
`
}

func (r Stubs) Filter() string {
	return `package DummyPackage

type DummyFilter struct {
}

// Signature The signature of the filter.
func (receiver *DummyFilter) Signature() string {
	return "DummyName"
}

// Handle The filter function to apply.
func (receiver *DummyFilter) Handle() any {
    // below is an example of a filter that does nothing
	return func() {}
}
`
}
