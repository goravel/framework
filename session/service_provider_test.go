package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockssession "github.com/goravel/framework/mocks/session"
	"github.com/goravel/framework/support/path"
)

func TestSessionServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()

	assert.Equal(t, []string{binding.Session}, relationship.Bindings)
	assert.Equal(t, binding.Bindings[binding.Session].Dependencies, relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestSessionServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)

	var callback func(contractsfoundation.Application) (any, error)
	app.EXPECT().Singleton(binding.Session, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, cb func(contractsfoundation.Application) (any, error)) {
		callback = cb
	}).Once()

	provider.Register(app)
	assert.NotNil(t, callback)

	t.Run("returns error when config facade is nil", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		callbackApp.EXPECT().MakeConfig().Return(nil).Once()

		instance, err := callback(callbackApp)

		assert.Nil(t, instance)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.ConfigFacadeNotSet))
	})

	t.Run("returns error when json parser is nil", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		config := mocksconfig.NewConfig(t)
		callbackApp.EXPECT().MakeConfig().Return(config).Once()
		callbackApp.EXPECT().GetJson().Return(nil).Once()

		instance, err := callback(callbackApp)

		assert.Nil(t, instance)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.JSONParserNotSet))
	})

	t.Run("creates session manager when dependencies are available", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		config := mocksconfig.NewConfig(t)
		j := json.New()
		callbackApp.EXPECT().MakeConfig().Return(config).Once()
		callbackApp.EXPECT().GetJson().Return(j).Once()
		config.EXPECT().GetString("session.cookie").Return("goravel_session").Once()
		config.EXPECT().GetString("session.default", "file").Return("file").Once()
		config.EXPECT().GetString("session.files").Return(path.Storage("framework/sessions")).Once()
		config.EXPECT().GetInt("session.gc_interval", 30).Return(30).Once()
		config.EXPECT().GetInt("session.lifetime", 120).Return(120).Once()
		config.EXPECT().GetString("session.drivers.file.driver").Return("file").Once()

		instance, err := callback(callbackApp)

		assert.NoError(t, err)
		assert.IsType(t, &Manager{}, instance)
	})
}

func TestSessionServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	session := mockssession.NewManager(t)
	config := mocksconfig.NewConfig(t)

	originSessionFacade := SessionFacade
	originConfigFacade := ConfigFacade
	t.Cleanup(func() {
		SessionFacade = originSessionFacade
		ConfigFacade = originConfigFacade
	})

	app.EXPECT().MakeSession().Return(session).Once()
	app.EXPECT().MakeConfig().Return(config).Once()

	provider.Boot(app)

	assert.Same(t, session, SessionFacade)
	assert.Same(t, config, ConfigFacade)
}
