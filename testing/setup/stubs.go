package main

type Stubs struct{}

func (s Stubs) TestingFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/testing"
)

func Testing() testing.Testing {
	return App().MakeTesting()
}
`
}
