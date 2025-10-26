package foundation

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support"
)

type ApplicationTestSuite struct {
	suite.Suite
	app    *Application
	cancel context.CancelFunc
}

func TestApplicationTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	s.app = &Application{
		Container:     NewContainer(),
		ctx:           ctx,
		publishes:     make(map[string]map[string]string),
		publishGroups: make(map[string]map[string]string),
	}
	s.cancel = cancel

	App = s.app
}

func (s *ApplicationTestSuite) TestPath() {
	s.Equal(filepath.Join(support.RootPath, "app", "goravel.go"), s.app.Path("goravel.go"))
}

func (s *ApplicationTestSuite) TestBasePath() {
	s.Equal(filepath.Join(support.RootPath, "goravel.go"), s.app.BasePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestConfigPath() {
	s.Equal(filepath.Join(support.RootPath, "config", "goravel.go"), s.app.ConfigPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestDatabasePath() {
	s.Equal(filepath.Join(support.RootPath, "database", "goravel.go"), s.app.DatabasePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestStoragePath() {
	s.Equal(filepath.Join(support.RootPath, "storage", "goravel.go"), s.app.StoragePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestResourcePath() {
	s.Equal(filepath.Join(support.RootPath, "resources", "goravel.go"), s.app.ResourcePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestLangPath() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("app.lang_path", "lang").Return("test").Once()

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	s.Equal(filepath.Join(support.RootPath, "test", "goravel.go"), s.app.LangPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestPublicPath() {
	s.Equal(filepath.Join(support.RootPath, "public", "goravel.go"), s.app.PublicPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestExecutablePath() {
	path, err := os.Getwd()
	s.NoError(err)

	executable := s.app.ExecutablePath()
	s.NotEmpty(executable)
	executable2 := s.app.ExecutablePath("test")
	s.Equal(filepath.Join(path, "test"), executable2)
	executable3 := s.app.ExecutablePath("test", "test2/test3")
	s.Equal(filepath.Join(path, "test", "test2/test3"), executable3)
}

func (s *ApplicationTestSuite) TestRun() {
	oneServiceProvider := mocksfoundation.NewServiceProviderWithRunners(s.T())
	oneRunner := mocksfoundation.NewRunner(s.T())
	oneRunner.EXPECT().ShouldRun().Return(true).Once()
	oneRunner.EXPECT().Run().Return(nil).Once()
	oneRunner.EXPECT().Shutdown().Return(nil).Once()
	oneServiceProvider.EXPECT().Runners(s.app).Return([]foundation.Runner{oneRunner}).Once()

	secondServiceProvider := mocksfoundation.NewServiceProviderWithRunners(s.T())
	secondRunner := mocksfoundation.NewRunner(s.T())
	secondRunner.EXPECT().ShouldRun().Return(true).Once()
	secondRunner.EXPECT().Run().Return(assert.AnError).Once()
	secondServiceProvider.EXPECT().Runners(s.app).Return([]foundation.Runner{secondRunner}).Once()

	thirdRunner := mocksfoundation.NewRunner(s.T())
	thirdRunner.EXPECT().ShouldRun().Return(true).Once()
	thirdRunner.EXPECT().Run().Return(nil).Once()
	thirdRunner.EXPECT().Shutdown().Return(nil).Once()

	fourthRunner := mocksfoundation.NewRunner(s.T())
	fourthRunner.EXPECT().ShouldRun().Return(false).Once()

	fifthServiceProvider := mocksfoundation.NewServiceProviderWithRunners(s.T())
	fifthRunner := mocksfoundation.NewRunner(s.T())
	fifthRunner.EXPECT().ShouldRun().Return(false).Once()
	fifthServiceProvider.EXPECT().Runners(s.app).Return([]foundation.Runner{fifthRunner}).Once()

	mockRepo := mocksfoundation.NewProviderRepository(s.T())
	s.app.providerRepository = mockRepo

	mockRepo.EXPECT().GetBooted().Return([]foundation.ServiceProvider{
		oneServiceProvider,
		secondServiceProvider,
		fifthServiceProvider,
	}).Once()

	s.app.Run(thirdRunner, fourthRunner)

	time.Sleep(100 * time.Millisecond) // Wait for goroutines to start

	s.cancel()

	time.Sleep(100 * time.Millisecond) // Wait for goroutines to end
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
