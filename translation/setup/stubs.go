package main

type Stubs struct{}

func (s Stubs) LangFacade() string {
	return `package facades

import (
	"context"

	"github.com/goravel/framework/contracts/translation"
)

func Lang(ctx context.Context) translation.Translator {
	return App().MakeLang(ctx)
}
`
}
