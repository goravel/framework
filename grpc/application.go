package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

type Application struct {
	config                       config.Config
	server                       *grpc.Server
	unaryClientInterceptorGroups map[string][]grpc.UnaryClientInterceptor

	// Mutex protects the clients map
	mu      sync.RWMutex
	clients map[string]*grpc.ClientConn
}

func NewApplication(config config.Config) *Application {
	return &Application{
		server:  grpc.NewServer(),
		config:  config,
		clients: make(map[string]*grpc.ClientConn),
	}
}

func (app *Application) Client(ctx context.Context, name string) (*grpc.ClientConn, error) {
	app.mu.RLock()
	conn, ok := app.clients[name]
	app.mu.RUnlock()

	if ok {
		// If connection exists and is healthy, return it immediately
		if conn.GetState() != connectivity.Shutdown {
			return conn, nil
		}
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	// Double-Check: Someone else might have created it while we waited for the lock
	if conn, ok = app.clients[name]; ok {
		if conn.GetState() != connectivity.Shutdown {
			return conn, nil
		}
		// Found a Shutdown connection. Close and remove it immediately.
		// This prevents stale connections from lingering if the subsequent creation fails.
		_ = conn.Close()
		delete(app.clients, name)
	}

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

	newConn, err := grpc.NewClient(
		host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(clientInterceptors...),
	)
	if err != nil {
		return nil, err
	}

	app.clients[name] = newConn

	return newConn, nil
}

func (app *Application) Listen(l net.Listener) error {
	color.Green().Println("[GRPC] Listening on: " + l.Addr().String())
	return app.server.Serve(l)
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

	color.Green().Println("[GRPC] Listening on: " + host[0])
	return app.server.Serve(listen)
}

func (app *Application) Server() *grpc.Server {
	return app.server
}

func (app *Application) Shutdown(force ...bool) error {
	if len(force) > 0 && force[0] {
		app.server.Stop()
	} else {
		app.server.GracefulStop()
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	for _, conn := range app.clients {
		_ = conn.Close()
	}

	// Clear the map to allow Garbage Collection
	app.clients = make(map[string]*grpc.ClientConn)

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
