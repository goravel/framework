package providers

import (
	"github.com/goravel/framework/console/support"
	"github.com/goravel/framework/foundation/console"
	"github.com/goravel/framework/support/facades"
)

type ArtisanServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (artisan *ArtisanServiceProvider) Boot() {

}

//Register Register any application services.
func (artisan *ArtisanServiceProvider) Register() {
	artisan.registerCommands()
}

func (artisan *ArtisanServiceProvider) registerCommands() {
	facades.Artisan.Register([]support.Command{
		console.KeyGenerateCommand{},
	})
}
