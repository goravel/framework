package log

import (
	"context"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"

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
	var handlers []slog.Handler
	
	if config != nil {
		if channel := config.GetString("logging.default"); channel != "" {
			channelHandlers, err := createHandlers(config, json, channel)
			if err != nil {
				return nil, err
			}
			handlers = append(handlers, channelHandlers...)
		}
	}
	
	// Use slog-multi to combine handlers
	var handler slog.Handler
	if len(handlers) == 0 {
		handler = NewDiscardHandler()
	} else if len(handlers) == 1 {
		handler = handlers[0]
	} else {
		handler = slogmulti.Fanout(handlers...)
	}
	
	instance := slog.New(handler)

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

	channelHandlers, err := createHandlers(r.config, r.json, channel)
	if err != nil {
		color.Errorln(err)
		return nil
	}
	
	var handler slog.Handler
	if len(channelHandlers) == 1 {
		handler = channelHandlers[0]
	} else {
		handler = slogmulti.Fanout(channelHandlers...)
	}
	
	instance := slog.New(handler)

	return NewWriter(instance, context.Background())
}

func (r *Application) Stack(channels []string) log.Writer {
	if r.config == nil || len(channels) == 0 {
		return r.Writer
	}

	var allHandlers []slog.Handler
	
	for _, channel := range channels {
		if channel == "" {
			continue
		}

		channelHandlers, err := createHandlers(r.config, r.json, channel)
		if err != nil {
			color.Errorln(err)
			return nil
		}
		allHandlers = append(allHandlers, channelHandlers...)
	}
	
	if len(allHandlers) == 0 {
		return r.Writer
	}
	
	handler := slogmulti.Fanout(allHandlers...)
	instance := slog.New(handler)

	return NewWriter(instance, context.Background())
}
