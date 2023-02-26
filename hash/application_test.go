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
	mockConfig.On("GetString", "hashing.driver", "argon2id").Return("argon2id").Once()
	mockConfig.On("GetInt", "hashing.argon2id.memory", 65536).Return(65536).Once()
	mockConfig.On("GetInt", "hashing.argon2id.time", 4).Return(4).Once()
	mockConfig.On("GetInt", "hashing.argon2id.threads", 1).Return(1).Once()

	facades.Hash = NewApplication()
	suite.Run(t, new(ApplicationTestSuite))
	mockConfig.AssertExpectations(t)
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestMakeHash() {
	s.NotEmpty(facades.Hash.Make("password"))

	mockConfig := mock.Config()
	mockConfig.On("GetString", "hashing.driver", "argon2id").Return("bcrypt").Once()
	mockConfig.On("GetInt", "hashing.bcrypt.rounds", 10).Return(10).Once()
	facades.Hash = NewApplication()
	s.NotEmpty(facades.Hash.Make("password"))
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestCheckHash() {
	hash, err := facades.Hash.Make("password")
	s.NoError(err)
	s.True(facades.Hash.Check("password", hash))

	mockConfig := mock.Config()
	mockConfig.On("GetString", "hashing.driver", "argon2id").Return("bcrypt").Once()
	mockConfig.On("GetInt", "hashing.bcrypt.rounds", 10).Return(10).Once()
	facades.Hash = NewApplication()
	hash, err = facades.Hash.Make("password")
	s.NoError(err)
	s.True(facades.Hash.Check("password", hash))
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestNeedsRehash() {
	hash, err := facades.Hash.Make("password")
	s.NoError(err)
	s.False(facades.Hash.NeedsRehash(hash))

	mockConfig := mock.Config()
	mockConfig.On("GetString", "hashing.driver", "argon2id").Return("bcrypt").Once()
	mockConfig.On("GetInt", "hashing.bcrypt.rounds", 10).Return(10).Once()
	facades.Hash = NewApplication()
	hash, err = facades.Hash.Make("password")
	s.NoError(err)
	s.False(facades.Hash.NeedsRehash(hash))
	mockConfig.AssertExpectations(s.T())
}
