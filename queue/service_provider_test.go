package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	contractscache "github.com/goravel/framework/contracts/cache"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()

	assert.Equal(t, []string{binding.Queue}, relationship.Bindings)
	assert.Equal(t, binding.Bindings[binding.Queue].Dependencies, relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}

	t.Run("config facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Queue, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			callbackApp.EXPECT().MakeConfig().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.ConfigFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("log facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Queue, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			config := mocksconfig.NewConfig(t)
			callbackApp.EXPECT().MakeConfig().Return(config).Once()
			callbackApp.EXPECT().MakeLog().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.LogFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("register queue singleton", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Queue, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			config := mocksconfig.NewConfig(t)
			log := mockslog.NewLog(t)
			cache := mockscache.NewCache(t)
			db := mocksdb.NewDB(t)
			json := mocksfoundation.NewJson(t)

			callbackApp.EXPECT().MakeConfig().Return(config).Once()
			callbackApp.EXPECT().MakeLog().Return(log).Once()
			config.EXPECT().GetString("queue.default").Return("default").Once()
			config.EXPECT().GetString("queue.connections.default.queue", "default").Return("default").Once()
			config.EXPECT().GetInt("queue.connections.default.concurrent", 1).Return(1).Once()
			config.EXPECT().GetString("app.name", "goravel").Return("goravel").Once()
			config.EXPECT().GetBool("app.debug").Return(false).Once()
			config.EXPECT().GetString("queue.failed.database").Return("").Once()
			config.EXPECT().GetString("queue.failed.table").Return("").Once()
			callbackApp.EXPECT().MakeCache().Return(cache).Once()
			callbackApp.EXPECT().MakeDB().Return(db).Once()
			callbackApp.EXPECT().Json().Return(json).Once()

			instance, err := callback(callbackApp)

			assert.NoError(t, err)
			assert.IsType(t, &Application{}, instance)
		}).Once()

		provider.Register(app)
	})
}

var (
	_ contractscache.Cache = (*mockscache.Cache)(nil)
	_ contractsdb.DB       = (*mocksdb.DB)(nil)
)

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}

	t.Run("queue facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().MakeQueue().Return(nil).Once()

		provider.Boot(app)
	})

	t.Run("json facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		queue := mocksqueue.NewQueue(t)
		app.EXPECT().MakeQueue().Return(queue).Once()
		app.EXPECT().Json().Return(nil).Once()

		provider.Boot(app)
	})

	t.Run("register queue commands", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		queue := mocksqueue.NewQueue(t)
		json := mocksfoundation.NewJson(t)

		app.EXPECT().MakeQueue().Return(queue).Once()
		app.EXPECT().Json().Return(json).Once()
		app.EXPECT().Commands(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
			return len(commands) == 3 && commands[0] != nil && commands[1] != nil && commands[2] != nil
		})).Once()

		provider.Boot(app)
	})
}
