package options

import "github.com/goravel/framework/contracts/packages/modify"

func Facade(facade string) modify.Option {
	return func(options map[string]any) {
		options["facade"] = facade
	}
}

func Force(force bool) modify.Option {
	return func(options map[string]any) {
		options["force"] = force
	}
}
