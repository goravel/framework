package main

import "strings"

type Stubs struct{}

func (s Stubs) CryptFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/crypt"
)

func Crypt() crypt.Crypt {
	return App().MakeCrypt()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
