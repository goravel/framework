package foundation

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ApplicationTestSuite struct {
	suite.Suite
	app *Application
}

func TestApplicationTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {
	s.app = &Application{
		publishes:     make(map[string]map[string]string),
		publishGroups: make(map[string]map[string]string),
	}
}

func (s *ApplicationTestSuite) TestConfigPath() {
	s.Equal("config/goravel.go", s.app.ConfigPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestDatabasePath() {
	s.Equal("database/goravel.go", s.app.DatabasePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestPublicPath() {
	s.Equal("public/goravel.go", s.app.PublicPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestPublishes() {
	s.app.Publishes("github.com/goravel/sms", map[string]string{
		"config.go": "config.go",
	})
	s.Equal(1, len(s.app.publishes["github.com/goravel/sms"]))
	s.Equal(0, len(s.app.publishGroups))

	s.app.Publishes("github.com/goravel/sms", map[string]string{
		"config.go":  "config1.go",
		"config1.go": "config1.go",
	}, "public", "private")
	s.Equal(2, len(s.app.publishes["github.com/goravel/sms"]))
	s.Equal("config1.go", s.app.publishes["github.com/goravel/sms"]["config.go"])
	s.Equal(2, len(s.app.publishGroups["public"]))
	s.Equal("config1.go", s.app.publishGroups["public"]["config.go"])
	s.Equal(2, len(s.app.publishGroups["private"]))
}

func (s *ApplicationTestSuite) TestAddPublishGroup() {
	s.app.addPublishGroup("public", map[string]string{
		"config.go": "config.go",
	})
	s.Equal(1, len(s.app.publishGroups["public"]))

	s.app.addPublishGroup("public", map[string]string{
		"config.go":  "config1.go",
		"config1.go": "config1.go",
	})
	s.Equal(2, len(s.app.publishGroups["public"]))
	s.Equal("config1.go", s.app.publishGroups["public"]["config.go"])
}
