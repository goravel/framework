package foundation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	contractsevent "github.com/goravel/framework/contracts/event"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsconfiguration "github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/contracts/validation"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
	mocksseeder "github.com/goravel/framework/mocks/database/seeder"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	mocksschedule "github.com/goravel/framework/mocks/schedule"
	mocksvalidation "github.com/goravel/framework/mocks/validation"
	"github.com/goravel/framework/support"
)

type ApplicationBuilderTestSuite struct {
	suite.Suite
	builder       *ApplicationBuilder
	mockApp       *mocksfoundation.Application
	originalPaths support.Paths
}

func TestApplicationBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationBuilderTestSuite))
}

func (s *ApplicationBuilderTestSuite) SetupTest() {
	s.originalPaths = support.Config.Paths
	s.mockApp = mocksfoundation.NewApplication(s.T())
	s.builder = NewApplicationBuilder(s.mockApp)
}

func (s *ApplicationBuilderTestSuite) TearDownTest() {
	support.Config.Paths = s.originalPaths
}

func (s *ApplicationBuilderTestSuite) TestWithConfig() {
	fn := func() {}

	builder := s.builder.WithConfig(fn)

	s.NotNil(builder)
	s.NotNil(s.builder.config)
}

func (s *ApplicationBuilderTestSuite) TestSetup() {
	originalApp := App
	t := s.T()
	t.Cleanup(func() {
		App = originalApp
	})

	App = s.mockApp

	builder := Setup()
	applicationBuilder, ok := builder.(*ApplicationBuilder)
	s.True(ok)
	s.Equal(s.mockApp, applicationBuilder.app)
}

func (s *ApplicationBuilderTestSuite) TestCreate() {
	app := mocksfoundation.NewApplication(s.T())
	builder := NewApplicationBuilder(app)
	app.EXPECT().SetBuilder(builder).Return(app).Once()
	app.EXPECT().Build().Return(app).Once()

	created := builder.Create()

	s.Equal(app, created)
}

func (s *ApplicationBuilderTestSuite) TestWithCommands() {
	command := mocksconsole.NewCommand(s.T())
	builder := s.builder.WithCommands(func() []contractsconsole.Command {
		return []contractsconsole.Command{command}
	})

	s.NotNil(builder)
	s.NotNil(s.builder.commands)
	commands := s.builder.commands()
	s.Len(commands, 1)
	s.Same(command, commands[0])
}

func (s *ApplicationBuilderTestSuite) TestWithEvents() {
	builder := s.builder.WithEvents(func() map[contractsevent.Event][]contractsevent.Listener {
		return nil
	})

	s.NotNil(builder)
	s.NotNil(s.builder.eventToListeners)
}

func (s *ApplicationBuilderTestSuite) TestWithMiddleware() {
	fn := func(middleware contractsconfiguration.Middleware) {}

	builder := s.builder.WithMiddleware(fn)

	s.NotNil(builder)
	s.NotNil(s.builder.middleware)
}

func (s *ApplicationBuilderTestSuite) TestWithPaths() {
	fn := func(paths contractsconfiguration.Paths) {}

	builder := s.builder.WithPaths(fn)

	s.NotNil(builder)
	s.NotNil(s.builder.paths)
}

func (s *ApplicationBuilderTestSuite) TestWithMigrations() {
	mockMigration := mocksschema.NewMigration(s.T())

	builder := s.builder.WithMigrations(func() []schema.Migration {
		return []schema.Migration{mockMigration}
	})

	s.NotNil(builder)
	s.NotNil(s.builder.migrations)
}

func (s *ApplicationBuilderTestSuite) TestWithRouting() {
	fn := func() {}

	builder := s.builder.WithRouting(fn)

	s.NotNil(builder)
	s.NotNil(s.builder.routes)
}

func (s *ApplicationBuilderTestSuite) TestWithProviders() {
	provider := mocksfoundation.NewServiceProvider(s.T())
	builder := s.builder.WithProviders(func() []contractsfoundation.ServiceProvider { return []contractsfoundation.ServiceProvider{provider} })

	s.NotNil(builder)
	s.NotNil(s.builder.configuredServiceProviders)
}

