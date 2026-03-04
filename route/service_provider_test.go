package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	consolecontract "github.com/goravel/framework/contracts/console"
	foundationcontract "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	configmock "github.com/goravel/framework/mocks/config"
	consolemock "github.com/goravel/framework/mocks/console"
	foundationmock "github.com/goravel/framework/mocks/foundation"
	routemock "github.com/goravel/framework/mocks/route"
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
		app := foundationmock.NewApplication(t)
		app.EXPECT().Singleton(binding.Route, mock.Anything).Run(func(_ interface{}, callback func(foundationcontract.Application) (interface{}, error)) {
			callbackApp := foundationmock.NewApplication(t)
			callbackApp.EXPECT().MakeConfig().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.ConfigFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("register route singleton", func(t *testing.T) {
		app := foundationmock.NewApplication(t)
		app.EXPECT().Singleton(binding.Route, mock.Anything).Run(func(_ interface{}, callback func(foundationcontract.Application) (interface{}, error)) {
			callbackApp := foundationmock.NewApplication(t)
			config := configmock.NewConfig(t)
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
	app := foundationmock.NewApplication(t)
	artisan := consolemock.NewArtisan(t)
	route := routemock.NewRoute(t)

	app.EXPECT().MakeArtisan().Return(artisan).Once()
	app.EXPECT().MakeRoute().Return(route).Once()
	artisan.EXPECT().Register(mock.MatchedBy(func(commands []consolecontract.Command) bool {
		return len(commands) == 1 && commands[0] != nil
	})).Once()

	provider.Boot(app)
}

func TestServiceProviderRunners(t *testing.T) {
	provider := &ServiceProvider{}
	app := foundationmock.NewApplication(t)
	config := configmock.NewConfig(t)
	route := routemock.NewRoute(t)

	app.EXPECT().MakeConfig().Return(config).Once()
	app.EXPECT().MakeRoute().Return(route).Once()

	runners := provider.Runners(app)

	assert.Len(t, runners, 1)
	runner, ok := runners[0].(*RouteRunner)
	assert.True(t, ok)
	assert.Equal(t, config, runner.config)
	assert.Equal(t, route, runner.route)
}
