package packages

import (
	"github.com/goravel/framework/contracts/packages/modify"
)

type Setup interface {
	Execute()
	PackageName() string
	Install(modifiers ...modify.Apply) Setup
	Uninstall(modifiers ...modify.Apply) Setup
}
