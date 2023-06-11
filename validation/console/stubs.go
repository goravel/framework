package console

type Stubs struct {
}

func (r Stubs) Request() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/validation"
)

type DummyRule struct {
}

//Signature The name of the rule.
func (receiver *DummyRule) Signature() string {
	return "DummyName"
}

//Passes Determine if the validation rule passes.
func (receiver *DummyRule) Passes(data validation.Data, val any, options ...any) bool {
	return true
}

//Message Get the validation error message.
func (receiver *DummyRule) Message() string {
	return ""
}
`
}
