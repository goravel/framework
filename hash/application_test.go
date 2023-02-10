package hash

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
	facades.Hash = NewApplication()
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestMakeHash() {
	initConfig()
	s.NotEmpty(facades.Hash.Make("password"))
}

func (s *ApplicationTestSuite) TestCheckHash() {
	initConfig()
	hash := facades.Hash.Make("password")
	s.True(facades.Hash.Check("password", hash))
}

func (s *ApplicationTestSuite) TestNeedsRehash() {
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
			"cost": application.Env("HASH_BCRYPT_COST", "10"),
		},
		"argon2id": map[string]any{
			"memory":  application.Env("HASH_ARGON2ID_MEMORY", "65536"),
			"time":    application.Env("HASH_ARGON2ID_TIME", "4"),
			"threads": application.Env("HASH_ARGON2ID_THREADS", "1"),
		},
	})

	facades.Config = application
}
