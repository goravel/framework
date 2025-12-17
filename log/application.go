package log

import (
	"context"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/log/logger"
	"github.com/goravel/framework/support/color"
	telemetrylog "github.com/goravel/framework/telemetry/instrumentation/log"
)

type Application struct {
	log.Writer
	config   config.Config
	json     foundation.Json
	handlers []slog.Handler
}

func NewApplication(config config.Config, json foundation.Json) (*Application, error) {
	var handlers []slog.Handler

	if config != nil {
		if channel := config.GetString("logging.default"); channel != "" {
			channelHandlers, err := getHandlers(config, json, channel)
			if err != nil {
				return nil, err
			}
			handlers = append(handlers, channelHandlers...)
		}
	}

	slogLogger := slog.New(slogmulti.Fanout(handlers...))

	return &Application{
		config:   config,
		json:     json,
		handlers: handlers,
		Writer:   NewWriter(slogLogger, context.Background()),
	}, nil
}

func (r *Application) WithContext(ctx context.Context) log.Writer {
	if httpCtx, ok := ctx.(http.Context); ok {
		ctx = httpCtx.Context()
	}

	slogLogger := slog.New(slogmulti.Fanout(r.handlers...))
	return NewWriter(slogLogger, ctx)
}

func (r *Application) Channel(channel string) log.Writer {
	if channel == "" || r.config == nil {
		return r.Writer
	}

	handlers, err := getHandlers(r.config, r.json, channel)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	slogLogger := slog.New(slogmulti.Fanout(handlers...))
	return NewWriter(slogLogger, context.Background())
}

func (r *Application) Stack(channels []string) log.Writer {
	if r.config == nil || len(channels) == 0 {
		return r.Writer
	}

	var handlers []slog.Handler
	for _, channel := range channels {
		if channel == "" {
			continue
		}

		channelHandlers, err := getHandlers(r.config, r.json, channel)
		if err != nil {
			color.Errorln(err)
			return nil
		}
		handlers = append(handlers, channelHandlers...)
	}

	slogLogger := slog.New(slogmulti.Fanout(handlers...))
	return NewWriter(slogLogger, context.Background())
}

// getHandlers returns slog log handlers for the specified channel.
func getHandlers(config config.Config, json foundation.Json, channel string) ([]slog.Handler, error) {
	channelPath := "logging.channels." + channel
	driver := config.GetString(channelPath + ".driver")

	switch driver {
	case log.DriverStack:
		var handlers []slog.Handler
		for _, stackChannel := range config.Get(channelPath + ".channels").([]string) {
			if stackChannel == channel {
				return nil, errors.LogDriverCircularReference.Args("stack")
			}

			channelHandlers, err := getHandlers(config, json, stackChannel)
			if err != nil {
				return nil, err
			}
			handlers = append(handlers, channelHandlers...)
		}
		return handlers, nil

	case log.DriverSingle:
		logLogger := logger.NewSingle(config, json)
		handler, err := logLogger.Handle(channelPath)
		if err != nil {
			return nil, err
		}

		handlers := []slog.Handler{HandlerToSlogHandler(handler)}
		if config.GetBool(channelPath + ".print") {
			handlers = append(handlers, HandlerToSlogHandler(logger.NewConsoleHandler(config, json)))
		}
		return handlers, nil

	case log.DriverDaily:
		logLogger := logger.NewDaily(config, json)
		handler, err := logLogger.Handle(channelPath)
		if err != nil {
			return nil, err
		}

		handlers := []slog.Handler{HandlerToSlogHandler(handler)}
		if config.GetBool(channelPath + ".print") {
			handlers = append(handlers, HandlerToSlogHandler(logger.NewConsoleHandler(config, json)))
		}
		return handlers, nil

	case log.DriverOtel:
		logLogger := telemetrylog.NewTelemetryChannel()
		handler, err := logLogger.Handle(channelPath)
		if err != nil {
			return nil, err
		}

		handlers := []slog.Handler{HandlerToSlogHandler(handler)}
		if config.GetBool(channelPath + ".print") {
			handlers = append(handlers, HandlerToSlogHandler(logger.NewConsoleHandler(config, json)))
		}
		return handlers, nil

	case log.DriverCustom:
		logLogger := config.Get(channelPath + ".via").(log.Logger)
		handler, err := logLogger.Handle(channelPath)
		if err != nil {
			return nil, err
		}
		return []slog.Handler{HandlerToSlogHandler(handler)}, nil

	default:
		return nil, errors.LogDriverNotSupported.Args(channel)
	}
}
