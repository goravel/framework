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

	s.NotEmpty(facades.Hash.Make("password"))
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestArgon2idCheckHash() {
	mockConfig := mock.Config()

	hash := facades.Hash.Make("password")
	s.True(facades.Hash.Check("password", hash))
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestArgon2idNeedsRehash() {
	mockConfig := mock.Config()

	hash := facades.Hash.Make("password")
	s.False(facades.Hash.NeedsRehash(hash))
	mockConfig.AssertExpectations(s.T())
}
