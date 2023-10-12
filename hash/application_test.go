package hash

import (
	"testing"

	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/hash"
)

type ApplicationTestSuite struct {
	suite.Suite
	hashers map[string]hash.Hash
}

func TestApplicationTestSuite(t *testing.T) {
	mockConfig := &configmock.Config{}
	argon2idHasher := getArgon2idHasher(mockConfig)
	bcryptHasher := getBcryptHasher(mockConfig)

	suite.Run(t, &ApplicationTestSuite{
		hashers: map[string]hash.Hash{
			"argon2id": argon2idHasher,
			"bcrypt":   bcryptHasher,
		},
	})
	mockConfig.AssertExpectations(t)
}

func (s *ApplicationTestSuite) SetupTest() {

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

func getArgon2idHasher(mockConfig *configmock.Config) *Argon2id {
	mockConfig.On("GetInt", "hashing.argon2id.memory", 65536).Return(65536).Once()
	mockConfig.On("GetInt", "hashing.argon2id.time", 4).Return(4).Once()
	mockConfig.On("GetInt", "hashing.argon2id.threads", 1).Return(1).Once()

	return NewArgon2id(mockConfig)
}

func getBcryptHasher(mockConfig *configmock.Config) *Bcrypt {
	mockConfig.On("GetInt", "hashing.bcrypt.rounds", 10).Return(10).Once()

	return NewBcrypt(mockConfig)
}
