package facades

import (
	"github.com/goravel/framework/contracts/console"
)

var Artisan console.Artisan

func NewArtisan() console.Artisan {
	return App().MakeArtisan()
}
