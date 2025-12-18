package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	configmock "github.com/goravel/framework/mocks/config"
)

type contextKey int

const (
	server contextKey = 0
	client contextKey = 1
)

func TestRun(t *testing.T) {
	var (
		app        *Application
		mockConfig *configmock.Config
		name       = "test"
	)

	beforeEach := func() {
		mockConfig = &configmock.Config{}

		app = NewApplication(mockConfig)
		app.UnaryServerInterceptors([]grpc.UnaryServerInterceptor{
			serverInterceptor,
		})
		app.UnaryClientInterceptorGroups(map[string][]grpc.UnaryClientInterceptor{
			"test": {
				clientInterceptor,
			},
		})
		RegisterTestServiceServer(app.Server(), &TestController{})
	}

	tests := []struct {
		name      string
		setup     func()
		expectErr bool
	}{
		{
			name: "success",
			setup: func() {
				host := "127.0.0.1:3030"
				mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.host", name)).Return(host).Once()
				mockConfig.On("Get", fmt.Sprintf("grpc.clients.%s.interceptors", name)).Return([]string{"test"}).Once()

				go func() {
					assert.Nil(t, app.Run(host))
				}()

				time.Sleep(1 * time.Second)
				client, err := app.Client(context.Background(), name)
				assert.Nil(t, err)
				testServiceClient := NewTestServiceClient(client)
				res, err := testServiceClient.Get(context.Background(), &TestRequest{
					Name: "success",
				})

				assert.Equal(t, "Goravel: server: goravel-server, client: goravel-client", res.GetMessage())
				assert.Equal(t, http.StatusOK, int(res.GetCode()))
				assert.Nil(t, err)
			},
		},
		{
			name: "success when host with port",
			setup: func() {
				mockConfig.On("GetString", "grpc.host").Return("127.0.0.1:3032").Once()
				go func() {
					assert.Nil(t, app.Run())
				}()
				time.Sleep(1 * time.Second)
			},
		},
		{
			name: "error when host is empty",
			setup: func() {
				mockConfig.On("GetString", "grpc.host").Return("").Once()
				assert.EqualError(t, app.Run(), "host can't be empty")
			},
		},
		{
			name: "error when port is empty",
			setup: func() {
				mockConfig.On("GetString", "grpc.host").Return("127.0.0.1").Once()
				mockConfig.On("GetString", "grpc.port").Return("").Once()
				assert.EqualError(t, app.Run(), "port can't be empty")
			},
		},
		{
			name: "error when request name = error",
			setup: func() {
				host := "127.0.0.1:3033"
				mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.host", name)).Return(host).Once()
				mockConfig.On("Get", fmt.Sprintf("grpc.clients.%s.interceptors", name)).Return([]string{"test"}).Once()

				go func() {
					assert.Nil(t, app.Run(host))
				}()

				time.Sleep(1 * time.Second)
				client, err := app.Client(context.Background(), "test")
				assert.Nil(t, err)
				testServiceClient := NewTestServiceClient(client)
				res, err := testServiceClient.Get(context.Background(), &TestRequest{
					Name: "error",
				})

				assert.Nil(t, res)
				assert.EqualError(t, err, "rpc error: code = Unknown desc = error")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestClient(t *testing.T) {
	var (
		app        *Application
		mockConfig *configmock.Config
		name       = "user"
		host       = "127.0.0.1:3030"
	)

	beforeEach := func() {
		mockConfig = &configmock.Config{}
		app = NewApplication(mockConfig)
	}

	tests := []struct {
		name      string
		setup     func()
		expectErr bool
	}{
		{
			name: "success",
			setup: func() {
				mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.host", name)).Return(host).Once()
				mockConfig.On("Get", fmt.Sprintf("grpc.clients.%s.interceptors", name)).Return([]string{"trace"}).Once()
				app.UnaryClientInterceptorGroups(map[string][]grpc.UnaryClientInterceptor{
					"trace": {opentracingClient},
				})
			},
		},
		{
			name: "success when interceptors is empty",
			setup: func() {
				mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.host", name)).Return(host).Once()
				mockConfig.On("Get", fmt.Sprintf("grpc.clients.%s.interceptors", name)).Return([]string{"trace"}).Once()
				app.UnaryClientInterceptorGroups(map[string][]grpc.UnaryClientInterceptor{})
			},
		},
		{
			name: "error when host is empty",
			setup: func() {
				mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.host", name)).Return("").Once()
			},
			expectErr: true,
		},
		{
			name: "error when host doesn't have port and port is empty",
			setup: func() {
				mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.host", name)).Return("127.0.0.1").Once()
				mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.port", name)).Return("").Once()
			},
			expectErr: true,
		},
		{
			name: "error when interceptors isn't []string",
			setup: func() {
				mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.host", name)).Return(host).Once()
				mockConfig.On("Get", fmt.Sprintf("grpc.clients.%s.interceptors", name)).Return("trace").Once()
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			client, err := app.Client(context.Background(), name)
			if !test.expectErr {
				assert.NotNil(t, client, test.name)
			}
			assert.Equal(t, test.expectErr, err != nil, test.name)
			mockConfig.AssertExpectations(t)
		})
	}
}

func TestClient_Caching_And_Reuse(t *testing.T) {
	var (
		app        *Application
		mockConfig *configmock.Config
		name       = "user-service"
		host       = "127.0.0.1:3035"
	)

	mockConfig = &configmock.Config{}
	app = NewApplication(mockConfig)

	// We expect GetString to be called ONLY ONCE, even though we call Client() twice.
	mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.host", name)).Return(host).Once()
	mockConfig.On("Get", fmt.Sprintf("grpc.clients.%s.interceptors", name)).Return([]string{}).Once()

	conn1, err := app.Client(context.Background(), name)
	assert.NoError(t, err)
	assert.NotNil(t, conn1)

	conn2, err := app.Client(context.Background(), name)
	assert.NoError(t, err)
	assert.NotNil(t, conn2)

	// The memory address of conn1 and conn2 must be identical
	assert.Same(t, conn1, conn2, "Client should return cached connection instance")

	mockConfig.AssertExpectations(t)
}

func TestShutdown(t *testing.T) {
	var (
		app        *Application
		mockConfig *configmock.Config
	)

	beforeEach := func() {
		mockConfig = &configmock.Config{}
		app = NewApplication(mockConfig)
	}

	tests := []struct {
		name  string
		setup func()
		force bool
	}{
		{
			name: "graceful shutdown",
			setup: func() {
				app.server = grpc.NewServer()
			},
			force: false,
		},
		{
			name: "force shutdown",
			setup: func() {
				app.server = grpc.NewServer()
			},
			force: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			if test.force {
				assert.NoError(t, app.Shutdown(true))
			} else {
				assert.NoError(t, app.Shutdown())
			}
		})
	}
}

func TestListen(t *testing.T) {
	var (
		app        *Application
		mockConfig *configmock.Config
	)

	beforeEach := func() {
		mockConfig = &configmock.Config{}
		app = NewApplication(mockConfig)
	}

	tests := []struct {
		name  string
		setup func() net.Listener
	}{
		{
			name: "success",
			setup: func() net.Listener {
				listener, _ := net.Listen("tcp", "127.0.0.1:0")
				return listener
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			listener := test.setup()

			done := make(chan bool)
			go func() {
				assert.NoError(t, app.Listen(listener), test.name)
				done <- true
			}()

			time.Sleep(1 * time.Second)
			assert.NoError(t, app.Shutdown())
			assert.True(t, <-done)
		})
	}
}

func opentracingClient(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return nil
}

func serverInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	ctx = context.WithValue(ctx, server, "goravel-server")
	if len(md["client"]) > 0 {
		ctx = context.WithValue(ctx, client, md["client"][0])
	}

	return handler(ctx, req)
}

func clientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}

	md["client"] = []string{"goravel-client"}

	if err := invoker(metadata.NewOutgoingContext(ctx, md), method, req, reply, cc, opts...); err != nil {
		return err
	}

	return nil
}

type TestController struct {
	UnimplementedTestServiceServer
}

func (r *TestController) Get(ctx context.Context, req *TestRequest) (*TestResponse, error) {
	if req.GetName() == "success" {
		return &TestResponse{
			Code:    http.StatusOK,
			Message: fmt.Sprintf("Goravel: server: %s, client: %s", ctx.Value(server), ctx.Value(client)),
		}, nil
	} else {
		return nil, errors.New("error")
	}
}
