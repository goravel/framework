package crypt

import (
	"github.com/goravel/framework/contracts/crypt"
)

type Application struct {
}

func NewApplication() crypt.Crypt {
	return NewAES()
}
