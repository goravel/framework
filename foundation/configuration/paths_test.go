package configuration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support"
)

type PathsTestSuite struct {
	suite.Suite
	paths         *Paths
	originalPaths support.Paths
}

func TestPathsTestSuite(t *testing.T) {
	suite.Run(t, new(PathsTestSuite))
}

func (s *PathsTestSuite) SetupTest() {
	s.paths = NewPaths()
	// Save original paths before each test
	s.originalPaths = support.Config.Paths
}

func (s *PathsTestSuite) TearDownTest() {
	// Restore original paths after each test
	support.Config.Paths = s.originalPaths
}

func (s *PathsTestSuite) TestNewPaths() {
	paths := NewPaths()
	s.NotNil(paths)
	s.IsType(&Paths{}, paths)
}

func (s *PathsTestSuite) TestApp() {
	result := s.paths.App("custom/app")
	s.Equal("custom/app", support.Config.Paths.App)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestBootstrap() {
	result := s.paths.Bootstrap("custom/bootstrap")
	s.Equal("custom/bootstrap", support.Config.Paths.Bootstrap)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestCommands() {
	result := s.paths.Commands("custom/commands")
	s.Equal("custom/commands", support.Config.Paths.Command)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestConfig() {
	result := s.paths.Config("custom/config")
	s.Equal("custom/config", support.Config.Paths.Config)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestControllers() {
	result := s.paths.Controllers("custom/controllers")
	s.Equal("custom/controllers", support.Config.Paths.Controller)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestDatabase() {
	result := s.paths.Database("custom/database")
	s.Equal("custom/database", support.Config.Paths.Database)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestEvents() {
	result := s.paths.Events("custom/events")
	s.Equal("custom/events", support.Config.Paths.Event)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestFacades() {
	result := s.paths.Facades("custom/facades")
	s.Equal("custom/facades", support.Config.Paths.Facades)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestFactories() {
	result := s.paths.Factories("custom/factories")
	s.Equal("custom/factories", support.Config.Paths.Factory)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestFilters() {
	result := s.paths.Filters("custom/filters")
	s.Equal("custom/filters", support.Config.Paths.Filter)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestJobs() {
	result := s.paths.Jobs("custom/jobs")
	s.Equal("custom/jobs", support.Config.Paths.Job)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestLang() {
	result := s.paths.Lang("custom/lang")
	s.Equal("custom/lang", support.Config.Paths.Lang)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestListeners() {
	result := s.paths.Listeners("custom/listeners")
	s.Equal("custom/listeners", support.Config.Paths.Listener)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestMails() {
	result := s.paths.Mails("custom/mails")
	s.Equal("custom/mails", support.Config.Paths.Mail)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestMiddleware() {
	result := s.paths.Middleware("custom/middleware")
	s.Equal("custom/middleware", support.Config.Paths.Middleware)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestMigrations() {
	result := s.paths.Migrations("custom/migrations")
	s.Equal("custom/migrations", support.Config.Paths.Migration)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestModels() {
	result := s.paths.Models("custom/models")
	s.Equal("custom/models", support.Config.Paths.Model)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestObservers() {
	result := s.paths.Observers("custom/observers")
	s.Equal("custom/observers", support.Config.Paths.Observer)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestPackages() {
	result := s.paths.Packages("custom/packages")
	s.Equal("custom/packages", support.Config.Paths.Package)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestPolicies() {
	result := s.paths.Policies("custom/policies")
	s.Equal("custom/policies", support.Config.Paths.Policy)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestProviders() {
	result := s.paths.Providers("custom/providers")
	s.Equal("custom/providers", support.Config.Paths.Provider)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestPublic() {
	result := s.paths.Public("custom/public")
	s.Equal("custom/public", support.Config.Paths.Public)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestRequests() {
	result := s.paths.Requests("custom/requests")
	s.Equal("custom/requests", support.Config.Paths.Request)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestResources() {
	result := s.paths.Resources("custom/resources")
	s.Equal("custom/resources", support.Config.Paths.Resources)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestRules() {
	result := s.paths.Rules("custom/rules")
	s.Equal("custom/rules", support.Config.Paths.Rule)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestSeeders() {
	result := s.paths.Seeders("custom/seeders")
	s.Equal("custom/seeders", support.Config.Paths.Seeder)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestStorage() {
	result := s.paths.Storage("custom/storage")
	s.Equal("custom/storage", support.Config.Paths.Storage)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestTests() {
	result := s.paths.Tests("custom/tests")
	s.Equal("custom/tests", support.Config.Paths.Test)
	s.Equal(s.paths, result)
}

func (s *PathsTestSuite) TestChaining() {
	result := s.paths.
		App("app").
		Config("config").
		Controllers("controllers").
		Models("models")

	s.Equal("app", support.Config.Paths.App)
	s.Equal("config", support.Config.Paths.Config)
	s.Equal("controllers", support.Config.Paths.Controller)
	s.Equal("models", support.Config.Paths.Model)
	s.Equal(s.paths, result)
}
