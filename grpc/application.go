package grpc

import (
	"net"

	"google.golang.org/grpc"
)

type Application struct {
	server *grpc.Server
}

func (app *Application) Init() *Application {
	app.server = grpc.NewServer()

	return app
}

func (app *Application) Server() *grpc.Server {
	return app.server
}

func (app *Application) SetServer(server *grpc.Server) {
	app.server = server
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
