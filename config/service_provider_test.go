package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()

	assert.Equal(t, []string{binding.Config}, relationship.Bindings)
	assert.Empty(t, relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	t.Setenv("APP_KEY", "12345678901234567890123456789012")

	var callback func(foundation.Application) (any, error)
	app.EXPECT().Singleton(binding.Config, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, cb func(foundation.Application) (any, error)) {
		callback = cb
	}).Once()

	provider.Register(app)
	assert.NotNil(t, callback)

	instance, err := callback(app)

	assert.NoError(t, err)
	assert.IsType(t, &Application{}, instance)
}

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)

	assert.NotPanics(t, func() {
		provider.Boot(app)
	})
}
