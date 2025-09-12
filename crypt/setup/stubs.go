package main

type Stubs struct{}

func (s Stubs) CryptFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/crypt"
)

func Crypt() crypt.Crypt {
	return App().MakeCrypt()
}
`
}
