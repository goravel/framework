package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()

	assert.Equal(t, []string{binding.Log}, relationship.Bindings)
	assert.Equal(t, binding.Bindings[binding.Log].Dependencies, relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}

	t.Run("config facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Log, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			callbackApp.EXPECT().MakeConfig().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.ConfigFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("json parser not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Log, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			config := mocksconfig.NewConfig(t)
			callbackApp.EXPECT().MakeConfig().Return(config).Once()
			callbackApp.EXPECT().Json().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.JSONParserNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("register log singleton", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Log, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			config := mocksconfig.NewConfig(t)
			json := mocksfoundation.NewJson(t)
			callbackApp.EXPECT().MakeConfig().Return(config).Once()
			callbackApp.EXPECT().Json().Return(json).Once()
			config.EXPECT().GetString("logging.default").Return("").Once()

			instance, err := callback(callbackApp)

			assert.NoError(t, err)
			assert.IsType(t, &Application{}, instance)
		}).Once()

		provider.Register(app)
	})
}

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}
	assert.NotPanics(t, func() {
		provider.Boot(nil)
	})
}
