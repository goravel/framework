package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsauthaccess "github.com/goravel/framework/contracts/auth/access"
	contractsbinding "github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	frameworkerrors "github.com/goravel/framework/errors"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockshttp "github.com/goravel/framework/mocks/http"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mockssession "github.com/goravel/framework/mocks/session"
	"github.com/goravel/framework/support/binding"
)

const expectedAuthCommandCount = 2

func TestAuthServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()
	bindings := []string{contractsbinding.Auth, contractsbinding.Gate}

	assert.Equal(t, bindings, relationship.Bindings)
	assert.Equal(t, binding.Dependencies(bindings...), relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestAuthServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)

	var authCallback func(contractsfoundation.Application, map[string]any) (any, error)
	var gateCallback func(contractsfoundation.Application) (any, error)
	app.EXPECT().BindWith(contractsbinding.Auth, mock.AnythingOfType("func(foundation.Application, map[string]interface {}) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application, map[string]any) (any, error)) {
		authCallback = callback
	}).Once()
	app.EXPECT().Singleton(contractsbinding.Gate, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
		gateCallback = callback
	}).Once()

	provider.Register(app)
	assert.NotNil(t, authCallback)
	assert.NotNil(t, gateCallback)

	t.Run("returns error when config facade is nil", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		callbackApp.EXPECT().MakeConfig().Return(nil).Once()

		instance, err := authCallback(callbackApp, map[string]any{})

		assert.Nil(t, instance)
		assert.Error(t, err)
		assert.True(t, frameworkerrors.Is(err, frameworkerrors.ConfigFacadeNotSet))
	})

	t.Run("returns error when log facade is nil", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		config := mocksconfig.NewConfig(t)
		callbackApp.EXPECT().MakeConfig().Return(config).Once()
		callbackApp.EXPECT().MakeLog().Return(nil).Once()

		instance, err := authCallback(callbackApp, map[string]any{})

		assert.Nil(t, instance)
		assert.Error(t, err)
		assert.True(t, frameworkerrors.Is(err, frameworkerrors.LogFacadeNotSet))
	})

	t.Run("creates auth instance without context", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		config := mocksconfig.NewConfig(t)
		log := mockslog.NewLog(t)
		callbackApp.EXPECT().MakeConfig().Return(config).Once()
		callbackApp.EXPECT().MakeLog().Return(log).Once()
		config.EXPECT().GetString("auth.defaults.guard").Return("").Once()

		instance, err := authCallback(callbackApp, map[string]any{})

		assert.NoError(t, err)
		assert.IsType(t, &Auth{}, instance)
	})

	t.Run("creates auth instance with context parameter", func(t *testing.T) {
		callbackApp := mocksfoundation.NewApplication(t)
		config := mocksconfig.NewConfig(t)
		log := mockslog.NewLog(t)
		ctx := mockshttp.NewContext(t)
		request := mockshttp.NewContextRequest(t)
		session := mockssession.NewSession(t)
		originOrmFacade := ormFacade
		ormFacade = mocksorm.NewOrm(t)
		t.Cleanup(func() {
			ormFacade = originOrmFacade
		})

		callbackApp.EXPECT().MakeConfig().Return(config).Once()
		callbackApp.EXPECT().MakeLog().Return(log).Once()
		config.EXPECT().GetString("auth.defaults.guard").Return("web").Once()
		config.EXPECT().GetString("auth.guards.web.driver").Return("session").Once()
		config.EXPECT().GetString("auth.guards.web.provider").Return("users").Once()
		config.EXPECT().GetString("auth.providers.users.driver").Return("orm").Once()
		ctx.EXPECT().Request().Return(request).Once()
		request.EXPECT().Session().Return(session).Once()

		instance, err := authCallback(callbackApp, map[string]any{
			"ctx": ctx,
		})

		assert.NoError(t, err)
		auth, ok := instance.(*Auth)
		assert.True(t, ok)
		assert.Same(t, ctx, auth.ctx)
		assert.NotNil(t, auth.GuardDriver)
	})

	t.Run("registers gate singleton", func(t *testing.T) {
		instance, err := gateCallback(app)

		assert.NoError(t, err)
		assert.Implements(t, (*contractsauthaccess.Gate)(nil), instance)
	})
}

func TestAuthServiceProviderBoot(t *testing.T) {
	// This test mutates package-level facades and should not run in parallel.
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	cache := mockscache.NewCache(t)
	config := mocksconfig.NewConfig(t)
	orm := mocksorm.NewOrm(t)

	originCacheFacade := cacheFacade
	originOrmFacade := ormFacade
	t.Cleanup(func() {
		cacheFacade = originCacheFacade
		ormFacade = originOrmFacade
	})

	app.EXPECT().MakeCache().Return(cache).Once()
	app.EXPECT().MakeOrm().Return(orm).Once()
	app.EXPECT().MakeConfig().Return(config).Once()
	app.EXPECT().Commands(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		if len(commands) != expectedAuthCommandCount {
			t.Logf("expected %d auth commands, got %d", expectedAuthCommandCount, len(commands))
			return false
		}
		for i, command := range commands {
			if command == nil {
				t.Logf("auth command %d is nil", i)
				return false
			}
		}

		return true
	})).Once()

	provider.Boot(app)

	assert.Same(t, cache, cacheFacade)
	assert.Same(t, orm, ormFacade)
}
