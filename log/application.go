package log

import (
	"context"
	"log/slog"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/color"
)

type Application struct {
	log.Writer
	instance *slog.Logger
	config   config.Config
	json     foundation.Json
}

func NewApplication(config config.Config, json foundation.Json) (*Application, error) {
	multiHandler := NewMultiHandler()
	instance := slog.New(multiHandler)
	
	if config != nil {
		if channel := config.GetString("logging.default"); channel != "" {
			if err := registerHandler(config, json, multiHandler, channel); err != nil {
				return nil, err
			}
		}
	}

	return &Application{
		instance: instance,
		Writer:   NewWriter(instance, context.Background()),
		config:   config,
		json:     json,
	}, nil
}

func (r *Application) WithContext(ctx context.Context) log.Writer {
	if httpCtx, ok := ctx.(http.Context); ok {
		return NewWriter(r.instance, httpCtx.Context())
	}

	return NewWriter(r.instance, ctx)
}

func (r *Application) Channel(channel string) log.Writer {
	if channel == "" || r.config == nil {
		return r.Writer
	}

	multiHandler := NewMultiHandler()
	instance := slog.New(multiHandler)
	
	if err := registerHandler(r.config, r.json, multiHandler, channel); err != nil {
		color.Errorln(err)
		return nil
	}

	return NewWriter(instance, context.Background())
}

func (r *Application) Stack(channels []string) log.Writer {
	if r.config == nil || len(channels) == 0 {
		return r.Writer
	}

	multiHandler := NewMultiHandler()
	instance := slog.New(multiHandler)
	
	for _, channel := range channels {
		if channel == "" {
			continue
		}

		if err := registerHandler(r.config, r.json, multiHandler, channel); err != nil {
			color.Errorln(err)
			return nil
		}
	}

	return NewWriter(instance, context.Background())
}
