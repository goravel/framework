package foundation

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/contracts/validation"
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
		cancel:        cancel,
		publishes:     make(map[string]map[string]string),
		publishGroups: make(map[string]map[string]string),
	}
	s.cancel = cancel

	App = s.app
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

func (s *ApplicationTestSuite) TestBasePath() {
	s.Equal(filepath.Join(support.RootPath, "goravel.go"), s.app.BasePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestConfigPath() {
	s.Equal(filepath.Join(support.RootPath, "config", "goravel.go"), s.app.ConfigPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestConfigureCallback() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "with callback",
			setup: func() {
				called := false
				builder := NewApplicationBuilder(s.app)
				builder.callback = func() {
					called = true
				}
				s.app.builder = builder
				s.app.configureCallback()
				s.True(called)
			},
		},
		{
			name: "without callback",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.callback = nil
				s.app.builder = builder
				s.app.configureCallback() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureCommands() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without commands function",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.commands = nil
				s.app.builder = builder
				s.app.configureCommands() // Should not panic
			},
		},
		{
			name: "with empty commands",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.commands = func() []console.Command {
					return []console.Command{}
				}
				s.app.builder = builder
				s.app.configureCommands() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureCustomConfig() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "with config",
			setup: func() {
				called := false
				builder := NewApplicationBuilder(s.app)
				builder.config = func() {
					called = true
				}
				s.app.builder = builder
				s.app.configureCustomConfig()
				s.True(called)
			},
		},
		{
			name: "without config",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.config = nil
				s.app.builder = builder
				s.app.configureCustomConfig() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureEventListeners() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without event function",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.eventToListeners = nil
				s.app.builder = builder
				s.app.configureEventListeners() // Should not panic
			},
		},
		{
			name: "with empty events",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.eventToListeners = func() map[event.Event][]event.Listener {
					return map[event.Event][]event.Listener{}
				}
				s.app.builder = builder
				s.app.configureEventListeners() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureGrpc() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without grpc configurations",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.grpcClientInterceptors = nil
				builder.grpcServerInterceptors = nil
				builder.grpcClientStatsHandlers = nil
				builder.grpcServerStatsHandlers = nil
				s.app.builder = builder
				s.app.configureGrpc() // Should not panic
			},
		},
		{
			name: "with empty grpc configurations",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.grpcClientInterceptors = func() map[string][]grpc.UnaryClientInterceptor {
					return map[string][]grpc.UnaryClientInterceptor{}
				}
				builder.grpcServerInterceptors = func() []grpc.UnaryServerInterceptor {
					return []grpc.UnaryServerInterceptor{}
				}
				builder.grpcClientStatsHandlers = func() map[string][]stats.Handler {
					return map[string][]stats.Handler{}
				}
				builder.grpcServerStatsHandlers = func() []stats.Handler {
					return []stats.Handler{}
				}
				s.app.builder = builder
				s.app.configureGrpc() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureJobs() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without jobs function",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.jobs = nil
				s.app.builder = builder
				s.app.configureJobs() // Should not panic
			},
		},
		{
			name: "with empty jobs",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.jobs = func() []queue.Job {
					return []queue.Job{}
				}
				s.app.builder = builder
				s.app.configureJobs() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureMiddleware() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without middleware function",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.middleware = nil
				s.app.builder = builder
				s.app.configureMiddleware() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureMigrations() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without migrations function",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.migrations = nil
				s.app.builder = builder
				s.app.configureMigrations() // Should not panic
			},
		},
		{
			name: "with empty migrations",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.migrations = func() []schema.Migration {
					return []schema.Migration{}
				}
				s.app.builder = builder
				s.app.configureMigrations() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigurePaths() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without paths function",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.paths = nil
				s.app.builder = builder
				s.app.configurePaths() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureRoutes() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "with routes",
			setup: func() {
				called := false
				builder := NewApplicationBuilder(s.app)
				builder.routes = func() {
					called = true
				}
				s.app.builder = builder
				s.app.configureRoutes()
				s.True(called)
			},
		},
		{
			name: "without routes",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.routes = nil
				s.app.builder = builder
				s.app.configureRoutes() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureRunners() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without runners",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.runners = nil
				s.app.builder = builder
				s.app.bootedRunners = nil
				s.app.runnersToRun = nil

				mockRepo := mocksfoundation.NewProviderRepository(s.T())
				mockRepo.EXPECT().GetBooted().Return([]foundation.ServiceProvider{}).Once()
				s.app.providerRepository = mockRepo

				s.app.configureRunners() // Should not panic
				s.Equal(0, len(s.app.runnersToRun))
			},
		},
		{
			name: "with runners from builder",
			setup: func() {
				runner := mocksfoundation.NewRunner(s.T())
				runner.EXPECT().Signature().Return("test-runner").Once()
				runner.EXPECT().ShouldRun().Return(true).Once()

				builder := NewApplicationBuilder(s.app)
				builder.runners = func() []foundation.Runner {
					return []foundation.Runner{runner}
				}
				s.app.builder = builder
				s.app.bootedRunners = nil
				s.app.runnersToRun = nil

				mockRepo := mocksfoundation.NewProviderRepository(s.T())
				mockRepo.EXPECT().GetBooted().Return([]foundation.ServiceProvider{}).Once()
				s.app.providerRepository = mockRepo

				s.app.configureRunners()
				s.Equal(1, len(s.app.runnersToRun))
				s.Equal("test-runner", s.app.runnersToRun[0].signature)
			},
		},
		{
			name: "with runners that should not run",
			setup: func() {
				runner := mocksfoundation.NewRunner(s.T())
				runner.EXPECT().Signature().Return("test-runner").Once()
				runner.EXPECT().ShouldRun().Return(false).Once()

				builder := NewApplicationBuilder(s.app)
				builder.runners = func() []foundation.Runner {
					return []foundation.Runner{runner}
				}
				s.app.builder = builder
				s.app.bootedRunners = nil
				s.app.runnersToRun = nil

				mockRepo := mocksfoundation.NewProviderRepository(s.T())
				mockRepo.EXPECT().GetBooted().Return([]foundation.ServiceProvider{}).Once()
				s.app.providerRepository = mockRepo

				s.app.configureRunners()
				s.Equal(0, len(s.app.runnersToRun))
			},
		},
		{
			name: "with runners from service provider",
			setup: func() {
				runner := mocksfoundation.NewRunner(s.T())
				runner.EXPECT().Signature().Return("provider-runner").Once()
				runner.EXPECT().ShouldRun().Return(true).Once()

				serviceProvider := mocksfoundation.NewServiceProviderWithRunners(s.T())
				serviceProvider.EXPECT().Runners(s.app).Return([]foundation.Runner{runner}).Once()

				builder := NewApplicationBuilder(s.app)
				builder.runners = nil
				s.app.builder = builder
				s.app.bootedRunners = nil
				s.app.runnersToRun = nil

				mockRepo := mocksfoundation.NewProviderRepository(s.T())
				mockRepo.EXPECT().GetBooted().Return([]foundation.ServiceProvider{serviceProvider}).Once()
				s.app.providerRepository = mockRepo

				s.app.configureRunners()
				s.Equal(1, len(s.app.runnersToRun))
				s.Equal("provider-runner", s.app.runnersToRun[0].signature)
			},
		},
		{
			name: "skip already booted runners",
			setup: func() {
				runner := mocksfoundation.NewRunner(s.T())
				runner.EXPECT().Signature().Return("test-runner").Once()

				builder := NewApplicationBuilder(s.app)
				builder.runners = func() []foundation.Runner {
					return []foundation.Runner{runner}
				}
				s.app.builder = builder
				s.app.bootedRunners = []string{"test-runner"}
				s.app.runnersToRun = nil

				mockRepo := mocksfoundation.NewProviderRepository(s.T())
				mockRepo.EXPECT().GetBooted().Return([]foundation.ServiceProvider{}).Once()
				s.app.providerRepository = mockRepo

				s.app.configureRunners()
				s.Equal(0, len(s.app.runnersToRun))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureSchedule() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without schedule function",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.schedule = nil
				s.app.builder = builder
				s.app.configureSchedule() // Should not panic
			},
		},
		{
			name: "with empty schedule",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.schedule = func() []schedule.Event {
					return []schedule.Event{}
				}
				s.app.builder = builder
				s.app.configureSchedule() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureSeeders() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without seeders function",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.seeders = nil
				s.app.builder = builder
				s.app.configureSeeders() // Should not panic
			},
		},
		{
			name: "with empty seeders",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.seeders = func() []seeder.Seeder {
					return []seeder.Seeder{}
				}
				s.app.builder = builder
				s.app.configureSeeders() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureServiceProviders() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without service providers function",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.configuredServiceProviders = nil
				s.app.builder = builder

				s.app.configureServiceProviders() // Should not panic
			},
		},
		{
			name: "with empty service providers",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.configuredServiceProviders = func() []foundation.ServiceProvider {
					return []foundation.ServiceProvider{}
				}
				s.app.builder = builder

				s.app.configureServiceProviders() // Should not panic
			},
		},
		{
			name: "with service providers",
			setup: func() {
				serviceProvider := mocksfoundation.NewServiceProvider(s.T())

				builder := NewApplicationBuilder(s.app)
				builder.configuredServiceProviders = func() []foundation.ServiceProvider {
					return []foundation.ServiceProvider{serviceProvider}
				}
				s.app.builder = builder

				mockRepo := mocksfoundation.NewProviderRepository(s.T())
				mockRepo.EXPECT().Add([]foundation.ServiceProvider{serviceProvider}).Once()
				s.app.providerRepository = mockRepo

				s.app.configureServiceProviders()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestConfigureValidation() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "without validation configurations",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.rules = nil
				builder.filters = nil
				s.app.builder = builder
				s.app.configureValidation() // Should not panic
			},
		},
		{
			name: "with empty validation configurations",
			setup: func() {
				builder := NewApplicationBuilder(s.app)
				builder.rules = func() []validation.Rule {
					return []validation.Rule{}
				}
				builder.filters = func() []validation.Filter {
					return []validation.Filter{}
				}
				s.app.builder = builder
				s.app.configureValidation() // Should not panic
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
		})
	}
}

