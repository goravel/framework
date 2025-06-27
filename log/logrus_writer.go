package log

import (
	"fmt"
	"io"
	"maps"
	"os"

	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/log/formatter"
	"github.com/goravel/framework/log/logger"
)

func NewLogrus() *logrus.Logger {
	instance := logrus.New()
	instance.SetLevel(logrus.DebugLevel)
	instance.SetOutput(io.Discard)

	return instance
}

type Writer struct {
	owner        any
	request      http.ContextRequest
	response     http.ContextResponse
	user         any
	instance     *logrus.Entry
	stacktrace   map[string]any
	with         map[string]any
	code         string
	domain       string
	hint         string
	message      string
	tags         []string
	stackEnabled bool
}

func NewWriter(instance *logrus.Entry) log.Writer {
	return &Writer{
		code:         "",
		domain:       "",
		hint:         "",
		instance:     instance,
		message:      "",
		owner:        nil,
		request:      nil,
		response:     nil,
		stackEnabled: false,
		stacktrace:   nil,
		tags:         []string{},
		user:         nil,
		with:         map[string]any{},
	}
}

func (r *Writer) Debug(args ...any) {
	r.instance.WithField("root", r.toMap()).Debug(args...)
}

func (r *Writer) Debugf(format string, args ...any) {
	r.instance.WithField("root", r.toMap()).Debugf(format, args...)
}

func (r *Writer) Info(args ...any) {
	r.instance.WithField("root", r.toMap()).Info(args...)
}

func (r *Writer) Infof(format string, args ...any) {
	r.instance.WithField("root", r.toMap()).Infof(format, args...)
}

func (r *Writer) Warning(args ...any) {
	r.instance.WithField("root", r.toMap()).Warning(args...)
}

func (r *Writer) Warningf(format string, args ...any) {
	r.instance.WithField("root", r.toMap()).Warningf(format, args...)
}

func (r *Writer) Error(args ...any) {
	r.withStackTrace(fmt.Sprint(args...))
	r.instance.WithField("root", r.toMap()).Error(args...)
}

func (r *Writer) Errorf(format string, args ...any) {
	r.withStackTrace(fmt.Sprintf(format, args...))
	r.instance.WithField("root", r.toMap()).Errorf(format, args...)
}

func (r *Writer) Fatal(args ...any) {
	r.withStackTrace(fmt.Sprint(args...))
	r.instance.WithField("root", r.toMap()).Fatal(args...)
}

func (r *Writer) Fatalf(format string, args ...any) {
	r.withStackTrace(fmt.Sprintf(format, args...))
	r.instance.WithField("root", r.toMap()).Fatalf(format, args...)
}

func (r *Writer) Panic(args ...any) {
	r.withStackTrace(fmt.Sprint(args...))
	r.instance.WithField("root", r.toMap()).Panic(args...)
}

func (r *Writer) Panicf(format string, args ...any) {
	r.withStackTrace(fmt.Sprintf(format, args...))
	r.instance.WithField("root", r.toMap()).Panicf(format, args...)
}

// Code set a code or slug that describes the error.
// Error messages are intended to be read by humans, but such code is expected to
// be read by machines and even transported over different services.
func (r *Writer) Code(code string) log.Writer {
	r.code = code
	return r
}

// Hint set a hint for faster debugging.
func (r *Writer) Hint(hint string) log.Writer {
	r.hint = hint

	return r
}

// In sets the feature category or domain in which the log entry is relevant.
func (r *Writer) In(domain string) log.Writer {
	r.domain = domain

	return r
}

// Owner set the name/email of the colleague/team responsible for handling this error.
// Useful for alerting purpose.
func (r *Writer) Owner(owner any) log.Writer {
	r.owner = owner

	return r
}

// Request supplies a http.Request.
func (r *Writer) Request(req http.ContextRequest) log.Writer {
	r.request = req

	return r
}

// Response supplies a http.Response.
func (r *Writer) Response(res http.ContextResponse) log.Writer {
	r.response = res

	return r
}

// Tags add multiple tags, describing the feature returning an error.
func (r *Writer) Tags(tags ...string) log.Writer {
	r.tags = append(r.tags, tags...)

	return r
}

// User sets the user associated with the log entry.
func (r *Writer) User(user any) log.Writer {
	r.user = user
	return r
}

