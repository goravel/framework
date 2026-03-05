package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksroute "github.com/goravel/framework/mocks/route"
)

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()

	assert.Equal(t, []string{binding.Route}, relationship.Bindings)
	assert.Equal(t, binding.Bindings[binding.Route].Dependencies, relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}

	t.Run("config facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Route, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			callbackApp.EXPECT().MakeConfig().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.ConfigFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("register route singleton", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Route, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			config := mocksconfig.NewConfig(t)
			callbackApp.EXPECT().MakeConfig().Return(config).Once()
			config.EXPECT().GetString("http.default").Return("").Once()

			instance, err := callback(callbackApp)

			assert.NoError(t, err)
			routeInstance, ok := instance.(*Route)
			assert.True(t, ok)
			assert.Equal(t, config, routeInstance.config)
		}).Once()

		provider.Register(app)
	})
}

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	artisan := mocksconsole.NewArtisan(t)
	route := mocksroute.NewRoute(t)

	app.EXPECT().MakeArtisan().Return(artisan).Once()
	app.EXPECT().MakeRoute().Return(route).Once()
	artisan.EXPECT().Register(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		return len(commands) == 1 && commands[0] != nil
	})).Once()

	provider.Boot(app)
}

func TestServiceProviderRunners(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	config := mocksconfig.NewConfig(t)
	route := mocksroute.NewRoute(t)

	app.EXPECT().MakeConfig().Return(config).Once()
	app.EXPECT().MakeRoute().Return(route).Once()

	runners := provider.Runners(app)

	assert.Len(t, runners, 1)
	runner, ok := runners[0].(*RouteRunner)
	assert.True(t, ok)
	assert.Equal(t, config, runner.config)
	assert.Equal(t, route, runner.route)
}
