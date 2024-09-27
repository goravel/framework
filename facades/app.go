package facades

import (
	foundationcontract "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func App() foundationcontract.Application {
	if foundation.App == nil {
		panic(ErrApplicationNotSet)
	} else {
		return foundation.App
	}
}
