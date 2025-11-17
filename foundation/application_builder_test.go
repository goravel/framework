package foundation

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	contractsconfiguration "github.com/goravel/framework/contracts/foundation/configuration"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/schedule"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
	mocksevent "github.com/goravel/framework/mocks/event"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksgrpc "github.com/goravel/framework/mocks/grpc"
	mocksroute "github.com/goravel/framework/mocks/route"
	mocksschedule "github.com/goravel/framework/mocks/schedule"
	"github.com/goravel/framework/support/color"
)

type ApplicationBuilderTestSuite struct {
	suite.Suite
	builder *ApplicationBuilder
	mockApp *mocksfoundation.Application
}

func TestApplicationBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationBuilderTestSuite))
}

func (s *ApplicationBuilderTestSuite) SetupTest() {
	s.mockApp = mocksfoundation.NewApplication(s.T())
	s.builder = &ApplicationBuilder{
		app: s.mockApp,
	}
}

func (s *ApplicationBuilderTestSuite) TestCreate() {
	s.Run("Without any With functions", func() {
		s.SetupTest()

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()

		app := s.builder.Create()

		s.NotNil(app)
	})

	s.Run("WithConfig", func() {
		s.SetupTest()
		calledConfig := false

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()

		app := s.builder.
			WithConfig(func() {
				calledConfig = true
			}).
			Create()

		s.NotNil(app)
		s.True(calledConfig)
	})

	s.Run("WithEvents but Event facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeEvent().Return(nil).Once()

		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.
				WithEvents(map[event.Event][]event.Listener{
					mocksevent.NewEvent(s.T()): {mocksevent.NewListener(s.T())},
				}).
				Create()

			s.NotNil(app)
		})

		s.Contains(got, "Event facade not found, please install it first: ./artisan package:install Event")
	})

	s.Run("WithEvents but Events is empty", func() {
		s.SetupTest()

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()

		app := s.builder.
			WithEvents(map[event.Event][]event.Listener{}).
			Create()

		s.NotNil(app)
	})

	s.Run("WithEvents", func() {
		s.SetupTest()

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()

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

		app := s.builder.
			WithEvents(map[event.Event][]event.Listener{
				mockEvent: {mockListener},
			}).
			Create()

		s.NotNil(app)
	})

	s.Run("WithMiddleware but Route facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeRoute().Return(nil).Once()

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

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeRoute().Return(mockRoute).Once()

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
		s.mockApp.EXPECT().Boot().Return().Once()

		app := s.builder.WithProviders(providers).Create()

		s.NotNil(app)
	})

	s.Run("WithRouting", func() {
		s.SetupTest()
		calledRouting := false

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()

		app := s.builder.
			WithRouting([]func(){
				func() {
					calledRouting = true
				},
			}).
			Create()

		s.NotNil(app)
		s.True(calledRouting)
	})

	s.Run("WithCommands but Artisan facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeArtisan().Return(nil).Once()

		mockCommand := mocksconsole.NewCommand(s.T())
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithCommands([]console.Command{mockCommand}).Create()

			s.NotNil(app)
		})

		s.Contains(got, "Artisan facade not found, please install it first: ./artisan package:install Artisan")
	})

	s.Run("WithCommands", func() {
		s.SetupTest()

		mockArtisan := mocksconsole.NewArtisan(s.T())
		mockCommand := mocksconsole.NewCommand(s.T())

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeArtisan().Return(mockArtisan).Once()
		mockArtisan.EXPECT().Register([]console.Command{mockCommand}).Return().Once()

		app := s.builder.WithCommands([]console.Command{mockCommand}).Create()

		s.NotNil(app)
	})

	s.Run("WithSchedule but Schedule facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeSchedule().Return(nil).Once()

		mockEvent := mocksschedule.NewEvent(s.T())
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithSchedule([]schedule.Event{mockEvent}).Create()

			s.NotNil(app)
		})

		s.Contains(got, "Schedule facade not found, please install it first: ./artisan package:install Schedule")
	})

	s.Run("WithSchedule", func() {
		s.SetupTest()

		mockSchedule := mocksschedule.NewSchedule(s.T())
		mockEvent := mocksschedule.NewEvent(s.T())

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeSchedule().Return(mockSchedule).Once()
		mockSchedule.EXPECT().Register([]schedule.Event{mockEvent}).Return().Once()

		app := s.builder.WithSchedule([]schedule.Event{mockEvent}).Create()

		s.NotNil(app)
	})

	s.Run("WithMigrations but Schema facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeSchema().Return(nil).Once()

		mockMigration := mocksschema.NewMigration(s.T())
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithMigrations([]schema.Migration{mockMigration}).Create()
			s.NotNil(app)
		})

		s.Contains(got, "Schema facade not found, please install it first: ./artisan package:install Schema")
	})

	s.Run("WithMigrations", func() {
		s.SetupTest()

		mockSchema := mocksschema.NewSchema(s.T())
		mockMigration := mocksschema.NewMigration(s.T())

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeSchema().Return(mockSchema).Once()
		mockSchema.EXPECT().Register([]schema.Migration{mockMigration}).Return().Once()

		app := s.builder.WithMigrations([]schema.Migration{mockMigration}).Create()

		s.NotNil(app)
	})

	s.Run("WithGrpcClientInterceptors but Grpc facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(nil).Once()

		interceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return nil
		}
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithGrpcClientInterceptors(map[string][]grpc.UnaryClientInterceptor{
				"test": {interceptor},
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

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(mockGrpc).Once()
		mockGrpc.EXPECT().UnaryClientInterceptorGroups(interceptors).Return().Once()

		app := s.builder.WithGrpcClientInterceptors(interceptors).Create()

		s.NotNil(app)
	})

	s.Run("WithGrpcServerInterceptors but Grpc facade is nil", func() {
		s.SetupTest()

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(nil).Once()

		interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return nil, nil
		}
		got := color.CaptureOutput(func(io.Writer) {
			app := s.builder.WithGrpcServerInterceptors([]grpc.UnaryServerInterceptor{interceptor}).Create()
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

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(mockGrpc).Once()
		mockGrpc.EXPECT().UnaryServerInterceptors(interceptors).Return().Once()

		app := s.builder.WithGrpcServerInterceptors(interceptors).Create()

		s.NotNil(app)
	})

	s.Run("WithGrpcClientInterceptors and WithGrpcServerInterceptors", func() {
		s.SetupTest()

		mockGrpc := mocksgrpc.NewGrpc(s.T())
		clientInterceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return nil
		}
		serverInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return nil, nil
		}
		clientInterceptors := map[string][]grpc.UnaryClientInterceptor{
			"test": {clientInterceptor},
		}
		serverInterceptors := []grpc.UnaryServerInterceptor{serverInterceptor}

		s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeGrpc().Return(mockGrpc).Once()
		mockGrpc.EXPECT().UnaryClientInterceptorGroups(clientInterceptors).Return().Once()
		mockGrpc.EXPECT().UnaryServerInterceptors(serverInterceptors).Return().Once()

		app := s.builder.
			WithGrpcClientInterceptors(clientInterceptors).
			WithGrpcServerInterceptors(serverInterceptors).
			Create()

		s.NotNil(app)
	})
}

