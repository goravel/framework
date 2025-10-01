package main

type Stubs struct{}

func (s Stubs) RouteFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/route"
)

func Route() route.Route {
	return App().MakeRoute()
}
`
}

func (s Stubs) Routes() string {
	return `package routes

func Web() {

}
`
}
