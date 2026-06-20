package console

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksprocess "github.com/goravel/framework/mocks/process"
)

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()

	assert.Equal(t, []string{binding.Artisan}, relationship.Bindings)
	assert.Equal(t, binding.Bindings[binding.Artisan].Dependencies, relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}

	t.Run("register artisan singleton", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Singleton(binding.Artisan, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			callbackApp.EXPECT().Version().Return("v1.18.0").Once()

			instance, err := callback(callbackApp)

			assert.NoError(t, err)
			assert.IsType(t, &Application{}, instance)
		}).Once()

		provider.Register(app)
	})
}

func signaturesOf(commands []contractsconsole.Command) []string {
	sigs := make([]string, 0, len(commands))
	for _, c := range commands {
		sigs = append(sigs, c.Signature())
	}
	sort.Strings(sigs)
	return sigs
}

func TestServiceProviderBootNilAllowlistRegistersAll(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	artisan := mocksconsole.NewArtisan(t)
	configFacade := mocksconfig.NewConfig(t)
	processFacade := mocksprocess.NewProcess(t)

	// nil allowlist means all four console/console/* commands are registered.
	app.EXPECT().MakeArtisan().Return(artisan).Once()
	app.EXPECT().MakeConfig().Return(configFacade).Once()
	app.EXPECT().MakeProcess().Return(processFacade).Once()
	app.EXPECT().CommandsFilter().Return(nil).Once()
	artisan.EXPECT().Register(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		return len(commands) == 4
	})).Once()

	provider.Boot(app)
}

func TestServiceProviderBootEmptyAllowlistRegistersNothing(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	artisan := mocksconsole.NewArtisan(t)
	configFacade := mocksconfig.NewConfig(t)
	processFacade := mocksprocess.NewProcess(t)

	// empty allowlist means all commands are dropped.
	app.EXPECT().MakeArtisan().Return(artisan).Once()
	app.EXPECT().MakeConfig().Return(configFacade).Once()
	app.EXPECT().MakeProcess().Return(processFacade).Once()
	app.EXPECT().CommandsFilter().Return([]string{}).Once()
	artisan.EXPECT().Register(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		return len(commands) == 0
	})).Once()

	provider.Boot(app)
}

func TestServiceProviderBootAllowlistApplies(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	artisan := mocksconsole.NewArtisan(t)
	configFacade := mocksconfig.NewConfig(t)
	processFacade := mocksprocess.NewProcess(t)

	// glob pattern keep-list: only "list" and the make sub-command survive.
	app.EXPECT().MakeArtisan().Return(artisan).Once()
	app.EXPECT().MakeConfig().Return(configFacade).Once()
	app.EXPECT().MakeProcess().Return(processFacade).Once()
	app.EXPECT().CommandsFilter().Return([]string{"list", "make:*"}).Once()
	artisan.EXPECT().Register(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		got := signaturesOf(commands)
		want := []string{"list", "make:command"}
		sort.Strings(want)
		if len(got) != len(want) {
			return false
		}
		for i := range got {
			if got[i] != want[i] {
				return false
			}
		}
		return true
	})).Once()

	provider.Boot(app)
}