func (s *ApplicationBuilderTestSuite) TestRun() {
	s.mockApp.EXPECT().AddServiceProviders([]foundation.ServiceProvider(nil)).Return().Once()
	s.mockApp.EXPECT().Boot().Return().Once()
	s.mockApp.EXPECT().Run().Return().Once()

	s.builder.Run()
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

func (s *ApplicationBuilderTestSuite) TestWithMigrations() {
	mockMigration := mocksschema.NewMigration(s.T())

	builder := s.builder.WithMigrations([]schema.Migration{mockMigration})

	s.NotNil(builder)
	s.Len(s.builder.migrations, 1)
}

func (s *ApplicationBuilderTestSuite) TestWithProviders() {
	mockProvider1 := mocksfoundation.NewServiceProvider(s.T())
	mockProvider2 := mocksfoundation.NewServiceProvider(s.T())
	providers1 := []foundation.ServiceProvider{mockProvider1}
	providers2 := []foundation.ServiceProvider{mockProvider2}

	builder := s.builder.WithProviders(providers1)

	s.NotNil(builder)
	s.Equal(providers1, s.builder.configuredServiceProviders)

	builder.WithProviders(providers2)
	expectedProviders := []foundation.ServiceProvider{mockProvider1, mockProvider2}
	s.Equal(expectedProviders, s.builder.configuredServiceProviders)
}

func (s *ApplicationBuilderTestSuite) TestWithRouting() {
	fn := func() {}

	builder := s.builder.WithRouting([]func(){fn})

	s.NotNil(builder)
	s.NotNil(s.builder.routes)
}

func (s *ApplicationBuilderTestSuite) TestWithSchedule() {
	mockEvent := mocksschedule.NewEvent(s.T())

	builder := s.builder.WithSchedule([]schedule.Event{mockEvent})

	s.NotNil(builder)
	s.Len(s.builder.scheduledEvents, 1)
}

func (s *ApplicationBuilderTestSuite) TestWithGrpcClientInterceptors() {
	interceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return nil
	}
	interceptors := map[string][]grpc.UnaryClientInterceptor{
		"test": {interceptor},
	}

	builder := s.builder.WithGrpcClientInterceptors(interceptors)

	s.NotNil(builder)
	s.Equal(interceptors, s.builder.grpcClientInterceptors)
}

func (s *ApplicationBuilderTestSuite) TestWithGrpcServerInterceptors() {
	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return nil, nil
	}
	interceptors := []grpc.UnaryServerInterceptor{interceptor}

	builder := s.builder.WithGrpcServerInterceptors(interceptors)

	s.NotNil(builder)
	s.Len(s.builder.grpcServerInterceptors, 1)
}
