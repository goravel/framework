package main

import "strings"

type Stubs struct{}

func (s Stubs) ProcessFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/process"
)

func Process() process.Process {
	return App().MakeProcess()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
