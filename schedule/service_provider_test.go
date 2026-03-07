package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksschedule "github.com/goravel/framework/mocks/schedule"
)

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()

	assert.Equal(t, []string{binding.Schedule}, relationship.Bindings)
	assert.Equal(t, binding.Bindings[binding.Schedule].Dependencies, relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}

	t.Run("config facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Schedule, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			callbackApp.EXPECT().MakeConfig().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.ConfigFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("console facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Schedule, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			config := mocksconfig.NewConfig(t)
			callbackApp.EXPECT().MakeConfig().Return(config).Once()
			callbackApp.EXPECT().MakeArtisan().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.ConsoleFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("log facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Schedule, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			config := mocksconfig.NewConfig(t)
			artisan := mocksconsole.NewArtisan(t)
			callbackApp.EXPECT().MakeConfig().Return(config).Once()
			callbackApp.EXPECT().MakeArtisan().Return(artisan).Once()
			callbackApp.EXPECT().MakeLog().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.LogFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("cache facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Schedule, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			config := mocksconfig.NewConfig(t)
			artisan := mocksconsole.NewArtisan(t)
			log := mockslog.NewLog(t)
			callbackApp.EXPECT().MakeConfig().Return(config).Once()
			callbackApp.EXPECT().MakeArtisan().Return(artisan).Once()
			callbackApp.EXPECT().MakeLog().Return(log).Once()
			callbackApp.EXPECT().MakeCache().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.CacheFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("register schedule singleton", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Schedule, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			config := mocksconfig.NewConfig(t)
			artisan := mocksconsole.NewArtisan(t)
			log := mockslog.NewLog(t)
			cache := mockscache.NewCache(t)
			callbackApp.EXPECT().MakeConfig().Return(config).Once()
			callbackApp.EXPECT().MakeArtisan().Return(artisan).Once()
			callbackApp.EXPECT().MakeLog().Return(log).Once()
			callbackApp.EXPECT().MakeCache().Return(cache).Once()
			config.EXPECT().GetBool("app.debug").Return(true).Once()

			instance, err := callback(callbackApp)

			assert.NoError(t, err)
			assert.IsType(t, &Application{}, instance)
		}).Once()

		provider.Register(app)
	})
}

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	artisan := mocksconsole.NewArtisan(t)
	schedule := mocksschedule.NewSchedule(t)

	app.EXPECT().MakeArtisan().Return(artisan).Once()
	app.EXPECT().MakeSchedule().Return(schedule).Twice()
	artisan.EXPECT().Register(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		if len(commands) != 2 || commands[0] == nil || commands[1] == nil {
			return false
		}

		_, okList := commands[0].(interface{ Signature() string })
		_, okRun := commands[1].(interface{ Signature() string })

		return okList && okRun && commands[0].Signature() == "schedule:list" && commands[1].Signature() == "schedule:run"
	})).Once()

	provider.Boot(app)
}

func TestServiceProviderRunners(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	config := mocksconfig.NewConfig(t)
	schedule := mocksschedule.NewSchedule(t)

	app.EXPECT().MakeConfig().Return(config).Once()
	app.EXPECT().MakeSchedule().Return(schedule).Once()

	runners := provider.Runners(app)

	assert.Len(t, runners, 1)
	runner, ok := runners[0].(*ScheduleRunner)
	assert.True(t, ok)
	assert.Equal(t, config, runner.config)
	assert.Equal(t, schedule, runner.schedule)
}
