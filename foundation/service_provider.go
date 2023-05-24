package foundation

import (
	"reflect"
)

var Publishes = make(map[string]map[string]string)

type ServiceProvider struct {
}

func (receiver ServiceProvider) Publishes(paths map[string]string, groups ...string) {
	a := reflect.TypeOf(receiver)
	Publishes[a.PkgPath()] = paths
}
