package main

import "strings"

type Stubs struct{}

func (s Stubs) EventFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/event"
)

func Event() event.Instance {
	return App().MakeEvent()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
