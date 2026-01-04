package main

import "strings"

type Stubs struct{}

func (s Stubs) LangFacade(pkg string) string {
	content := `package DummyPackage

import (
	"context"

	"github.com/goravel/framework/contracts/translation"
)

func Lang(ctx context.Context) translation.Translator {
	return App().MakeLang(ctx)
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