// With adds key-value pairs to the context of the log entry
func (r *Writer) With(data map[string]any) log.Writer {
	maps.Copy(r.with, data)

	return r
}

// WithTrace adds a stack trace to the log entry.
func (r *Writer) WithTrace() log.Writer {
	r.withStackTrace("")
	return r
}

func (r *Writer) withStackTrace(message string) {
	erisNew := eris.New(message)
	if erisNew == nil {
		return
	}

	r.message = erisNew.Error()
	format := eris.NewDefaultJSONFormat(eris.FormatOptions{
		InvertOutput: true,
		WithTrace:    true,
		InvertTrace:  true,
	})
	r.stacktrace = eris.ToCustomJSON(erisNew, format)
	r.stackEnabled = true
}

// resetAll resets all properties of the log.Writer for a new log entry.
func (r *Writer) resetAll() {
	r.code = ""
	r.domain = ""
	r.hint = ""
	r.message = ""
	r.owner = nil
	r.request = nil
	r.response = nil
	r.stacktrace = nil
	r.stackEnabled = false
	r.tags = []string{}
	r.user = nil
	r.with = map[string]any{}
}

// toMap returns a map representation of the error.
func (r *Writer) toMap() map[string]any {
	payload := map[string]any{}

	if code := r.code; code != "" {
		payload["code"] = code
	}
	if ctx := r.instance.Context; ctx != nil {
		values := make(map[any]any)
		getContextValues(ctx, values)
		if len(values) > 0 {
			payload["context"] = values
		}
	}
	if domain := r.domain; domain != "" {
		payload["domain"] = domain
	}
	if hint := r.hint; hint != "" {
		payload["hint"] = hint
	}
	if message := r.message; message != "" {
		payload["message"] = message
	}
	if owner := r.owner; owner != nil {
		payload["owner"] = owner
	}
	if req := r.request; req != nil {
		payload["request"] = map[string]any{
			"method": req.Method(),
			"uri":    req.FullUrl(),
			"header": req.Headers(),
			"body":   req.All(),
		}
	}
	if res := r.response; res != nil {
		payload["response"] = map[string]any{
			"status": res.Origin().Status(),
			"header": res.Origin().Header(),
			"body":   res.Origin().Body(),
			"size":   res.Origin().Size(),
		}
	}
	if stacktrace := r.stacktrace; stacktrace != nil || r.stackEnabled {
		payload["stacktrace"] = stacktrace
	}
	if tags := r.tags; len(tags) > 0 {
		payload["tags"] = tags
	}
	if r.user != nil {
		payload["user"] = r.user
	}
	if with := r.with; len(with) > 0 {
		payload["with"] = with
	}

	// reset all properties for a new log entry
	r.resetAll()

	return payload
}

func registerHook(config config.Config, json foundation.Json, instance *logrus.Logger, channel string) error {
	var (
		hook logrus.Hook
		err  error

		channelPath = "logging.channels." + channel
		driver      = config.GetString(channelPath + ".driver")
	)

	switch driver {
	case log.StackDriver:
		for _, stackChannel := range config.Get(channelPath + ".channels").([]string) {
			if stackChannel == channel {
				return errors.LogDriverCircularReference.Args("stack")
			}

			if err := registerHook(config, json, instance, stackChannel); err != nil {
				return err
			}
		}

		return nil
	case log.SingleDriver:
		logLogger := logger.NewSingle(config, json)
		hook, err = logLogger.Handle(channelPath)
		if err != nil {
			return err
		}

		if config.GetBool(channelPath + ".print") {
			instance.SetOutput(os.Stdout)
			instance.SetFormatter(formatter.NewGeneral(config, json))
		}
	case log.DailyDriver:
		logLogger := logger.NewDaily(config, json)
		hook, err = logLogger.Handle(channelPath)
		if err != nil {
			return err
		}

		if config.GetBool(channelPath + ".print") {
			instance.SetOutput(os.Stdout)
			instance.SetFormatter(formatter.NewGeneral(config, json))
		}
	case log.CustomDriver:
		logLogger := config.Get(channelPath + ".via").(log.Logger)
		logHook, err := logLogger.Handle(channelPath)
		if err != nil {
			return err
		}

		hook = &Hook{logHook}
	default:
		return errors.LogDriverNotSupported.Args(channel)
	}

	instance.AddHook(hook)

	return nil
}