func (s *ApplicationTestSuite) TestDatabasePath() {
	s.Equal(filepath.Join(support.RootPath, "database", "goravel.go"), s.app.DatabasePath("goravel.go"))
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

func (s *ApplicationTestSuite) TestModelPath() {
	s.Equal(filepath.Join(support.RootPath, "app", "models", "goravel.go"), s.app.ModelPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestPath() {
	s.Equal(filepath.Join(support.RootPath, "app", "goravel.go"), s.app.Path("goravel.go"))
}

func (s *ApplicationTestSuite) TestPublicPath() {
	s.Equal(filepath.Join(support.RootPath, "public", "goravel.go"), s.app.PublicPath("goravel.go"))
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

func (s *ApplicationTestSuite) TestResourcePath() {
	s.Equal(filepath.Join(support.RootPath, "resources", "goravel.go"), s.app.ResourcePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestStart() {
	tests := []struct {
		name        string
		setup       func() []foundation.Runner
		expectPanic bool
	}{
		{
			name: "happy path",
			setup: func() []foundation.Runner {
				runner := mocksfoundation.NewRunner(s.T())
				runner.EXPECT().Signature().Return("test-runner").Once()
				runner.EXPECT().ShouldRun().Return(true).Once()
				runner.EXPECT().Run().Return(nil).Once()
				runner.EXPECT().Shutdown().Return(nil).Once()

				return []foundation.Runner{runner}
			},
			expectPanic: false,
		},
		{
			name: "failed to run",
			setup: func() []foundation.Runner {
				runner := mocksfoundation.NewRunner(s.T())
				runner.EXPECT().Signature().Return("test-runner").Once()
				runner.EXPECT().ShouldRun().Return(true).Once()
				runner.EXPECT().Run().Return(assert.AnError).Once()

				return []foundation.Runner{runner}
			},
			expectPanic: true,
		},
		{
			name: "should not be run",
			setup: func() []foundation.Runner {
				runner := mocksfoundation.NewRunner(s.T())
				runner.EXPECT().Signature().Return("test-runner").Once()
				runner.EXPECT().ShouldRun().Return(false).Once()

				return []foundation.Runner{runner}
			},
			expectPanic: false,
		},
		{
			name: "multiple runners all successful",
			setup: func() []foundation.Runner {
				runner1 := mocksfoundation.NewRunner(s.T())
				runner1.EXPECT().Signature().Return("test-runner-1").Once()
				runner1.EXPECT().ShouldRun().Return(true).Once()
				runner1.EXPECT().Run().Return(nil).Once()
				runner1.EXPECT().Shutdown().Return(nil).Once()

				runner2 := mocksfoundation.NewRunner(s.T())
				runner2.EXPECT().Signature().Return("test-runner-2").Once()
				runner2.EXPECT().ShouldRun().Return(true).Once()
				runner2.EXPECT().Run().Return(nil).Once()
				runner2.EXPECT().Shutdown().Return(nil).Once()

				runner3 := mocksfoundation.NewRunner(s.T())
				runner3.EXPECT().Signature().Return("test-runner-3").Once()
				runner3.EXPECT().ShouldRun().Return(true).Once()
				runner3.EXPECT().Run().Return(nil).Once()
				runner3.EXPECT().Shutdown().Return(nil).Once()

				return []foundation.Runner{runner1, runner2, runner3}
			},
			expectPanic: false,
		},
		{
			name: "multiple runners with one failure",
			setup: func() []foundation.Runner {
				runner1 := mocksfoundation.NewRunner(s.T())
				runner1.EXPECT().Signature().Return("test-runner-1").Once()
				runner1.EXPECT().ShouldRun().Return(true).Once()
				runner1.EXPECT().Run().Return(nil).Once()
				runner1.EXPECT().Shutdown().Return(nil).Maybe()

				runner2 := mocksfoundation.NewRunner(s.T())
				runner2.EXPECT().Signature().Return("test-runner-2").Once()
				runner2.EXPECT().ShouldRun().Return(true).Once()
				runner2.EXPECT().Run().Return(assert.AnError).Once()

				return []foundation.Runner{runner1, runner2}
			},
			expectPanic: true,
		},
		{
			name: "multiple runners with mixed should run",
			setup: func() []foundation.Runner {
				runner1 := mocksfoundation.NewRunner(s.T())
				runner1.EXPECT().Signature().Return("test-runner-1").Once()
				runner1.EXPECT().ShouldRun().Return(true).Once()
				runner1.EXPECT().Run().Return(nil).Once()
				runner1.EXPECT().Shutdown().Return(nil).Once()

				runner2 := mocksfoundation.NewRunner(s.T())
				runner2.EXPECT().Signature().Return("test-runner-2").Once()
				runner2.EXPECT().ShouldRun().Return(false).Once()

				runner3 := mocksfoundation.NewRunner(s.T())
				runner3.EXPECT().Signature().Return("test-runner-3").Once()
				runner3.EXPECT().ShouldRun().Return(true).Once()
				runner3.EXPECT().Run().Return(nil).Once()
				runner3.EXPECT().Shutdown().Return(nil).Once()

				return []foundation.Runner{runner1, runner2, runner3}
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			runners := tt.setup()
			serviceProvider := mocksfoundation.NewServiceProviderWithRunners(s.T())
			serviceProvider.EXPECT().Runners(s.app).Return(runners).Once()

			mockRepo := mocksfoundation.NewProviderRepository(s.T())
			mockRepo.EXPECT().GetBooted().Return([]foundation.ServiceProvider{
				serviceProvider,
			}).Once()

			s.app.builder = NewApplicationBuilder(s.app)
			s.app.providerRepository = mockRepo
			s.app.bootedRunners = nil
			s.app.runnersToRun = nil
			s.app.configureRunners()

			go func() {
				time.Sleep(100 * time.Millisecond) // Wait for goroutines to start
				s.cancel()
			}()

			if tt.expectPanic {
				s.Panics(func() {
					s.app.Start()
				})
			} else {
				s.NotPanics(func() {
					s.app.Start()
				})
			}
		})
	}
}

func (s *ApplicationTestSuite) TestStoragePath() {
	s.Equal(filepath.Join(support.RootPath, "storage", "goravel.go"), s.app.StoragePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestShutdown() {
	tests := []struct {
		name        string
		setup       func() []foundation.Runner
		expectError bool
	}{
		{
			name: "shutdown with no runners",
			setup: func() []foundation.Runner {
				return []foundation.Runner{}
			},
			expectError: false,
		},
		{
			name: "shutdown with running runners",
			setup: func() []foundation.Runner {
				runner := mocksfoundation.NewRunner(s.T())
				runner.EXPECT().Signature().Return("test-runner").Once()
				runner.EXPECT().ShouldRun().Return(true).Once()
				runner.EXPECT().Run().Return(nil).Once()
				runner.EXPECT().Shutdown().Return(nil).Once()

				return []foundation.Runner{runner}
			},
			expectError: false,
		},
		{
			name: "shutdown with multiple runners",
			setup: func() []foundation.Runner {
				runner1 := mocksfoundation.NewRunner(s.T())
				runner1.EXPECT().Signature().Return("test-runner-1").Once()
				runner1.EXPECT().ShouldRun().Return(true).Once()
				runner1.EXPECT().Run().Return(nil).Once()
				runner1.EXPECT().Shutdown().Return(nil).Once()

				runner2 := mocksfoundation.NewRunner(s.T())
				runner2.EXPECT().Signature().Return("test-runner-2").Once()
				runner2.EXPECT().ShouldRun().Return(true).Once()
				runner2.EXPECT().Run().Return(nil).Once()
				runner2.EXPECT().Shutdown().Return(nil).Once()

				return []foundation.Runner{runner1, runner2}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			runners := tt.setup()

			if len(runners) > 0 {
				serviceProvider := mocksfoundation.NewServiceProviderWithRunners(s.T())
				serviceProvider.EXPECT().Runners(s.app).Return(runners).Once()

				mockRepo := mocksfoundation.NewProviderRepository(s.T())
				mockRepo.EXPECT().GetBooted().Return([]foundation.ServiceProvider{
					serviceProvider,
				}).Once()

				s.app.builder = NewApplicationBuilder(s.app)
				s.app.providerRepository = mockRepo
				s.app.bootedRunners = nil
				s.app.configureRunners()

				// Start runners in the background
				go s.app.Start()

				// Wait a moment for runners to start
				time.Sleep(50 * time.Millisecond)
			}

			// Shutdown the application
			err := s.app.Shutdown()

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)

				// Verify all runners have stopped
				for _, runner := range s.app.runnersToRun {
					s.False(runner.running.Load())
				}
			}
		})
	}
}

// TestRunnerWithInfoRaceFreedom tests that RunnerWithInfo operations are race-free
func TestRunnerWithInfoRaceFreedom(t *testing.T) {
	runner := &RunnerWithInfo{
		signature: "test-runner",
	}

	var wg sync.WaitGroup

	// Simulate multiple goroutines accessing running field
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				runner.running.Store(true)
				time.Sleep(1 * time.Millisecond)
				_ = runner.running.Load()
				runner.running.Store(false)
			}
		}()
	}

	// Simulate multiple goroutines calling doneOnce
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.doneOnce.Do(func() {
				// This should only be called once
			})
		}()
	}

	wg.Wait()

	// Verify final state
	assert.False(t, runner.running.Load())
}

// TestDoneOnceGuarantee verifies that runnerWg.Done() is called exactly once
func TestDoneOnceGuarantee(t *testing.T) {
	runner := &RunnerWithInfo{
		signature: "test-runner",
	}

	var wg sync.WaitGroup
	counter := 0
	var mu sync.Mutex

	// Simulate multiple goroutines trying to call Done
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.doneOnce.Do(func() {
				mu.Lock()
				counter++
				mu.Unlock()
			})
		}()
	}

	wg.Wait()

	// Verify counter was incremented exactly once
	assert.Equal(t, 1, counter, "doneOnce should ensure the function is called exactly once")
}
