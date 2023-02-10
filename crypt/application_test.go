package crypt

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/facades"
)

type ApplicationTestSuite struct {
	suite.Suite
}

func TestApplicationTestSuite(t *testing.T) {
	facades.Crypt = NewApplication()
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestEncryptString() {
	initConfig()
	s.NotEmpty(facades.Crypt.EncryptString("123"))
}

func (s *ApplicationTestSuite) TestDecryptString() {
	initConfig()
	iv, ciphertext := facades.Crypt.EncryptString("123")
	s.Equal("123", facades.Crypt.DecryptString(iv, ciphertext))
}

func initConfig() {
	application := config.NewApplication("../.env")
	application.Add("app", map[string]any{
		"key": "11111111111111111111111111111111",
	})

	facades.Config = application
}
