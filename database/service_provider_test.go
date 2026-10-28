package database

import (
	"testing"

	"github.com/stretchr/testify/mock"

	contractsconsole "github.com/goravel/framework/contracts/console"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
	mocksseeder "github.com/goravel/framework/mocks/database/seeder"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockslog "github.com/goravel/framework/mocks/log"
)

func TestServiceProviderBoot(t *testing.T) {
	t.Run("registers commands through app.Commands", func(t *testing.T) {
		provider := &ServiceProvider{}
		app := mocksfoundation.NewApplication(t)
		config := mocksconfig.NewConfig(t)

		app.EXPECT().MakeArtisan().Return(mocksconsole.NewArtisan(t)).Once()
		app.EXPECT().MakeConfig().Return(config).Once()
		app.EXPECT().MakeLog().Return(mockslog.NewLog(t)).Once()
		app.EXPECT().MakeSchema().Return(mocksschema.NewSchema(t)).Once()
		app.EXPECT().MakeSeeder().Return(mocksseeder.NewFacade(t)).Once()
		config.EXPECT().GetString("database.migrations.table").Return("migrations").Once()
		app.EXPECT().Commands(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
			return len(commands) == 15 && commands[0].Signature() == "make:migration"
		})).Once()

		provider.Boot(app)
	})

	t.Run("skips registration when a dependency is missing", func(t *testing.T) {
		provider := &ServiceProvider{}
		app := mocksfoundation.NewApplication(t)

		app.EXPECT().MakeArtisan().Return(nil).Once()
		app.EXPECT().MakeConfig().Return(nil).Once()
		app.EXPECT().MakeLog().Return(nil).Once()
		app.EXPECT().MakeSchema().Return(nil).Once()
		app.EXPECT().MakeSeeder().Return(nil).Once()

		provider.Boot(app)
	})
}
