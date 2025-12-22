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
	"google.golang.org/grpc/stats"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

type Application struct {
	config config.Config
	server *grpc.Server

	// Server Options
	unaryServerInterceptors []grpc.UnaryServerInterceptor
	serverStatsHandlers     []stats.Handler

	// Client Options
	unaryClientInterceptorGroups map[string][]grpc.UnaryClientInterceptor
	clientStatsHandlerGroups     map[string][]stats.Handler

	// Mutex protects the clients map
	mu      sync.RWMutex
	clients map[string]*grpc.ClientConn
}

func NewApplication(config config.Config) *Application {
	return &Application{
		config:                       config,
		clients:                      make(map[string]*grpc.ClientConn),
		unaryServerInterceptors:      make([]grpc.UnaryServerInterceptor, 0),
		serverStatsHandlers:          make([]stats.Handler, 0),
		unaryClientInterceptorGroups: make(map[string][]grpc.UnaryClientInterceptor),
		clientStatsHandlerGroups:     make(map[string][]stats.Handler),
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

	interceptorKeys, ok := app.config.Get(fmt.Sprintf("grpc.clients.%s.interceptors", name)).([]string)
	if !ok {
		return nil, errors.GrpcInvalidInterceptorsType.Args(name)
	}

	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if interceptors := app.getClientInterceptors(interceptorKeys); len(interceptors) > 0 {
		dialOpts = append(dialOpts, grpc.WithChainUnaryInterceptor(interceptors...))
	}

	if handlers := app.getClientStatsHandlers(interceptorKeys); len(handlers) > 0 {
		for _, h := range handlers {
			if h == nil {
				continue
			}
			dialOpts = append(dialOpts, grpc.WithStatsHandler(h))
		}
	}

	newConn, err := grpc.NewClient(host, dialOpts...)
	if err != nil {
		return nil, err
	}

	app.clients[name] = newConn

	return newConn, nil
}

func (app *Application) Listen(l net.Listener) error {
	color.Green().Println("[GRPC] Listening on: " + l.Addr().String())
	return app.Server().Serve(l)
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
	return app.Server().Serve(listen)
}

func (app *Application) Server() *grpc.Server {
	if app.server != nil {
		return app.server
	}

	var opts []grpc.ServerOption

	if len(app.unaryServerInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(app.unaryServerInterceptors...))
	}

	for _, h := range app.serverStatsHandlers {
		if h == nil {
			continue
		}
		opts = append(opts, grpc.StatsHandler(h))
	}

	app.server = grpc.NewServer(opts...)
	return app.server
}

func (app *Application) Shutdown(force ...bool) error {
	if app.server != nil {
		if len(force) > 0 && force[0] {
			app.server.Stop()
		} else {
			app.server.GracefulStop()
		}
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
	if app.server != nil {
		color.Warningln("[GRPC] Server already initialized; unary server interceptor registration ignored.")
		return
	}
	app.unaryServerInterceptors = append(app.unaryServerInterceptors, unaryServerInterceptors...)
}

func (app *Application) ServerStatsHandlers(handlers []stats.Handler) {
	if app.server != nil {
		color.Warningln("[GRPC] Server already initialized; server stats handler registration ignored.")
		return
	}
	app.serverStatsHandlers = append(app.serverStatsHandlers, handlers...)
}

func (app *Application) UnaryClientInterceptorGroups(groups map[string][]grpc.UnaryClientInterceptor) {
	for key, interceptors := range groups {
		app.unaryClientInterceptorGroups[key] = append(app.unaryClientInterceptorGroups[key], interceptors...)
	}
}

func (app *Application) ClientStatsHandlerGroups(groups map[string][]stats.Handler) {
	for key, handlers := range groups {
		app.clientStatsHandlerGroups[key] = append(app.clientStatsHandlerGroups[key], handlers...)
	}
}

func (app *Application) getClientInterceptors(keys []string) []grpc.UnaryClientInterceptor {
	var result []grpc.UnaryClientInterceptor
	for _, key := range keys {
		if group, ok := app.unaryClientInterceptorGroups[key]; ok {
			result = append(result, group...)
		}
	}
	return result
}

func (app *Application) getClientStatsHandlers(keys []string) []stats.Handler {
	var result []stats.Handler
	for _, key := range keys {
		if group, ok := app.clientStatsHandlerGroups[key]; ok {
			result = append(result, group...)
		}
	}
	return result
}
