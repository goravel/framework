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
	telemetrylog "github.com/goravel/framework/telemetry/instrumentation/log"
)

type Application struct {
	log.Writer
	ctx               context.Context
	channels          []string
	channelToHandlers map[string][]slog.Handler
	config            config.Config
	json              foundation.Json
}

func NewApplication(ctx context.Context, channels []string, config config.Config, json foundation.Json) (*Application, error) {
	var handlers []slog.Handler

	if len(channels) == 0 && config != nil {
		if channel := config.GetString("logging.default"); channel != "" {
			channels = append(channels, channel)
		}
	}

	channelToHandlers := make(map[string][]slog.Handler)
	for _, channel := range channels {
		channelHandlers, err := getHandlers(config, json, channel)
		if err != nil {
			return nil, err
		}

		handlers = append(handlers, channelHandlers...)
		channelToHandlers[channel] = channelHandlers
	}

	slogLogger := slog.New(slogmulti.Fanout(handlers...))

	return &Application{
		ctx:               ctx,
		channels:          channels,
		config:            config,
		json:              json,
		channelToHandlers: channelToHandlers,
		Writer:            NewWriter(ctx, slogLogger),
	}, nil
}

func (r *Application) WithContext(ctx context.Context) log.Log {
	if httpCtx, ok := ctx.(http.Context); ok {
		ctx = httpCtx.Context()
	}

	var handles []slog.Handler
	for _, handlers := range r.channelToHandlers {
		handles = append(handles, handlers...)
	}

	copy := *r
	copy.ctx = ctx
	copy.Writer = NewWriter(ctx, slog.New(slogmulti.Fanout(handles...)))

	return &copy
}

func (r *Application) Channel(channel string) log.Log {
	if channel == "" {
		return r
	}

	handles, exist := r.channelToHandlers[channel]
	if !exist {
		panic(errors.LogChannelNotFound.Args(channel))
	}

	copy := *r
	copy.Writer = NewWriter(r.ctx, slog.New(slogmulti.Fanout(handles...)))

	return &copy
}

func (r *Application) Stack(channels []string) log.Log {
	if len(channels) == 0 {
		return r
	}

	var handles []slog.Handler
	for _, channel := range channels {
		handlers, exist := r.channelToHandlers[channel]
		if !exist {
			panic(errors.LogChannelNotFound.Args(channel))
		}

		handles = append(handles, handlers...)
	}

	copy := *r
	copy.Writer = NewWriter(r.ctx, slog.New(slogmulti.Fanout(handles...)))

	return &copy
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
			level := logger.GetLevelFromString(config.GetString(channelPath + ".level"))
			formatter := config.GetString(channelPath+".formatter", logger.FormatterText)
			handlers = append(handlers, HandlerToSlogHandler(logger.NewConsoleHandler(config, json, level, formatter)))
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
			level := logger.GetLevelFromString(config.GetString(channelPath + ".level"))
			formatter := config.GetString(channelPath+".formatter", logger.FormatterText)
			handlers = append(handlers, HandlerToSlogHandler(logger.NewConsoleHandler(config, json, level, formatter)))
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
			level := logger.GetLevelFromString(config.GetString(channelPath + ".level"))
			formatter := config.GetString(channelPath+".formatter", logger.FormatterText)
			handlers = append(handlers, HandlerToSlogHandler(logger.NewConsoleHandler(config, json, level, formatter)))
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
