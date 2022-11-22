package grpc

import (
	"context"
	"fmt"
	"testing"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/testing/mock"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

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
		app.UnaryServerInterceptors([]grpc.UnaryServerInterceptor{})
		go app.Run(host)
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
					"trace": {OpentracingClient},
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

func OpentracingClient(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return nil
}
