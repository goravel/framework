package hash

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/testing/mock"
)

type ApplicationTestSuite struct {
	suite.Suite
}

func TestApplicationTestSuite(t *testing.T) {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "hashing.driver", "argon2id").Return("argon2id")
	mockConfig.On("GetInt", "hashing.argon2id.memory", 65536).Return(65536)
	mockConfig.On("GetInt", "hashing.argon2id.time", 4).Return(4)
	mockConfig.On("GetInt", "hashing.argon2id.threads", 1).Return(1)

	facades.Hash = NewApplication()
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestArgon2idMakeHash() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "hashing.driver", "argon2id").Return("argon2id")
	mockConfig.On("GetInt", "hashing.argon2id.memory", 65536).Return(65536)
	mockConfig.On("GetInt", "hashing.argon2id.time", 4).Return(4)
	mockConfig.On("GetInt", "hashing.argon2id.threads", 1).Return(1)

	s.NotEmpty(facades.Hash.Make("password"))
}

func (s *ApplicationTestSuite) TestArgon2idCheckHash() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "hashing.driver", "argon2id").Return("argon2id")
	mockConfig.On("GetInt", "hashing.argon2id.memory", 65536).Return(65536)
	mockConfig.On("GetInt", "hashing.argon2id.time", 4).Return(4)
	mockConfig.On("GetInt", "hashing.argon2id.threads").Return(1)

	hash := facades.Hash.Make("password")
	s.True(facades.Hash.Check("password", hash))
}

func (s *ApplicationTestSuite) TestArgon2idNeedsRehash() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "hashing.driver", "argon2id").Return("argon2id")
	mockConfig.On("GetInt", "hashing.argon2id.memory", 65536).Return(65536)
	mockConfig.On("GetInt", "hashing.argon2id.time", 4).Return(4)
	mockConfig.On("GetInt", "hashing.argon2id.threads", 1).Return(1)

	hash := facades.Hash.Make("password")
	s.False(facades.Hash.NeedsRehash(hash))
}

func (s *ApplicationTestSuite) TestBcryptMakeHash() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "hashing.driver", "bcrypt").Return("bcrypt")
	mockConfig.On("GetInt", "hashing.bcrypt.cost", 10).Return(10)

	s.NotEmpty(facades.Hash.Make("password"))
}

func (s *ApplicationTestSuite) TestBcryptCheckHash() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "hashing.driver", "bcrypt").Return("bcrypt")
	mockConfig.On("GetInt", "hashing.bcrypt.cost", 10).Return(10)

	hash := facades.Hash.Make("password")
	s.True(facades.Hash.Check("password", hash))
}

func (s *ApplicationTestSuite) TestBcryptNeedsRehash() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "hashing.driver", "bcrypt").Return("bcrypt")
	mockConfig.On("GetInt", "hashing.bcrypt.cost", 10).Return(10)

	hash := facades.Hash.Make("password")
	s.False(facades.Hash.NeedsRehash(hash))
}
