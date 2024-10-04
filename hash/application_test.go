package hash

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/hash"
	configmock "github.com/goravel/framework/mocks/config"
)

type ApplicationTestSuite struct {
	suite.Suite
	config  *configmock.Config
	hashers map[string]hash.Hash
}

func TestApplicationTestSuite(t *testing.T) {
	suite.Run(t, &ApplicationTestSuite{})
}

func (s *ApplicationTestSuite) SetupTest() {
	s.config = &configmock.Config{}
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

func getArgon2idHasher(mockConfig *configmock.Config) *Argon2id {
	mockConfig.On("GetInt", "hashing.argon2id.memory", 65536).Return(65536).Once()
	mockConfig.On("GetInt", "hashing.argon2id.time", 4).Return(4).Once()
	mockConfig.On("GetInt", "hashing.argon2id.threads", 1).Return(1).Once()

	return NewArgon2id(mockConfig)
}

func getBcryptHasher(mockConfig *configmock.Config) *Bcrypt {
	mockConfig.On("GetInt", "hashing.bcrypt.rounds", 12).Return(10).Once()

	return NewBcrypt(mockConfig)
}
