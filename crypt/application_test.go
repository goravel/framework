package crypt

import (
	"github.com/gookit/color"
	"github.com/goravel/framework/support/file"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/facades"
)

type ApplicationTestSuite struct {
	suite.Suite
}

func TestApplicationTestSuite(t *testing.T) {
	if !file.Exists("../.env") {
		color.Redln("No crypt tests run, need create .env based on .env.example, then initialize it")
		return
	}
	initConfig()
	facades.Crypt = NewApplication()
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestEncryptString() {
	if !file.Exists("../.env") {
		color.Redln("No crypt tests run, need create .env based on .env.example, then initialize it")
		return
	}
	initConfig()
	s.NotEmpty(facades.Crypt.EncryptString("Goravel"))
}

func (s *ApplicationTestSuite) TestDecryptString() {
	if !file.Exists("../.env") {
		color.Redln("No crypt tests run, need create .env based on .env.example, then initialize it")
		return
	}
	initConfig()
	iv, ciphertext := facades.Crypt.EncryptString("Goravel")
	s.Equal("Goravel", facades.Crypt.DecryptString(iv, ciphertext))
}

func initConfig() {
	application := config.NewApplication("../.env")
	application.Add("app", map[string]any{
		"key": "11111111111111111111111111111111",
	})

	facades.Config = application
}
