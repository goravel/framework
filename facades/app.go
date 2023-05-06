package facades

import (
	foundationcontract "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func App() foundationcontract.Application {
	return foundation.App
}
