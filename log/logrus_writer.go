package log

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/log/logger"
)

type Writer struct {
	instance *logrus.Entry

	message  string
	code     string
	time     time.Time
	duration time.Duration

	// context
	domain  string
	tags    []string
	context map[string]any

	trace string
	span  string

	hint  string
	owner string

	// user
	userID   string
	userData map[string]any

	// http
	request  http.Request
	response http.Response

	// stacktrace
	stackEnabled bool
	stacktrace   string
}

func NewWriter(instance *logrus.Entry) log.Writer {
	return &Writer{
		instance: instance,

		message:  "",
		code:     "",
		time:     time.Now(),
		duration: 0,

		// context
		domain:  "",
		tags:    []string{},
		context: map[string]any{},

		trace: "",
		span:  "",

		hint:  "",
		owner: "",

		// user
		userID:   "",
		userData: map[string]any{},

		// http
		request:  nil,
		response: nil,

		// stacktrace
		stackEnabled: false,
		stacktrace:   "",
	}
}

func (r *Writer) Debug(args ...any) {
	r.instance.WithField("debug", r.toMap()).Debug(args...)
}

func (r *Writer) Debugf(format string, args ...any) {
	r.instance.WithField("debug", r.toMap()).Debugf(format, args...)
}

func (r *Writer) Info(args ...any) {
	r.instance.WithField("info", r.toMap()).Info(args...)
}

func (r *Writer) Infof(format string, args ...any) {
	r.instance.WithField("info", r.toMap()).Infof(format, args...)
}

func (r *Writer) Warning(args ...any) {
	r.instance.WithField("warning", r.toMap()).Warning(args...)
}

func (r *Writer) Warningf(format string, args ...any) {
	r.instance.WithField("warning", r.toMap()).Warningf(format, args...)
}

func (r *Writer) Error(args ...any) {
	r.WithStackTrace(fmt.Sprint(args...))
	r.instance.WithField("error", r.toMap()).Error(args...)
}

func (r *Writer) Errorf(format string, args ...any) {
	r.WithStackTrace(fmt.Sprintf(format, args...))
	r.instance.WithField("error", r.toMap()).Errorf(format, args...)
}

func (r *Writer) Fatal(args ...any) {
	r.WithStackTrace(fmt.Sprint(args...))
	r.instance.WithField("fatal", r.toMap()).Fatal(args...)
}

func (r *Writer) Fatalf(format string, args ...any) {
	r.WithStackTrace(fmt.Sprintf(format, args...))
	r.instance.WithField("fatal", r.toMap()).Fatalf(format, args...)
}

func (r *Writer) Panic(args ...any) {
	r.WithStackTrace(fmt.Sprint(args...))
	r.instance.WithField("panic", r.toMap()).Panic(args...)
}

func (r *Writer) Panicf(format string, args ...any) {
	r.WithStackTrace(fmt.Sprintf(format, args...))
	r.instance.WithField("panic", r.toMap()).Panicf(format, args...)
}

func (r *Writer) User(userID string, userData ...map[string]any) log.Writer {
	r.userID = userID
	if len(userData) > 0 {
		r.userData = userData[0]
	}
	return r
}

func (r *Writer) Owner(ownerID string) log.Writer {
	r.owner = ownerID

	return r
}

func (r *Writer) Hint(hint string) log.Writer {
	r.hint = hint

	return r
}

func (r *Writer) Trace(trace string) log.Writer {
	r.trace = trace

	return r
}

func (r *Writer) Code(code string) log.Writer {
	r.code = code
	return r
}

func (r *Writer) With(data map[string]any) log.Writer {
	for k, v := range data {
		r.context[k] = v
	}

	return r
}

func (r *Writer) Tags(tags []string) log.Writer {
	r.tags = append(r.tags, tags...)

	return r
}

func (r *Writer) Request(req http.Request) log.Writer {
	r.request = req

	return r
}

func (r *Writer) Response(res http.Response) log.Writer {
	r.response = res

	return r
}

func (r *Writer) In(domain string) log.Writer {
	r.domain = domain

	return r
}

func (r *Writer) WithStackTrace(message string) {
	erisNew := eris.New(message)
	r.message = erisNew.Error()
	r.stacktrace = eris.ToString(erisNew, true)
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

	if t := r.time; t != (time.Time{}) {
		payload["time"] = t.UTC()
	}

	if duration := r.duration; duration != 0 {
		payload["duration"] = duration.String()
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

	if span := r.span; span != "" {
		payload["span"] = span
	}

	if hint := r.hint; hint != "" {
		payload["hint"] = hint
	}

	if owner := r.owner; owner != "" {
		payload["owner"] = owner
	}

	if r.userID != "" || len(r.userData) > 0 {
		user := make(map[string]any)
		for k, v := range r.userData {
			user[k] = v
		}
		if r.userID != "" {
			user["id"] = r.userID
		}
		payload["user"] = user
	}

	if req := r.request; req != nil {
		payload["request"] = req
	}

	if res := r.response; res != nil {
		payload["response"] = res
	}

	if stacktrace := r.stacktrace; stacktrace != "" || r.stackEnabled {
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
