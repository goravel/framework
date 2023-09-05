package log

import (
	"errors"
	"fmt"
	"io"

	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/log/formatter"
	"github.com/goravel/framework/log/logger"
)

type Writer struct {
	code string

	// context
	context map[string]any
	domain  string

	hint     string
	instance *logrus.Entry
	message  string
	owner    any

	// http
	request  http.ContextRequest
	response http.ContextResponse

	// stacktrace
	stackEnabled bool
	stacktrace   map[string]any

	tags  []string
	trace string

	// user
	user any
}

func NewWriter(instance *logrus.Entry) log.Writer {
	return &Writer{
		code: "",

		// context
		context: map[string]any{},
		domain:  "",

		hint:     "",
		instance: instance,
		message:  "",
		owner:    nil,

		// http
		request:  nil,
		response: nil,

		// stacktrace
		stackEnabled: false,
		stacktrace:   nil,

		tags:  []string{},
		trace: "",

		// user
		user: nil,
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
	for k, v := range data {
		r.context[k] = v
	}

	return r
}

func (r *Writer) withStackTrace(message string) {
	erisNew := eris.New(message)
	r.message = erisNew.Error()
	format := eris.NewDefaultJSONFormat(eris.FormatOptions{
		InvertOutput: true,
		WithTrace:    true,
		InvertTrace:  true,
	})
	r.stacktrace = eris.ToCustomJSON(erisNew, format)
	r.stackEnabled = true
}

// ToMap returns a map representation of the error.
func (r *Writer) toMap() map[string]any {
	payload := map[string]any{}

	if message := r.message; message != "" {
		payload["message"] = message
	}

	if code := r.code; code != "" {
		payload["code"] = code
	}

	if domain := r.domain; domain != "" {
		payload["domain"] = domain
	}

	if tags := r.tags; len(tags) > 0 {
		payload["tags"] = tags
	}

	if context := r.context; len(context) > 0 {
		payload["context"] = context
	}

	if trace := r.trace; trace != "" {
		payload["trace"] = trace
	}

	if hint := r.hint; hint != "" {
		payload["hint"] = hint
	}

	if owner := r.owner; owner != nil {
		payload["owner"] = owner
	}

	if r.user != nil {
		payload["user"] = r.user
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

	return payload
}

func registerHook(config config.Config, instance *logrus.Logger, channel string) error {
	channelPath := "logging.channels." + channel
	driver := config.GetString(channelPath + ".driver")

	var hook logrus.Hook
	var err error
	switch driver {
	case log.StackDriver:
		for _, stackChannel := range config.Get(channelPath + ".channels").([]string) {
			if stackChannel == channel {
				return errors.New("stack drive can't include self channel")
			}

			if err := registerHook(config, instance, stackChannel); err != nil {
				return err
			}
		}

		return nil
	case log.SingleDriver:
		if !config.GetBool(channelPath + ".print") {
			instance.SetOutput(io.Discard)
		}

		logLogger := logger.NewSingle(config)
		hook, err = logLogger.Handle(channelPath)
		if err != nil {
			return err
		}
	case log.DailyDriver:
		if !config.GetBool(channelPath + ".print") {
			instance.SetOutput(io.Discard)
		}

		logLogger := logger.NewDaily(config)
		hook, err = logLogger.Handle(channelPath)
		if err != nil {
			return err
		}
	case log.CustomDriver:
		logLogger := config.Get(channelPath + ".via").(log.Logger)
		logHook, err := logLogger.Handle(channelPath)
		if err != nil {
			return err
		}

		hook = &Hook{logHook}
	default:
		return errors.New("Error logging channel: " + channel)
	}

	instance.SetFormatter(formatter.NewGeneral(config))

	instance.AddHook(hook)

	return nil
}

type Hook struct {
	instance log.Hook
}

func (h *Hook) Levels() []logrus.Level {
	levels := h.instance.Levels()
	var logrusLevels []logrus.Level
	for _, item := range levels {
		logrusLevels = append(logrusLevels, logrus.Level(item))
	}

	return logrusLevels
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	return h.instance.Fire(&Entry{
		ctx:     entry.Context,
		level:   log.Level(entry.Level),
		time:    entry.Time,
		message: entry.Message,
	})
}
