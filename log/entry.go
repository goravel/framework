package log

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/log"
)

type Entry struct {
	ctx        context.Context
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

func (r *Entry) Context() context.Context {
	return r.ctx
}

func (r *Entry) Level() log.Level {
	return r.level
}

func (r *Entry) Time() time.Time {
	return r.time
}

func (r *Entry) Message() string {
	return r.message
}

func (r *Entry) Code() string {
	return r.code
}

func (r *Entry) With() map[string]any {
	return r.with
}

func (r *Entry) User() any {
	return r.user
}

func (r *Entry) Tags() []string {
	return r.tags
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

func (r *Entry) Trace() map[string]any {
	return r.stacktrace
}
