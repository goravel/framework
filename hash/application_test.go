package hash

import (
	"github.com/gookit/color"
	"github.com/goravel/framework/support/file"
	"github.com/spf13/cast"
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
		color.Redln("No mail tests run, need create .env based on .env.example, then initialize it")
		return
	}
	initConfig()
	facades.Hash = NewApplication()
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestMakeHash() {
	if !file.Exists("../.env") {
		color.Redln("No hash tests run, need create .env based on .env.example, then initialize it")
		return
	}
	initConfig()
	s.NotEmpty(facades.Hash.Make("password"))
}

func (s *ApplicationTestSuite) TestCheckHash() {
	if !file.Exists("../.env") {
		color.Redln("No hash tests run, need create .env based on .env.example, then initialize it")
		return
	}
	initConfig()
	hash := facades.Hash.Make("password")
	s.True(facades.Hash.Check("password", hash))
}

func (s *ApplicationTestSuite) TestNeedsRehash() {
	if !file.Exists("../.env") {
		color.Redln("No hash tests run, need create .env based on .env.example, then initialize it")
		return
	}
	initConfig()
	hash := facades.Hash.Make("password")
	s.False(facades.Hash.NeedsRehash(hash))
}

func initConfig() {
	application := config.NewApplication("../.env")
	application.Add("app", map[string]any{
		"name": "goravel",
	})
	application.Add("hashing", map[string]any{
		"driver": "argon2id",
		"bcrypt": map[string]any{
			"cost": cast.ToInt(application.Env("HASH_BCRYPT_COST", 10)),
		},
		"argon2id": map[string]any{
			"memory":  cast.ToUint32(application.Env("HASH_ARGON2ID_MEMORY", 65536)),
			"time":    cast.ToUint32(application.Env("HASH_ARGON2ID_TIME", 4)),
			"threads": cast.ToUint8(application.Env("HASH_ARGON2ID_THREADS", 1)),
		},
	})

	facades.Config = application
}
