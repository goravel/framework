package grpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/testing/mock"
)

func TestRun(t *testing.T) {
	var (
		app        *Application
		mockConfig *configmocks.Config
		name       = "test"
	)

	beforeEach := func() {
		mockConfig = mock.Config()
		mockConfig.On("Get", fmt.Sprintf("grpc.clients.%s.interceptors", name)).Return([]string{"test"}).Once()

		app = NewApplication()
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
				host := "127.0.0.1:3001"
				mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.host", name)).Return(host).Once()

				go func() {
					assert.Nil(t, app.Run(host))
				}()

				client, err := app.Client(context.Background(), name)
				assert.Nil(t, err)
				testServiceClient := NewTestServiceClient(client)
				res, err := testServiceClient.Get(context.Background(), &TestRequest{
					Name: "success",
				})

				assert.Equal(t, &TestResponse{Code: http.StatusOK, Message: "Goravel: server: goravel-server, client: goravel-client"}, res)
				assert.Nil(t, err)
			},
		},
		{
			name: "error",
			setup: func() {
				host := "127.0.0.1:3002"
				mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.host", name)).Return(host).Once()

				go func() {
					assert.Nil(t, app.Run(host))
				}()

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
		beforeEach()
		test.setup()
		mockConfig.AssertExpectations(t)
	}
}

func TestClient(t *testing.T) {
	var (
		app        *Application
		mockConfig *configmocks.Config
		name       = "user"
		host       = "127.0.0.1:3001"
	)

	beforeEach := func() {
		mockConfig = mock.Config()
		app = NewApplication()
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
			name: "error when interceptors isn't []string",
			setup: func() {
				mockConfig.On("GetString", fmt.Sprintf("grpc.clients.%s.host", name)).Return(host).Once()
				mockConfig.On("Get", fmt.Sprintf("grpc.clients.%s.interceptors", name)).Return("trace").Once()
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		beforeEach()
		test.setup()
		client, err := app.Client(context.Background(), name)
		if !test.expectErr {
			assert.NotNil(t, client, test.name)
		}
		assert.Equal(t, test.expectErr, err != nil, test.name)
		mockConfig.AssertExpectations(t)
	}
}

func opentracingClient(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return nil
}

func serverInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	ctx = context.WithValue(ctx, "server", "goravel-server")
	if len(md["client"]) > 0 {
		ctx = context.WithValue(ctx, "client", md["client"][0])
	}

	return handler(ctx, req)
}

func clientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
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
}

func (r *TestController) Get(ctx context.Context, req *TestRequest) (*TestResponse, error) {
	if req.GetName() == "success" {
		return &TestResponse{
			Code:    http.StatusOK,
			Message: fmt.Sprintf("Goravel: server: %s, client: %s", ctx.Value("server"), ctx.Value("client")),
		}, nil
	} else {
		return nil, errors.New("error")
	}
}
