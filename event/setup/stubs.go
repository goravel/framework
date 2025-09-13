package main

type Stubs struct{}

func (s Stubs) EventFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/event"
)

func Event() event.Instance {
	return App().MakeEvent()
}
`
}
