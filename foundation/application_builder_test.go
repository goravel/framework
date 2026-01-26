package foundation

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	contractsconfiguration "github.com/goravel/framework/contracts/foundation/configuration"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/contracts/validation"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
	mocksseeder "github.com/goravel/framework/mocks/database/seeder"
	mocksevent "github.com/goravel/framework/mocks/event"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksgrpc "github.com/goravel/framework/mocks/grpc"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	mocksroute "github.com/goravel/framework/mocks/route"
	mocksschedule "github.com/goravel/framework/mocks/schedule"
	mocksvalidation "github.com/goravel/framework/mocks/validation"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
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

func (s *ApplicationBuilderTestSuite) TestCreate() {
	s.Run("Without any With functions", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.Create()

		s.NotNil(app)
	})

	s.Run("WithConfig", func() {
		s.SetupTest()
		calledConfig := false

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.
			WithConfig(func() {
				calledConfig = true
			}).
			Create()

		s.NotNil(app)
		s.True(calledConfig)
	})

	s.Run("WithPaths", func() {
		s.SetupTest()
		defer s.TearDownTest()
		calledPaths := false

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.
			WithPaths(func(paths contractsconfiguration.Paths) {
				calledPaths = true
				paths.Models("custom/models")
				paths.Controllers("custom/controllers")
			}).
			Create()

		s.NotNil(app)
		s.True(calledPaths)
	})

	s.Run("WithEvents but Event facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeEvent().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.
				WithEvents(func() map[event.Event][]event.Listener {
					return map[event.Event][]event.Listener{
						mocksevent.NewEvent(s.T()): {mocksevent.NewListener(s.T())},
					}
				}).
				Create()

			s.NotNil(app)
		})

		s.Contains(got, "Event facade not found, please install it first: ./artisan package:install Event")
	})

	s.Run("WithEvents but Events is empty", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.
			WithEvents(func() map[event.Event][]event.Listener {
				return map[event.Event][]event.Listener{}
			}).
			Create()

		s.NotNil(app)
	})

	s.Run("WithEvents", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()

		mockEvent := mocksevent.NewEvent(s.T())
		mockListener := mocksevent.NewListener(s.T())
		mockEventFacade := mocksevent.NewInstance(s.T())
		mockEventFacade.EXPECT().
			Register(map[event.Event][]event.Listener{
				mockEvent: {mockListener},
			}).
			Return().
			Once()
		s.mockApp.EXPECT().MakeEvent().Return(mockEventFacade).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.
			WithEvents(func() map[event.Event][]event.Listener {
				return map[event.Event][]event.Listener{
					mockEvent: {mockListener},
				}
			}).
			Create()

		s.NotNil(app)
	})

	s.Run("WithMiddleware but Route facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeRoute().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		calledMiddleware := false
		fn := func(middleware contractsconfiguration.Middleware) {
			calledMiddleware = true
		}

		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithMiddleware(fn).Create()

			s.NotNil(app)
		})

		s.Contains(got, "Route facade not found, please install it first: ./artisan package:install Route")
		s.False(calledMiddleware)
	})

	s.Run("WithMiddleware", func() {
		s.SetupTest()

		mockRoute := mocksroute.NewRoute(s.T())
		mockRoute.EXPECT().GetGlobalMiddleware().Return(nil).Once()
		mockRoute.EXPECT().SetGlobalMiddleware([]contractshttp.Middleware(nil)).Return().Once()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeRoute().Return(mockRoute).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		calledMiddleware := false
		fn := func(middleware contractsconfiguration.Middleware) {
			calledMiddleware = true
		}

		app := s.builder.WithMiddleware(fn).Create()

		s.NotNil(app)
		s.True(calledMiddleware)
	})

	s.Run("WithProviders", func() {
		s.SetupTest()

		mockProvider := mocksfoundation.NewServiceProvider(s.T())
		providers := []foundation.ServiceProvider{mockProvider}

		s.mockApp.EXPECT().AddServiceProviders(providers).Return().Once()
		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithProviders(func() []foundation.ServiceProvider {
			return providers
		}).Create()

		s.NotNil(app)
	})

	s.Run("WithRouting", func() {
		s.SetupTest()
		calledRouting := false

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.
			WithRouting(func() {
				calledRouting = true
			}).
			Create()

		s.NotNil(app)
		s.True(calledRouting)
	})

	s.Run("WithCommands but Artisan facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeArtisan().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		mockCommand := mocksconsole.NewCommand(s.T())
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithCommands(func() []console.Command {
				return []console.Command{mockCommand}
			}).Create()

			s.NotNil(app)
		})

		s.Contains(got, "Artisan facade not found, please install it first: ./artisan package:install Artisan")
	})

	s.Run("WithCommands", func() {
		s.SetupTest()

		mockArtisan := mocksconsole.NewArtisan(s.T())
		mockCommand := mocksconsole.NewCommand(s.T())

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeArtisan().Return(mockArtisan).Once()
		mockArtisan.EXPECT().Register([]console.Command{mockCommand}).Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithCommands(func() []console.Command {
			return []console.Command{mockCommand}
		}).Create()

		s.NotNil(app)
	})

	s.Run("WithSchedule but Schedule facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeSchedule().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		mockEvent := mocksschedule.NewEvent(s.T())
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithSchedule(func() []schedule.Event {
				return []schedule.Event{mockEvent}
			}).Create()

			s.NotNil(app)
		})

		s.Contains(got, "Schedule facade not found, please install it first: ./artisan package:install Schedule")
	})

	s.Run("WithSchedule", func() {
		s.SetupTest()

		mockSchedule := mocksschedule.NewSchedule(s.T())
		mockEvent := mocksschedule.NewEvent(s.T())

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeSchedule().Return(mockSchedule).Once()
		mockSchedule.EXPECT().Register([]schedule.Event{mockEvent}).Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithSchedule(func() []schedule.Event {
			return []schedule.Event{mockEvent}
		}).Create()

		s.NotNil(app)
	})

	s.Run("WithMigrations but Schema facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeSchema().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		mockMigration := mocksschema.NewMigration(s.T())
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithMigrations(func() []schema.Migration {
				return []schema.Migration{mockMigration}
			}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "Schema facade not found, please install it first: ./artisan package:install Schema")
	})

	s.Run("WithMigrations", func() {
		s.SetupTest()

		mockSchema := mocksschema.NewSchema(s.T())
		mockMigration := mocksschema.NewMigration(s.T())

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeSchema().Return(mockSchema).Once()
		mockSchema.EXPECT().Register([]schema.Migration{mockMigration}).Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithMigrations(func() []schema.Migration {
			return []schema.Migration{mockMigration}
		}).Create()

		s.NotNil(app)
	})

	s.Run("WithGrpcClientInterceptors but Grpc facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		interceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return nil
		}
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithGrpcClientInterceptors(func() map[string][]grpc.UnaryClientInterceptor {
				return map[string][]grpc.UnaryClientInterceptor{
					"test": {interceptor},
				}
			}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "gRPC facade not found, please install it first: ./artisan package:install Grpc")
	})

	s.Run("WithGrpcClientInterceptors", func() {
		s.SetupTest()

		mockGrpc := mocksgrpc.NewGrpc(s.T())
		interceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return nil
		}
		interceptors := map[string][]grpc.UnaryClientInterceptor{
			"test": {interceptor},
		}

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(mockGrpc).Once()
		mockGrpc.EXPECT().UnaryClientInterceptorGroups(mock.MatchedBy(func(interceptors map[string][]grpc.UnaryClientInterceptor) bool {
			return len(interceptors) == 1 && len(interceptors["test"]) == 1
		})).Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithGrpcClientInterceptors(func() map[string][]grpc.UnaryClientInterceptor {
			return interceptors
		}).Create()

		s.NotNil(app)
	})

	s.Run("WithGrpcServerInterceptors but Grpc facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return nil, nil
		}
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithGrpcServerInterceptors(func() []grpc.UnaryServerInterceptor {
				return []grpc.UnaryServerInterceptor{interceptor}
			}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "gRPC facade not found, please install it first: ./artisan package:install Grpc")
	})

	s.Run("WithGrpcServerInterceptors", func() {
		s.SetupTest()

		mockGrpc := mocksgrpc.NewGrpc(s.T())
		interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return nil, nil
		}
		interceptors := []grpc.UnaryServerInterceptor{interceptor}

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(mockGrpc).Once()
		mockGrpc.EXPECT().UnaryServerInterceptors(mock.MatchedBy(func(interceptors []grpc.UnaryServerInterceptor) bool {
			return len(interceptors) == 1
		})).Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithGrpcServerInterceptors(func() []grpc.UnaryServerInterceptor {
			return interceptors
		}).Create()

		s.NotNil(app)
	})

	s.Run("WithGrpcClientStatsHandlers but Grpc facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		handler := &mockStatsHandler{}
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithGrpcClientStatsHandlers(func() map[string][]stats.Handler {
				return map[string][]stats.Handler{
					"test": {handler},
				}
			}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "gRPC facade not found, please install it first: ./artisan package:install Grpc")
	})

	s.Run("WithGrpcClientStatsHandlers", func() {
		s.SetupTest()

		mockGrpc := mocksgrpc.NewGrpc(s.T())
		handler := &mockStatsHandler{}
		handlers := map[string][]stats.Handler{
			"test": {handler},
		}

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(mockGrpc).Once()
		mockGrpc.EXPECT().ClientStatsHandlerGroups(handlers).Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithGrpcClientStatsHandlers(func() map[string][]stats.Handler {
			return handlers
		}).Create()

		s.NotNil(app)
	})

	s.Run("WithGrpcServerStatsHandlers but Grpc facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		handler := &mockStatsHandler{}
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithGrpcServerStatsHandlers(func() []stats.Handler {
				return []stats.Handler{handler}
			}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "gRPC facade not found, please install it first: ./artisan package:install Grpc")
	})

	s.Run("WithGrpcServerStatsHandlers", func() {
		s.SetupTest()

		mockGrpc := mocksgrpc.NewGrpc(s.T())
		handler := &mockStatsHandler{}
		handlers := []stats.Handler{handler}

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(mockGrpc).Once()
		mockGrpc.EXPECT().ServerStatsHandlers(handlers).Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithGrpcServerStatsHandlers(func() []stats.Handler {
			return handlers
		}).Create()

		s.NotNil(app)
	})

	s.Run("With All gRPC Components", func() {
		s.SetupTest()

		mockGrpc := mocksgrpc.NewGrpc(s.T())
		clientInterceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return nil
		}
		serverInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return nil, nil
		}
		clientStatsHandler := &mockStatsHandler{}
		serverStatsHandler := &mockStatsHandler{}

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(mockGrpc).Once()

		mockGrpc.EXPECT().UnaryClientInterceptorGroups(mock.MatchedBy(func(m map[string][]grpc.UnaryClientInterceptor) bool {
			return len(m["test"]) == 1
		})).Return().Once()
		mockGrpc.EXPECT().UnaryServerInterceptors(mock.MatchedBy(func(s []grpc.UnaryServerInterceptor) bool {
			return len(s) == 1
		})).Return().Once()
		mockGrpc.EXPECT().ClientStatsHandlerGroups(map[string][]stats.Handler{"test": {clientStatsHandler}}).Return().Once()
		mockGrpc.EXPECT().ServerStatsHandlers([]stats.Handler{serverStatsHandler}).Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.
			WithGrpcClientInterceptors(func() map[string][]grpc.UnaryClientInterceptor {
				return map[string][]grpc.UnaryClientInterceptor{"test": {clientInterceptor}}
			}).
			WithGrpcServerInterceptors(func() []grpc.UnaryServerInterceptor {
				return []grpc.UnaryServerInterceptor{serverInterceptor}
			}).
			WithGrpcClientStatsHandlers(func() map[string][]stats.Handler {
				return map[string][]stats.Handler{"test": {clientStatsHandler}}
			}).
			WithGrpcServerStatsHandlers(func() []stats.Handler {
				return []stats.Handler{serverStatsHandler}
			}).
			Create()

		s.NotNil(app)
	})

	s.Run("WithJobs but Queue facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeQueue().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		mockJob := mocksqueue.NewJob(s.T())
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithJobs(func() []queue.Job {
				return []queue.Job{mockJob}
			}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "Queue facade not found, please install it first: ./artisan package:install Queue")
	})

	s.Run("WithJobs", func() {
		s.SetupTest()

		mockQueue := mocksqueue.NewQueue(s.T())
		mockJob := mocksqueue.NewJob(s.T())

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeQueue().Return(mockQueue).Once()
		mockQueue.EXPECT().Register([]queue.Job{mockJob}).Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithJobs(func() []queue.Job {
			return []queue.Job{mockJob}
		}).Create()

		s.NotNil(app)
	})

	s.Run("WithSeeders but Seeder facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeSeeder().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		mockSeeder := mocksseeder.NewSeeder(s.T())
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithSeeders(func() []seeder.Seeder {
				return []seeder.Seeder{mockSeeder}
			}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "Seeder facade not found, please install it first: ./artisan package:install Seeder")
	})

	s.Run("WithSeeders", func() {
		s.SetupTest()

		mockSeederFacade := mocksseeder.NewFacade(s.T())
		mockSeeder := mocksseeder.NewSeeder(s.T())

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeSeeder().Return(mockSeederFacade).Once()
		mockSeederFacade.EXPECT().Register([]seeder.Seeder{mockSeeder}).Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithSeeders(func() []seeder.Seeder {
			return []seeder.Seeder{mockSeeder}
		}).Create()

		s.NotNil(app)
	})

	s.Run("WithFilters but Validation facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeValidation().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		mockFilter := mocksvalidation.NewFilter(s.T())
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithFilters(func() []validation.Filter {
				return []validation.Filter{mockFilter}
			}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "Validation facade not found, please install it first: ./artisan package:install Validation")
	})

	s.Run("WithFilters but AddFilters returns error", func() {
		s.SetupTest()

		mockValidation := mocksvalidation.NewValidation(s.T())
		mockFilter := mocksvalidation.NewFilter(s.T())

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeValidation().Return(mockValidation).Once()
		mockValidation.EXPECT().AddFilters([]validation.Filter{mockFilter}).Return(errors.New("validation error")).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithFilters(func() []validation.Filter {
				return []validation.Filter{mockFilter}
			}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "add validation filters error:")
	})

	s.Run("WithFilters", func() {
		s.SetupTest()

		mockValidation := mocksvalidation.NewValidation(s.T())
		mockFilter := mocksvalidation.NewFilter(s.T())

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeValidation().Return(mockValidation).Once()
		mockValidation.EXPECT().AddFilters([]validation.Filter{mockFilter}).Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithFilters(func() []validation.Filter {
			return []validation.Filter{mockFilter}
		}).Create()

		s.NotNil(app)
	})

	s.Run("WithRules but Validation facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeValidation().Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		mockRule := mocksvalidation.NewRule(s.T())
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithRules(func() []validation.Rule {
				return []validation.Rule{mockRule}
			}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "Validation facade not found, please install it first: ./artisan package:install Validation")
	})

	s.Run("WithRules but AddRules returns error", func() {
		s.SetupTest()

		mockValidation := mocksvalidation.NewValidation(s.T())
		mockRule := mocksvalidation.NewRule(s.T())

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeValidation().Return(mockValidation).Once()
		mockValidation.EXPECT().AddRules([]validation.Rule{mockRule}).Return(errors.New("validation error")).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithRules(func() []validation.Rule {
				return []validation.Rule{mockRule}
			}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "add validation rules error:")
	})

	s.Run("WithRules", func() {
		s.SetupTest()

		mockValidation := mocksvalidation.NewValidation(s.T())
		mockRule := mocksvalidation.NewRule(s.T())

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().MakeValidation().Return(mockValidation).Once()
		mockValidation.EXPECT().AddRules([]validation.Rule{mockRule}).Return(nil).Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.WithRules(func() []validation.Rule {
			return []validation.Rule{mockRule}
		}).Create()

		s.NotNil(app)
	})

	s.Run("WithCallback", func() {
		s.SetupTest()
		calledCallback := false

		s.mockApp.EXPECT().RegisterServiceProviders().Return().Once()
		s.mockApp.EXPECT().BootServiceProviders().Return().Once()

		app := s.builder.
			WithCallback(func() {
				calledCallback = true
			}).
			Create()

		s.NotNil(app)
		s.True(calledCallback)
	})
}

func (s *ApplicationBuilderTestSuite) TestWithConfig() {
	fn := func() {}

	builder := s.builder.WithConfig(fn)

	s.NotNil(builder)
	s.NotNil(s.builder.config)
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