func (s *ApplicationBuilderTestSuite) TestWithSchedule() {
	mockEvent := mocksschedule.NewEvent(s.T())

	builder := s.builder.WithSchedule(func() []schedule.Event {
		return []schedule.Event{mockEvent}
	})

	s.NotNil(builder)
	s.NotNil(s.builder.schedule)
}

func (s *ApplicationBuilderTestSuite) TestWithGrpcClientInterceptors() {
	interceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return nil
	}
	builder := s.builder.WithGrpcClientInterceptors(func() map[string][]grpc.UnaryClientInterceptor {
		return map[string][]grpc.UnaryClientInterceptor{
			"test": {interceptor},
		}
	})

	s.NotNil(builder)
	s.NotNil(s.builder.grpcClientInterceptors)
}

func (s *ApplicationBuilderTestSuite) TestWithGrpcClientStatsHandlers() {
	handler := &mockStatsHandler{}
	builder := s.builder.WithGrpcClientStatsHandlers(func() map[string][]stats.Handler {
		return map[string][]stats.Handler{
			"service-a": {handler},
		}
	})

	s.NotNil(builder)
	s.NotNil(s.builder.grpcClientStatsHandlers)
}

func (s *ApplicationBuilderTestSuite) TestWithGrpcServerInterceptors() {
	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return nil, nil
	}
	builder := s.builder.WithGrpcServerInterceptors(func() []grpc.UnaryServerInterceptor { return []grpc.UnaryServerInterceptor{interceptor} })

	s.NotNil(builder)
	s.NotNil(s.builder.grpcServerInterceptors)
}

func (s *ApplicationBuilderTestSuite) TestWithGrpcServerStatsHandlers() {
	handler := &mockStatsHandler{}
	builder := s.builder.WithGrpcServerStatsHandlers(func() []stats.Handler { return []stats.Handler{handler} })

	s.NotNil(builder)
	s.NotNil(s.builder.grpcServerStatsHandlers)
}

func (s *ApplicationBuilderTestSuite) TestWithJobs() {
	mockJob := mocksqueue.NewJob(s.T())

	builder := s.builder.WithJobs(func() []queue.Job { return []queue.Job{mockJob} })

	s.NotNil(builder)
	s.NotNil(s.builder.jobs)
}

func (s *ApplicationBuilderTestSuite) TestWithSeeders() {
	mockSeeder := mocksseeder.NewSeeder(s.T())

	builder := s.builder.WithSeeders(func() []seeder.Seeder { return []seeder.Seeder{mockSeeder} })

	s.NotNil(builder)
	s.NotNil(s.builder.seeders)
}

func (s *ApplicationBuilderTestSuite) TestWithFilters() {
	mockFilter := mocksvalidation.NewFilter(s.T())

	builder := s.builder.WithFilters(func() []validation.Filter { return []validation.Filter{mockFilter} })

	s.NotNil(builder)
	s.NotNil(s.builder.filters)
}

func (s *ApplicationBuilderTestSuite) TestWithRules() {
	mockRule := mocksvalidation.NewRule(s.T())

	builder := s.builder.WithRules(func() []validation.Rule { return []validation.Rule{mockRule} })

	s.NotNil(builder)
	s.NotNil(s.builder.rules)
}

func (s *ApplicationBuilderTestSuite) TestWithRunners() {
	runner := mocksfoundation.NewRunner(s.T())
	builder := s.builder.WithRunners(func() []contractsfoundation.Runner { return []contractsfoundation.Runner{runner} })

	s.NotNil(builder)
	s.NotNil(s.builder.runners)
}

func (s *ApplicationBuilderTestSuite) TestWithCallback() {
	called := false
	fn := func() {
		called = true
	}

	builder := s.builder.WithCallback(fn)

	s.NotNil(builder)
	s.NotNil(s.builder.callback)

	// Verify callback can be executed
	s.builder.callback()
	s.True(called)
}

type mockStatsHandler struct{ stats.Handler }
