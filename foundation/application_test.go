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

	"github.com/goravel/framework/contracts/foundation"
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

func (s *ApplicationTestSuite) TestModelPath() {
	s.Equal(filepath.Join(support.RootPath, "app", "models", "goravel.go"), s.app.ModelPath("goravel.go"))
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
	// Create a fresh container for this subtest
	app := &Application{
		Container:     NewContainer(),
		ctx:           s.app.ctx,
		publishes:     make(map[string]map[string]string),
		publishGroups: make(map[string]map[string]string),
	}

	s.Equal(filepath.Join(support.RootPath, support.Config.Paths.Lang, "goravel.go"), app.LangPath("goravel.go"))
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

func (s *ApplicationTestSuite) TestStart() {
	tests := []struct {
		name  string
		setup func() foundation.Runner
	}{
		{
			name: "happy path",
			setup: func() foundation.Runner {
				runner := mocksfoundation.NewRunner(s.T())
				runner.EXPECT().ShouldRun().Return(true).Once()
				runner.EXPECT().Run().Return(nil).Once()
				runner.EXPECT().Shutdown().Return(nil).Once()

				return runner
			},
		},
		{
			name: "failed to run",
			setup: func() foundation.Runner {
				runner := mocksfoundation.NewRunner(s.T())
				runner.EXPECT().ShouldRun().Return(true).Once()
				runner.EXPECT().Run().Return(assert.AnError).Once()

				return runner
			},
		},
		{
			name: "should not be run",
			setup: func() foundation.Runner {
				runner := mocksfoundation.NewRunner(s.T())
				runner.EXPECT().ShouldRun().Return(false).Once()

				return runner
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			runner := tt.setup()
			serviceProvider := mocksfoundation.NewServiceProviderWithRunners(s.T())
			serviceProvider.EXPECT().Runners(s.app).Return([]foundation.Runner{runner}).Once()

			mockRepo := mocksfoundation.NewProviderRepository(s.T())
			mockRepo.EXPECT().GetBooted().Return([]foundation.ServiceProvider{
				serviceProvider,
			}).Once()

			s.app.providerRepository = mockRepo
			s.app.runnerNames = nil
			app := s.app.Start()

			go func() {
				time.Sleep(100 * time.Millisecond) // Wait for goroutines to start
				s.cancel()
			}()

			app.Wait()
		})
	}
}

func (s *ApplicationTestSuite) TestStart_Complex() {
	s.Run("With additional runner", func() {
		s.SetupTest()
		runner := mocksfoundation.NewRunner(s.T())
		runner.EXPECT().ShouldRun().Return(true).Once()
		runner.EXPECT().Run().Return(nil).Once()
		runner.EXPECT().Shutdown().Return(nil).Once()

		mockRepo := mocksfoundation.NewProviderRepository(s.T())
		mockRepo.EXPECT().GetBooted().Return(nil).Once()

		s.app.providerRepository = mockRepo
		app := s.app.Start(runner)

		go func() {
			time.Sleep(200 * time.Millisecond) // Wait for goroutines to start
			s.cancel()
		}()

		app.Wait()
	})

	s.Run("With duplicated runners", func() {
		s.SetupTest()
		runner := mocksfoundation.NewRunner(s.T())
		runner.EXPECT().ShouldRun().Return(true).Once()
		runner.EXPECT().Run().Return(nil).Once()
		runner.EXPECT().Shutdown().Return(nil).Once()

		serviceProvider := mocksfoundation.NewServiceProviderWithRunners(s.T())
		serviceProvider.EXPECT().Runners(s.app).Return([]foundation.Runner{runner}).Once()

		mockRepo := mocksfoundation.NewProviderRepository(s.T())
		mockRepo.EXPECT().GetBooted().Return([]foundation.ServiceProvider{
			serviceProvider,
		}).Once()

		s.app.providerRepository = mockRepo
		app := s.app.Start(runner)

		go func() {
			time.Sleep(200 * time.Millisecond) // Wait for goroutines to start
			s.cancel()
		}()

		app.Wait()
	})

	s.Run("Call Start several times", func() {
		s.SetupTest()
		runner := mocksfoundation.NewRunner(s.T())
		runner.EXPECT().ShouldRun().Return(true).Once()
		runner.EXPECT().Run().Return(nil).Once()
		runner.EXPECT().Shutdown().Return(nil).Once()

		mockRepo := mocksfoundation.NewProviderRepository(s.T())
		mockRepo.EXPECT().GetBooted().Return(nil).Twice()

		s.app.providerRepository = mockRepo
		app := s.app.Start(runner)
		app = app.Start(runner)

		go func() {
			time.Sleep(200 * time.Millisecond) // Wait for goroutines to start
			s.cancel()
		}()

		app.Wait()
	})
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
