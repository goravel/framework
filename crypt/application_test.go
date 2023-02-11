package crypt

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
	mockConfig.On("GetString", "app.key").Return("11111111111111111111111111111111")

	facades.Crypt = NewApplication()
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestEncryptString() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "app.key").Return("11111111111111111111111111111111")

	s.NotEmpty(facades.Crypt.EncryptString("Goravel"))
}

func (s *ApplicationTestSuite) TestDecryptString() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "app.key").Return("11111111111111111111111111111111")

	iv, ciphertext := facades.Crypt.EncryptString("Goravel")
	s.Equal("Goravel", facades.Crypt.DecryptString(iv, ciphertext))
}
