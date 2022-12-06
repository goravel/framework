package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/goravel/framework/facades"
	"net"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Application struct {
	server                       *grpc.Server
	unaryClientInterceptorGroups map[string][]grpc.UnaryClientInterceptor
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Server() *grpc.Server {
	return app.server
}

func (app *Application) Client(ctx context.Context, name string) (*grpc.ClientConn, error) {
	host := facades.Config.GetString(fmt.Sprintf("grpc.clients.%s.host", name))
	if host == "" {
		return nil, errors.New("client is not defined")
	}

	interceptors, ok := facades.Config.Get(fmt.Sprintf("grpc.clients.%s.interceptors", name)).([]string)
	if !ok {
		return nil, fmt.Errorf("the type of clients.%s.interceptors must be []string", name)
	}

	clientInterceptors := app.getClientInterceptors(interceptors)

	return grpc.DialContext(
		ctx,
		host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(clientInterceptors...),
	)
}

func (app *Application) Run(host string) error {
	listen, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}
	if err := app.server.Serve(listen); err != nil {
		return err
	}

	return nil
}

func (app *Application) UnaryServerInterceptors(unaryServerInterceptors []grpc.UnaryServerInterceptor) {
	app.server = grpc.NewServer(grpc.UnaryInterceptor(
		grpcmiddleware.ChainUnaryServer(unaryServerInterceptors...),
	))
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
