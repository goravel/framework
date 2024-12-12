package log

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/log"
)

type Entry struct {
	ctx        context.Context
	data       log.Data
	level      log.Level
	time       time.Time
	message    string
	code       string
	user       any
	tags       []string
	owner      any
	request    map[string]any
	response   map[string]any
	with       map[string]any
	stacktrace map[string]any
}

func (r *Entry) Code() string {
	return r.code
}

func (r *Entry) Context() context.Context {
	return r.ctx
}

func (r *Entry) Data() log.Data {
	return r.data
}

func (r *Entry) Level() log.Level {
	return r.level
}

func (r *Entry) Message() string {
	return r.message
}

func (r *Entry) Owner() any {
	return r.owner
}

func (r *Entry) Request() map[string]any {
	return r.request
}

func (r *Entry) Response() map[string]any {
	return r.response
}

func (r *Entry) Tags() []string {
	return r.tags
}

func (r *Entry) Time() time.Time {
	return r.time
}

func (r *Entry) Trace() map[string]any {
	return r.stacktrace
}

func (r *Entry) User() any {
	return r.user
}

func (r *Entry) With() map[string]any {
	return r.with
}
