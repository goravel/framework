package main

import "strings"

type Stubs struct{}

func (s Stubs) ArtisanFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/console"
)

func Artisan() console.Artisan {
	return App().MakeArtisan()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
