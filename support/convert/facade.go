package convert

import (
	"github.com/goravel/framework/contracts/facades"
)

func BindingToFacade(binding string) string {
	for facade, b := range facades.FacadeToBinding {
		if b == binding {
			return facade
		}
	}

	return ""
}

func FacadeToBinding(facade string) string {
	return facades.FacadeToBinding[facade]
}
