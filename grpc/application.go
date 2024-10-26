package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

type Application struct {
	config                       config.Config
	server                       *grpc.Server
	unaryClientInterceptorGroups map[string][]grpc.UnaryClientInterceptor
}

func NewApplication(config config.Config) *Application {
	return &Application{
		config: config,
	}
}

func (app *Application) Server() *grpc.Server {
	return app.server
}

func (app *Application) Client(ctx context.Context, name string) (*grpc.ClientConn, error) {
	host := app.config.GetString(fmt.Sprintf("grpc.clients.%s.host", name))
	if host == "" {
		return nil, errors.GrpcEmptyClientHost
	}
	if !strings.Contains(host, ":") {
		port := app.config.GetString(fmt.Sprintf("grpc.clients.%s.port", name))
		if port == "" {
			return nil, errors.GrpcEmptyClientPort
		}

		host += ":" + port
	}

	interceptors, ok := app.config.Get(fmt.Sprintf("grpc.clients.%s.interceptors", name)).([]string)
	if !ok {
		return nil, errors.GrpcInvalidInterceptorsType.Args(name)
	}

	clientInterceptors := app.getClientInterceptors(interceptors)

	return grpc.NewClient(
		host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(clientInterceptors...),
	)
}

func (app *Application) Run(host ...string) error {
	if len(host) == 0 {
		defaultHost := app.config.GetString("grpc.host")
		if defaultHost == "" {
			return errors.GrpcEmptyServerHost
		}

		if !strings.Contains(defaultHost, ":") {
			defaultPort := app.config.GetString("grpc.port")
			if defaultPort == "" {
				return errors.GrpcEmptyServerPort
			}
			defaultHost += ":" + defaultPort
		}

		host = append(host, defaultHost)
	}

	listen, err := net.Listen("tcp", host[0])
	if err != nil {
		return err
	}
	color.Successln("[GRPC] Listening and serving gRPC on " + host[0])
	if err := app.server.Serve(listen); err != nil {
		return err
	}

	return nil
}

func (app *Application) UnaryServerInterceptors(unaryServerInterceptors []grpc.UnaryServerInterceptor) {
	app.server = grpc.NewServer(grpc.ChainUnaryInterceptor(unaryServerInterceptors...))
}

func (app *Application) UnaryClientInterceptorGroups(unaryClientInterceptorGroups map[string][]grpc.UnaryClientInterceptor) {
	app.unaryClientInterceptorGroups = unaryClientInterceptorGroups
}

func (app *Application) getClientInterceptors(interceptors []string) []grpc.UnaryClientInterceptor {
	var unaryClientInterceptors []grpc.UnaryClientInterceptor
	for _, interceptor := range interceptors {
		for client, clientInterceptors := range app.unaryClientInterceptorGroups {
			if interceptor == client {
				unaryClientInterceptors = append(unaryClientInterceptors, clientInterceptors...)
			}
		}
	}

	return unaryClientInterceptors
}
