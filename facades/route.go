package facades

import (
	"github.com/goravel/framework/contracts/route"
)

func Route() route.Engine {
	return App().MakeRoute()
}
