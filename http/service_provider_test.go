package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsbinding "github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	frameworkerrors "github.com/goravel/framework/errors"
	"github.com/goravel/framework/http/client"
	foundationjson "github.com/goravel/framework/foundation/json"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/binding"
)

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()
	bindings := []string{contractsbinding.Http, contractsbinding.RateLimiter, contractsbinding.View}

	assert.Equal(t, bindings, relationship.Bindings)
	assert.Equal(t, binding.Dependencies(bindings...), relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)

	var rateLimiterCallback func(contractsfoundation.Application) (any, error)
	var httpCallback func(contractsfoundation.Application) (any, error)
	app.EXPECT().Singleton(contractsbinding.RateLimiter, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
		rateLimiterCallback = callback
	}).Once()
	app.EXPECT().Singleton(contractsbinding.Http, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
		httpCallback = callback
	}).Once()

	provider.Register(app)
	assert.NotNil(t, rateLimiterCallback)
	assert.NotNil(t, httpCallback)

	t.Run("creates rate limiter singleton", func(t *testing.T) {
		instance, err := rateLimiterCallback(app)

		assert.NoError(t, err)
		assert.IsType(t, &RateLimiter{}, instance)
	})

	t.Run("returns error when config facade is nil", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		callbackApp.EXPECT().MakeConfig().Return(nil).Once()

		instance, err := httpCallback(callbackApp)

		assert.Nil(t, instance)
		assert.Error(t, err)
		assert.True(t, frameworkerrors.Is(err, frameworkerrors.ConfigFacadeNotSet))
	})

	t.Run("returns error when json parser is nil", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		config := mocksconfig.NewConfig(t)
		callbackApp.EXPECT().MakeConfig().Return(config).Once()
		callbackApp.EXPECT().Json().Return(nil).Once()

		instance, err := httpCallback(callbackApp)

		assert.Nil(t, instance)
		assert.Error(t, err)
		assert.True(t, frameworkerrors.Is(err, frameworkerrors.JSONParserNotSet))
	})

	t.Run("returns unmarshal error", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		config := mocksconfig.NewConfig(t)
		j := foundationjson.New()
		callbackApp.EXPECT().MakeConfig().Return(config).Once()
		callbackApp.EXPECT().Json().Return(j).Once()
		config.EXPECT().UnmarshalKey("http", mock.AnythingOfType("*client.FactoryConfig")).Return(assert.AnError).Once()

		instance, err := httpCallback(callbackApp)

		assert.Nil(t, instance)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("creates http factory", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		config := mocksconfig.NewConfig(t)
		j := foundationjson.New()
		callbackApp.EXPECT().MakeConfig().Return(config).Once()
		callbackApp.EXPECT().Json().Return(j).Once()
		config.EXPECT().UnmarshalKey("http", mock.AnythingOfType("*client.FactoryConfig")).RunAndReturn(func(_ string, target any) error {
			factoryConfig, ok := target.(*client.FactoryConfig)
			if !ok {
				return assert.AnError
			}
			factoryConfig.Default = "default"
			factoryConfig.Clients = map[string]client.Config{
				"default": {},
			}

			return nil
		}).Once()

		instance, err := httpCallback(callbackApp)

		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("returns error when default client is missing from clients", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		config := mocksconfig.NewConfig(t)
		j := foundationjson.New()
		callbackApp.EXPECT().MakeConfig().Return(config).Once()
		callbackApp.EXPECT().Json().Return(j).Once()
		config.EXPECT().UnmarshalKey("http", mock.AnythingOfType("*client.FactoryConfig")).RunAndReturn(func(_ string, target any) error {
			factoryConfig, ok := target.(*client.FactoryConfig)
			if !ok {
				return assert.AnError
			}
			factoryConfig.Default = "missing_client"
			return nil
		}).Once()

		instance, err := httpCallback(callbackApp)

		assert.Nil(t, instance)
		assert.Error(t, err)
		assert.True(t, frameworkerrors.Is(err, frameworkerrors.HttpClientConnectionNotFound.Args("missing_client")))
	})
}

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	originApp := App
	t.Cleanup(func() {
		App = originApp
	})

	app.EXPECT().Commands(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		return len(commands) == 3 && commands[0] != nil && commands[1] != nil && commands[2] != nil
	})).Once()

	provider.Boot(app)

	assert.Same(t, app, App)
}
