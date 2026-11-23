package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsbinding "github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/binding"
)

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}
	bindings := []string{
		contractsbinding.Testing,
		contractsbinding.Fake,
	}

	assert.Equal(t, contractsbinding.Relationship{
		Bindings:     bindings,
		Dependencies: binding.Dependencies(bindings...),
		ProvideFor:   []string{},
	}, provider.Relationship())
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)

	var (
		testingCallback func(foundation.Application) (any, error)
		fakeCallback    func(foundation.Application) (any, error)
	)

	app.EXPECT().Singleton(contractsbinding.Testing, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, cb func(foundation.Application) (any, error)) {
		testingCallback = cb
	}).Once()
	app.EXPECT().Singleton(contractsbinding.Fake, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, cb func(foundation.Application) (any, error)) {
		fakeCallback = cb
	}).Once()

	provider.Register(app)

	t.Run("testing singleton returns error when config is nil", func(t *testing.T) {
		app.EXPECT().MakeConfig().Return(nil).Once()
		instance, err := testingCallback(app)
		assert.Nil(t, instance)
		assert.Equal(t, errors.ConfigFacadeNotSet.SetModule(errors.ModuleTesting), err)
	})

	t.Run("testing singleton returns Application when config is present", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockArtisan := mocksconsole.NewArtisan(t)
		app.EXPECT().MakeConfig().Return(mockConfig).Once()
		app.EXPECT().MakeArtisan().Return(mockArtisan).Once()
		app.EXPECT().MakeCache().Return(nil).Once()
		app.EXPECT().MakeOrm().Return(nil).Once()
		app.EXPECT().MakeProcess().Return(nil).Once()

		instance, err := testingCallback(app)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})

	t.Run("fake singleton returns gofakeit Faker instance", func(t *testing.T) {
		instance, err := fakeCallback(app)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
	})
}
