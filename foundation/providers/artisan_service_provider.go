package providers

import (
	console2 "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/foundation/console"
	"github.com/goravel/framework/support/facades"
)

type ArtisanServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (artisan *ArtisanServiceProvider) Boot() {
	artisan.registerCommands()
}

//Register any application services.
func (artisan *ArtisanServiceProvider) Register() {

}

func (artisan *ArtisanServiceProvider) registerCommands() {
	facades.Artisan.Register([]console2.Command{
		&console.KeyGenerateCommand{},
		&console.ConsoleMakeCommand{},
		&console.ListCommand{},
	})
}
