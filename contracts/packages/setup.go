package packages

import (
	"github.com/goravel/framework/contracts/packages/modify"
)

type Setup interface {
	Install(modifiers ...modify.GoFile) Setup
	Uninstall(modifiers ...modify.GoFile) Setup
	Execute()
}
