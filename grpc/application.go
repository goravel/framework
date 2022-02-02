package grpc

import (
	g "google.golang.org/grpc"
	"net"
)

type Application struct {
	server *g.Server
}

func (app *Application) Init() *Application {
	app.server = g.NewServer()

	return app
}

func (app *Application) Server() *g.Server {
	return app.server
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
