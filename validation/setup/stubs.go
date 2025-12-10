package main

type Stubs struct{}

func (s Stubs) ValidationFacade(pkg string) string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/validation"
)

func Validation() validation.Validation {
	return App().MakeValidation()
}
`
}
