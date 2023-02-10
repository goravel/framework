package crypt

import (
	"github.com/goravel/framework/contracts/crypt"
)

func NewApplication() crypt.Crypt {
	return NewAES()
}
