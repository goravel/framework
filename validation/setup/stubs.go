package main

type Stubs struct{}

func (s Stubs) ValidationFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/validation"
)

func Validation() validation.Validation {
	return App().MakeValidation()
}
`
}
