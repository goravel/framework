package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	frameworkerrors "github.com/goravel/framework/errors"
	configmock "github.com/goravel/framework/mocks/config"
	foundationmock "github.com/goravel/framework/mocks/foundation"
)

func TestNewApplication(t *testing.T) {
	t.Run("uses bcrypt driver when configured", func(t *testing.T) {
		config := configmock.NewConfig(t)
		config.EXPECT().GetString("hashing.driver", "argon2id").Return(DriverBcrypt).Once()
		config.EXPECT().GetInt("hashing.bcrypt.rounds", 12).Return(10).Once()

		hasher := NewApplication(config)

		_, ok := hasher.(*Bcrypt)
		assert.True(t, ok)
	})

	t.Run("falls back to argon2id for unknown driver", func(t *testing.T) {
		config := configmock.NewConfig(t)
		config.EXPECT().GetString("hashing.driver", "argon2id").Return("unknown").Once()
		config.EXPECT().GetInt("hashing.argon2id.time", 4).Return(4).Once()
		config.EXPECT().GetInt("hashing.argon2id.memory", 65536).Return(65536).Once()
		config.EXPECT().GetInt("hashing.argon2id.threads", 1).Return(1).Once()

		hasher := NewApplication(config)

		_, ok := hasher.(*Argon2id)
		assert.True(t, ok)
	})
}

func TestBcryptMakeReturnsErrorWhenRoundsInvalid(t *testing.T) {
	config := configmock.NewConfig(t)
	config.EXPECT().GetInt("hashing.bcrypt.rounds", 12).Return(32).Once()

	hasher := NewBcrypt(config)
	hash, err := hasher.Make("password")

	assert.Error(t, err)
	assert.Empty(t, hash)
}

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()

	assert.Equal(t, []string{binding.Hash}, relationship.Bindings)
	assert.Equal(t, binding.Bindings[binding.Hash].Dependencies, relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}
	app := foundationmock.NewApplication(t)

	var callback func(foundation.Application) (any, error)
	app.On("Singleton", binding.Hash, mock.Anything).Run(func(args mock.Arguments) {
		callback = args.Get(1).(func(foundation.Application) (any, error))
	}).Once()

	provider.Register(app)
	assert.NotNil(t, callback)

	t.Run("returns hash application when config is available", func(t *testing.T) {
		config := configmock.NewConfig(t)
		app.On("MakeConfig").Return(config).Once()
		config.EXPECT().GetString("hashing.driver", "argon2id").Return(DriverBcrypt).Once()
		config.EXPECT().GetInt("hashing.bcrypt.rounds", 12).Return(10).Once()

		hash, err := callback(app)

		assert.NoError(t, err)
		_, ok := hash.(*Bcrypt)
		assert.True(t, ok)
	})

	t.Run("returns error when config facade is nil", func(t *testing.T) {
		app.On("MakeConfig").Return(nil).Once()

		hash, err := callback(app)

		assert.Nil(t, hash)
		assert.Error(t, err)
		assert.True(t, frameworkerrors.Is(err, frameworkerrors.ConfigFacadeNotSet))
	})
}

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}

	assert.NotPanics(t, func() {
		provider.Boot(nil)
	})
}
