package log

import (
	"time"

	"github.com/goravel/framework/contracts/log"
)

type Entry struct {
	level   log.Level
	time    time.Time
	message string
}

func (r *Entry) GetLevel() log.Level {
	return r.level
}

func (r *Entry) GetTime() time.Time {
	return r.time
}

func (r *Entry) GetMessage() string {
	return r.message
}
