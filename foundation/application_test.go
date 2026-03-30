package foundation

import (
	"context"
	"os"
	"path/filepath"
	"sync"
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
	ctx, cancel := context.WithCancel(context.Background())

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
	s.Contains(s.app.BasePath("goravel.go"), filepath.Join("framework", "goravel.go"))
}

func (s *ApplicationTestSuite) TestConfigPath() {
	s.Contains(s.app.ConfigPath("goravel.go"), filepath.Join("framework", "config", "goravel.go"))
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
	s.Contains(s.app.DatabasePath("goravel.go"), filepath.Join("framework", "database", "goravel.go"))
}

func (s *ApplicationTestSuite) TestExecutablePath() {
	executable := s.app.ExecutablePath()
	s.NotEmpty(executable)
	executable2 := s.app.ExecutablePath("test")
	s.Contains(executable2, filepath.Join("framework", "test"))
	executable3 := s.app.ExecutablePath("test", "test2/test3")
	s.Contains(executable3, filepath.Join("framework", "test", "test2", "test3"))
}

func (s *ApplicationTestSuite) TestLangPath() {
	// Create a fresh container for this subtest
	app := &Application{
		Container:     NewContainer(),
		ctx:           s.app.ctx,
		publishes:     make(map[string]map[string]string),
		publishGroups: make(map[string]map[string]string),
	}

	s.Contains(app.LangPath("goravel.go"), filepath.Join("framework", "lang", "goravel.go"))
}

func (s *ApplicationTestSuite) TestModelPath() {
	s.Contains(s.app.ModelPath("goravel.go"), filepath.Join("framework", "app", "models", "goravel.go"))
}

func (s *ApplicationTestSuite) TestPath() {
	s.Contains(s.app.Path("goravel.go"), filepath.Join("framework", "app", "goravel.go"))
}

func (s *ApplicationTestSuite) TestPublicPath() {
	s.Contains(s.app.PublicPath("goravel.go"), filepath.Join("framework", "public", "goravel.go"))
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
	s.Contains(s.app.ResourcePath("goravel.go"), filepath.Join("framework", "resources", "goravel.go"))
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
				runner1.EXPECT().Shutdown().Return(nil).Once()

				runner2 := mocksfoundation.NewRunner(s.T())
				runner2.EXPECT().Signature().Return("test-runner-2").Once()
				runner2.EXPECT().ShouldRun().Return(true).Once()
				runner2.EXPECT().Run().RunAndReturn(func() error {
					time.Sleep(1 * time.Second)
					return assert.AnError
				}).Once()

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

			if tt.expectPanic {
				// s.Panics does not work well with goroutines, so we capture the panic manually
				panicChan := make(chan interface{}, 1)
				go func() {
					defer func() {
						if r := recover(); r != nil {
							panicChan <- r
						}
					}()
					s.app.Start()
				}()

				select {
				case <-panicChan:
					// Panic occurred as expected
				case <-time.After(5 * time.Second):
					s.Fail("expected panic but none occurred")
				}
			} else {
				// Only trigger cancel for non-panic cases
				// For panic cases, the error handling will call cancel automatically
				cancel := s.cancel
				go func() {
					time.Sleep(2 * time.Second) // Wait for goroutines to start
					cancel()
				}()

				s.NotPanics(func() {
					s.app.Start()
				})
			}
		})
	}
}

func (s *ApplicationTestSuite) TestStoragePath() {
	s.Contains(s.app.StoragePath("goravel.go"), filepath.Join("framework", "storage", "goravel.go"))
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

			var err error

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
				startDone := make(chan struct{})
				go func() {
					defer close(startDone)
					s.app.Start()
				}()

				// Wait a moment for runners to start
				time.Sleep(50 * time.Millisecond)

				// Shutdown the application
				err = s.app.Shutdown()

				// Wait for Start() to complete before mock cleanup
				<-startDone
			} else {
				// Shutdown the application
				err = s.app.Shutdown()
			}

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

func backupSetEnvGlobalState(t *testing.T) {
	originalArgs := append([]string(nil), os.Args...)
	originalRuntimeMode := support.RuntimeMode
	originalDontVerifyAppKey := support.DontVerifyAppKey
	originalEnvFilePath := support.EnvFilePath
	originalRelativePath := support.RelativePath

	t.Cleanup(func() {
		os.Args = originalArgs
		support.RuntimeMode = originalRuntimeMode
		support.DontVerifyAppKey = originalDontVerifyAppKey
		support.EnvFilePath = originalEnvFilePath
		support.RelativePath = originalRelativePath
	})
}

func TestGetEnvFilePath_NoPanicOnMissingValue(t *testing.T) {
	backupSetEnvGlobalState(t)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "missing value for --env at end",
			args: []string{"goravel", "--env"},
		},
		{
			name: "missing value for -env at end",
			args: []string{"goravel", "-env"},
		},
		{
			name: "missing value for -e at end",
			args: []string{"goravel", "-e"},
		},
		{
			name: "missing value for --env followed by another flag",
			args: []string{"goravel", "--env", "--ansi"},
		},
		{
			name: "missing value for -env followed by another flag",
			args: []string{"goravel", "-env", "-v"},
		},
		{
			name: "missing value for -e followed by another flag",
			args: []string{"goravel", "-e", "-v"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			assert.NotPanics(t, func() {
				assert.Equal(t, ".env", getEnvFilePath())
			})
		})
	}
}

func TestGetEnvFilePath_ParseValue(t *testing.T) {
	backupSetEnvGlobalState(t)

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "parse value from --env=",
			args:     []string{"goravel", "--env=.env.testing"},
			expected: ".env.testing",
		},
		{
			name:     "parse value from -env=",
			args:     []string{"goravel", "-env=.env.dev"},
			expected: ".env.dev",
		},
		{
			name:     "parse value from -e=",
			args:     []string{"goravel", "-e=.env.staging"},
			expected: ".env.staging",
		},
		{
			name:     "parse split value from --env",
			args:     []string{"goravel", "--env", ".env.prod"},
			expected: ".env.prod",
		},
		{
			name:     "parse split value from -env",
			args:     []string{"goravel", "-env", ".env.qa"},
			expected: ".env.qa",
		},
		{
			name:     "parse split value from -e",
			args:     []string{"goravel", "-e", ".env.local"},
			expected: ".env.local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			assert.Equal(t, tt.expected, getEnvFilePath())
		})
	}
}

func TestSetEnv_DebugBinaryWithoutTestArgs(t *testing.T) {
	backupSetEnvGlobalState(t)

	support.RuntimeMode = ""
	support.DontVerifyAppKey = false
	support.EnvFilePath = ".env"
	support.RelativePath = ""
	os.Args = []string{"/tmp/__debug_bin"}

	setEnv()

	assert.Equal(t, "", support.RuntimeMode)
	assert.False(t, support.DontVerifyAppKey)
	assert.Equal(t, ".env", support.EnvFilePath)
}

func TestSetEnv_DebugBinaryWithTestArgs(t *testing.T) {
	backupSetEnvGlobalState(t)

	support.RuntimeMode = ""
	support.DontVerifyAppKey = false
	support.EnvFilePath = ".env"
	support.RelativePath = ""
	os.Args = []string{"/tmp/__debug_bin", "-test.v=true"}

	setEnv()

	assert.Equal(t, support.RuntimeTest, support.RuntimeMode)
	assert.True(t, support.DontVerifyAppKey)
}
