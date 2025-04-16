package packages

import (
	"github.com/goravel/framework/contracts/packages/modify"
)

type Setup interface {
	Install(modifiers ...modify.File) Setup
	Uninstall(modifiers ...modify.File) Setup
	Execute()
}
