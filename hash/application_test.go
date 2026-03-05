package hash

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/hash"
	frameworkerrors "github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

type ApplicationTestSuite struct {
	suite.Suite
	config  *mocksconfig.Config
	hashers map[string]hash.Hash
}

const invalidDriver = "invalid-driver"
const expectedBcryptRounds = 10
const singletonCallbackType = "func(foundation.Application) (interface {}, error)"

func TestApplicationTestSuite(t *testing.T) {
	suite.Run(t, &ApplicationTestSuite{})
}

func (s *ApplicationTestSuite) SetupTest() {
	s.config = &mocksconfig.Config{}
	s.hashers = map[string]hash.Hash{
		"argon2id": getArgon2idHasher(s.config),
		"bcrypt":   getBcryptHasher(s.config),
	}
}

func (s *ApplicationTestSuite) TearDownSuite() {
	s.config.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeHash() {
	for name, hasher := range s.hashers {
		s.Run(name, func() {
			s.NotEmpty(hasher.Make("password"))
		})
	}
}

func (s *ApplicationTestSuite) TestCheckHash() {
	for name, hasher := range s.hashers {
		s.Run(name, func() {
			value, err := hasher.Make("password")
			s.NoError(err)
			s.True(hasher.Check("password", value))
			s.False(hasher.Check("password1", value))
			s.False(hasher.Check("password", "hash"))
			s.False(hasher.Check("password", "hashhash"))
			s.False(hasher.Check("password", "$argon2id$v=20$m=16,t=2,p=1$dTltTmtGb0JmNE9Zb0lTeQ$2lHJsAodBnV4u7j39gj7Uw"))
			s.False(hasher.Check("password", "$argon2id$v=$m=16,t=2,p=1$dTltTmtGb0JmNE9Zb0lTeQ$2lHJsAodBnV4u7j39gj7Uw"))
			s.False(hasher.Check("password", "$argon2id$v=19$m=16,t=2$dTltTmtGb0JmNE9Zb0lTeQ$2lHJsAodBnV4u7j39gj7Uw"))
			s.False(hasher.Check("password", "$argon2id$v=19$m=16,t=2,p=1$dTltTmtGb0JmNE9Zb0lTeQ$123456"))
			s.False(hasher.Check("password", "$argon2id$v=19$m=16,t=2,p=1$123456$2lHJsAodBnV4u7j39gj7xx"))
		})
	}
}

func (s *ApplicationTestSuite) TestConfigurationOverride() {
	value := "$argon2id$v=19$m=65536,t=8,p=1$NlVjQm5PQUdWTHVTM1RBUg$Q5T7WfeCI7ucIdk6Na6AdQ"
	s.NotNil(s.hashers["argon2id"])
	s.True(s.hashers["argon2id"].Check("goravel", value))
	s.True(s.hashers["argon2id"].NeedsRehash(value))
}

func (s *ApplicationTestSuite) TestNeedsRehash() {
	for name, hasher := range s.hashers {
		s.Run(name, func() {
			value, err := hasher.Make("password")
			s.NoError(err)
			s.False(hasher.NeedsRehash(value))
			s.True(hasher.NeedsRehash("hash"))
			s.True(hasher.NeedsRehash("hashhash"))
			s.True(hasher.NeedsRehash("$argon2id$v=$m=16,t=2,p=1$dTltTmtGb0JmNE9Zb0lTeQ$2lHJsAodBnV4u7j39gj7Uw"))
			s.True(hasher.NeedsRehash("$argon2id$v=19$m=16,t=2$dTltTmtGb0JmNE9Zb0lTeQ$2lHJsAodBnV4u7j39gj7Uw"))
		})
	}
}

func TestNewApplication(t *testing.T) {
	t.Run("uses bcrypt driver when configured", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
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
		config := mocksconfig.NewConfig(t)
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
		config := mocksconfig.NewConfig(t)
		config.EXPECT().GetInt("hashing.bcrypt.rounds", 12).Return(32).Once()

		hasher := NewBcrypt(config)
		hash, err := hasher.Make("password")

		require.Error(t, err)
		assert.ErrorContains(t, err, "cost")
		assert.Empty(t, hash)
	})

	t.Run("uses default cost when rounds below minimum", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
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
	mockApp := mocksfoundation.NewApplication(t)

	var callback func(foundation.Application) (any, error)
	mockApp.EXPECT().Singleton(binding.Hash, mock.AnythingOfType(singletonCallbackType)).Run(func(key any, cb func(foundation.Application) (any, error)) {
		assert.Equal(t, binding.Hash, key)
		callback = cb
	}).Once()

	provider.Register(mockApp)
	assert.NotNil(t, callback)

	t.Run("returns hash application when config is available", func(t *testing.T) {
		config := mocksconfig.NewConfig(t)
		mockApp.EXPECT().MakeConfig().Return(config).Once()
		config.EXPECT().GetString("hashing.driver", "argon2id").Return(DriverBcrypt).Once()
		config.EXPECT().GetInt("hashing.bcrypt.rounds", 12).Return(expectedBcryptRounds).Once()

		hash, err := callback(mockApp)

		assert.NoError(t, err)
		_, ok := hash.(*Bcrypt)
		assert.True(t, ok)
	})

	t.Run("returns error when config facade is nil", func(t *testing.T) {
		mockApp.EXPECT().MakeConfig().Return(nil).Once()

		_, err := callback(mockApp)

		assert.Error(t, err)
		assert.True(t, frameworkerrors.Is(err, frameworkerrors.ConfigFacadeNotSet))
	})
}

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)

	t.Run("does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			provider.Boot(app)
		})
	})
}

func BenchmarkMakeHash(b *testing.B) {
	s := new(ApplicationTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()
	b.StartTimer()
	b.ResetTimer()
	for name, hasher := range s.hashers {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := hasher.Make("password")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	b.StopTimer()
}

func BenchmarkCheckHash(b *testing.B) {
	s := new(ApplicationTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()
	b.StartTimer()
	b.ResetTimer()
	for name, hasher := range s.hashers {
		b.Run(name, func(b *testing.B) {
			value, err := hasher.Make("password")
			if err != nil {
				b.Fatal(err)
			}
			for i := 0; i < b.N; i++ {
				if !hasher.Check("password", value) {
					b.Fatal("hash check failed")
				}
			}
		})
	}
	b.StopTimer()
}

func BenchmarkNeedsRehash(b *testing.B) {
	s := new(ApplicationTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()
	b.StartTimer()
	b.ResetTimer()
	for name, hasher := range s.hashers {
		b.Run(name, func(b *testing.B) {
			value, err := hasher.Make("password")
			if err != nil {
				b.Fatal(err)
			}
			for i := 0; i < b.N; i++ {
				hasher.NeedsRehash(value)
			}
		})
	}
	b.StopTimer()
}

func getArgon2idHasher(mockConfig *mocksconfig.Config) *Argon2id {
	mockConfig.On("GetInt", "hashing.argon2id.memory", 65536).Return(65536).Once()
	mockConfig.On("GetInt", "hashing.argon2id.time", 4).Return(4).Once()
	mockConfig.On("GetInt", "hashing.argon2id.threads", 1).Return(1).Once()

	return NewArgon2id(mockConfig)
}

func getBcryptHasher(mockConfig *mocksconfig.Config) *Bcrypt {
	mockConfig.On("GetInt", "hashing.bcrypt.rounds", 12).Return(10).Once()

	return NewBcrypt(mockConfig)
}
