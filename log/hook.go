package log

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/log"
)

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
	e := &Entry{
		ctx:     entry.Context,
		data:    map[string]any(entry.Data),
		level:   log.Level(entry.Level),
		message: entry.Message,
		time:    entry.Time,
	}

	data := entry.Data
	if len(data) > 0 {
		root, err := cast.ToStringMapE(data["root"])
		if err != nil {
			return err
		}

		if code, err := cast.ToStringE(root["code"]); err == nil {
			e.code = code
		}
		if domain, err := cast.ToStringE(root["domain"]); err == nil {
			e.domain = domain
		}
		if hint, err := cast.ToStringE(root["hint"]); err == nil {
			e.hint = hint
		}

		e.owner = root["owner"]

		if request, err := cast.ToStringMapE(root["request"]); err == nil {
			e.request = request
		}
		if response, err := cast.ToStringMapE(root["response"]); err == nil {
			e.response = response
		}
		if stacktrace, err := cast.ToStringMapE(root["stacktrace"]); err == nil {
			e.stacktrace = stacktrace
		}
		if tags, err := cast.ToStringSliceE(root["tags"]); err == nil {
			e.tags = tags
		}

		e.user = root["user"]

		if with, err := cast.ToStringMapE(root["with"]); err == nil {
			e.with = with
		}
	}

	return h.instance.Fire(e)
}
