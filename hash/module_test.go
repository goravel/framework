package hash

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	frameworkerrors "github.com/goravel/framework/errors"
	configmock "github.com/goravel/framework/mocks/config"
	foundationmock "github.com/goravel/framework/mocks/foundation"
)

const invalidDriver = "invalid-driver"
const expectedBcryptRounds = 10
const singletonCallbackType = "func(foundation.Application) (interface {}, error)"

func TestNewApplication(t *testing.T) {
	t.Run("uses bcrypt driver when configured", func(t *testing.T) {
		config := configmock.NewConfig(t)
		config.EXPECT().GetString("hashing.driver", "argon2id").Return(DriverBcrypt).Once()
		config.EXPECT().GetInt("hashing.bcrypt.rounds", 12).Return(expectedBcryptRounds).Once()

		hasher := NewApplication(config)

		bcryptHasher, ok := hasher.(*Bcrypt)
		require.True(t, ok)

		hashedValue, err := bcryptHasher.Make("test")
		require.NoError(t, err)
		cost, err := bcrypt.Cost([]byte(hashedValue))
		require.NoError(t, err)
		assert.Equal(t, expectedBcryptRounds, cost)
	})

	t.Run("falls back to argon2id for unknown driver", func(t *testing.T) {
		config := configmock.NewConfig(t)
		config.EXPECT().GetString("hashing.driver", "argon2id").Return(invalidDriver).Once()
		config.EXPECT().GetInt("hashing.argon2id.time", 4).Return(4).Once()
		config.EXPECT().GetInt("hashing.argon2id.memory", 65536).Return(65536).Once()
		config.EXPECT().GetInt("hashing.argon2id.threads", 1).Return(1).Once()

		hasher := NewApplication(config)

		_, ok := hasher.(*Argon2id)
		assert.True(t, ok)
	})
}

func TestBcrypt(t *testing.T) {
	t.Run("returns error when rounds exceed maximum", func(t *testing.T) {
		config := configmock.NewConfig(t)
		config.EXPECT().GetInt("hashing.bcrypt.rounds", 12).Return(32).Once()

		hasher := NewBcrypt(config)
		hash, err := hasher.Make("password")

		require.Error(t, err)
		assert.ErrorContains(t, err, "cost")
		assert.Empty(t, hash)
	})

	t.Run("uses default cost when rounds below minimum", func(t *testing.T) {
		config := configmock.NewConfig(t)
		config.EXPECT().GetInt("hashing.bcrypt.rounds", 12).Return(3).Once()

		hasher := NewBcrypt(config)
		hash, err := hasher.Make("password")
		require.NoError(t, err)

		cost, err := bcrypt.Cost([]byte(hash))
		require.NoError(t, err)
		assert.Equal(t, bcrypt.DefaultCost, cost)
	})
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
	app.On("Singleton", binding.Hash, mock.AnythingOfType(singletonCallbackType)).Run(func(args mock.Arguments) {
		cb, ok := args.Get(1).(func(foundation.Application) (any, error))
		require.True(t, ok)
		callback = cb
	}).Once()

	provider.Register(app)
	assert.NotNil(t, callback)

	t.Run("returns hash application when config is available", func(t *testing.T) {
		config := configmock.NewConfig(t)
		app.On("MakeConfig").Return(config).Once()
		config.EXPECT().GetString("hashing.driver", "argon2id").Return(DriverBcrypt).Once()
		config.EXPECT().GetInt("hashing.bcrypt.rounds", 12).Return(expectedBcryptRounds).Once()

		hash, err := callback(app)

		assert.NoError(t, err)
		_, ok := hash.(*Bcrypt)
		assert.True(t, ok)
	})

	t.Run("returns error when config facade is nil", func(t *testing.T) {
		app.On("MakeConfig").Return(nil).Once()

		_, err := callback(app)

		assert.Error(t, err)
		assert.True(t, frameworkerrors.Is(err, frameworkerrors.ConfigFacadeNotSet))
	})
}

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}
	app := foundationmock.NewApplication(t)

	t.Run("does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			provider.Boot(app)
		})
	})
}
