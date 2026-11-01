package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsconfig "github.com/goravel/framework/contracts/config"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	contractsseeder "github.com/goravel/framework/contracts/database/seeder"
	contractslog "github.com/goravel/framework/contracts/log"
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

		expectedSignatures := []string{
			"make:migration",
			"migrate",
			"migrate:rollback",
			"migrate:reset",
			"migrate:refresh",
			"migrate:fresh",
			"migrate:status",
			"make:model",
			"make:observer",
			"db:seed",
			"make:seeder",
			"make:factory",
			"db:table",
			"db:show",
			"db:wipe",
		}
		app.EXPECT().Commands(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
			signatures := make([]string, len(commands))
			for i, command := range commands {
				signatures[i] = command.Signature()
			}

			return assert.ObjectsAreEqual(expectedSignatures, signatures)
		})).Once()

		provider.Boot(app)
	})

	tests := []struct {
		name    string
		artisan func(t *testing.T) contractsconsole.Artisan
		config  func(t *testing.T) contractsconfig.Config
		log     func(t *testing.T) contractslog.Log
		schema  func(t *testing.T) contractsschema.Schema
		seeder  func(t *testing.T) contractsseeder.Facade
	}{
		{name: "artisan is nil", artisan: nilFactory[contractsconsole.Artisan]},
		{name: "config is nil", config: nilFactory[contractsconfig.Config]},
		{name: "log is nil", log: nilFactory[contractslog.Log]},
		{name: "schema is nil", schema: nilFactory[contractsschema.Schema]},
		{name: "seeder is nil", seeder: nilFactory[contractsseeder.Facade]},
	}

	for _, tt := range tests {
		t.Run("skips registration when "+tt.name, func(t *testing.T) {
			provider := &ServiceProvider{}
			app := mocksfoundation.NewApplication(t)

			app.EXPECT().MakeArtisan().Return(orDefault(t, tt.artisan, func(t *testing.T) contractsconsole.Artisan { return mocksconsole.NewArtisan(t) })).Once()
			app.EXPECT().MakeConfig().Return(orDefault(t, tt.config, func(t *testing.T) contractsconfig.Config { return mocksconfig.NewConfig(t) })).Once()
			app.EXPECT().MakeLog().Return(orDefault(t, tt.log, func(t *testing.T) contractslog.Log { return mockslog.NewLog(t) })).Once()
			app.EXPECT().MakeSchema().Return(orDefault(t, tt.schema, func(t *testing.T) contractsschema.Schema { return mocksschema.NewSchema(t) })).Once()
			app.EXPECT().MakeSeeder().Return(orDefault(t, tt.seeder, func(t *testing.T) contractsseeder.Facade { return mocksseeder.NewFacade(t) })).Once()

			// app.Commands is not expected; the mock fails the test if it is called.
			provider.Boot(app)
		})
	}
}

func nilFactory[T any](*testing.T) T {
	var zero T
	return zero
}

func orDefault[T any](t *testing.T, override, fallback func(*testing.T) T) T {
	if override != nil {
		return override(t)
	}

	return fallback(t)
}
