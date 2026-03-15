package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()

	assert.Equal(t, []string{binding.Process}, relationship.Bindings)
	assert.Equal(t, binding.Bindings[binding.Process].Dependencies, relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	app.EXPECT().Bind(
		binding.Process,
		mock.AnythingOfType("func(foundation.Application) (interface {}, error)"),
	).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
		instance, err := callback(mocksfoundation.NewApplication(t))
		assert.NoError(t, err)
		assert.IsType(t, &Process{}, instance)
	}).Once()

	provider.Register(app)
}

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}
	assert.NotPanics(t, func() {
		provider.Boot(nil)
	})
}
