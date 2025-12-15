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
	logger *slog.Logger
	config config.Config
	json   foundation.Json
}

func NewApplication(config config.Config, json foundation.Json) (*Application, error) {
	var handlers []slog.Handler

	if config != nil {
		if channel := config.GetString("logging.default"); channel != "" {
			h, err := getHandlers(config, json, channel)
			if err != nil {
				return nil, err
			}
			handlers = append(handlers, h...)
		}
	}

	var handler slog.Handler
	if len(handlers) == 0 {
		// Default handler that discards all logs
		handler = &discardHandler{}
	} else if len(handlers) == 1 {
		handler = handlers[0]
	} else {
		handler = slogmulti.Fanout(handlers...)
	}

	logger := slog.New(handler)

	return &Application{
		logger: logger,
		Writer: NewWriter(logger.With(), context.Background()),
		config: config,
		json:   json,
	}, nil
}

func (r *Application) WithContext(ctx context.Context) log.Writer {
	if httpCtx, ok := ctx.(http.Context); ok {
		return NewWriter(r.logger, httpCtx.Context())
	}

	return NewWriter(r.logger, ctx)
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

	var handler slog.Handler
	if len(handlers) == 0 {
		handler = &discardHandler{}
	} else if len(handlers) == 1 {
		handler = handlers[0]
	} else {
		handler = slogmulti.Fanout(handlers...)
	}

	logger := slog.New(handler)
	return NewWriter(logger, context.Background())
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

		h, err := getHandlers(r.config, r.json, channel)
		if err != nil {
			color.Errorln(err)
			return nil
		}
		handlers = append(handlers, h...)
	}

	var handler slog.Handler
	if len(handlers) == 0 {
		handler = &discardHandler{}
	} else if len(handlers) == 1 {
		handler = handlers[0]
	} else {
		handler = slogmulti.Fanout(handlers...)
	}

	logger := slog.New(handler)
	return NewWriter(logger, context.Background())
}

// discardHandler is a handler that discards all logs
type discardHandler struct{}

func (h *discardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (h *discardHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (h *discardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *discardHandler) WithGroup(_ string) slog.Handler {
	return h
}
