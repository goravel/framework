package hash

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/hash"
	"github.com/goravel/framework/testing/mock"
)

type ApplicationTestSuite struct {
	suite.Suite
	hashers map[string]hash.Hash
}

func TestApplicationTestSuite(t *testing.T) {
	mockConfig := mock.Config()
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
		})
	}
}

func (s *ApplicationTestSuite) TestNeedsRehash() {
	for name, hasher := range s.hashers {
		s.Run(name, func() {
			value, err := hasher.Make("password")
			s.NoError(err)
			s.False(hasher.NeedsRehash(value))
			s.True(hasher.NeedsRehash("hash"))
		})
	}
}

func getArgon2idHasher(mockConfig *mocks.Config) *Argon2id {
	mockConfig.On("GetInt", "hashing.argon2id.memory", 65536).Return(65536).Once()
	mockConfig.On("GetInt", "hashing.argon2id.time", 4).Return(4).Once()
	mockConfig.On("GetInt", "hashing.argon2id.threads", 1).Return(1).Once()

	return NewArgon2id()
}

func getBcryptHasher(mockConfig *mocks.Config) *Bcrypt {
	mockConfig.On("GetInt", "hashing.bcrypt.rounds", 10).Return(10).Once()

	return NewBcrypt()
}
