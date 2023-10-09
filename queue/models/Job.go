package models

import (
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/queue"
	"github.com/goravel/framework/support/carbon"
)

type Job struct {
	ID            uint                 `gorm:"primaryKey"`
	Queue         string               `gorm:"type:text;not null"`
	Job           string               `gorm:"type:text;not null"`
	Arg           []contractsqueue.Arg `gorm:"type:json;not null;serializer:json"`
	Attempts      uint                 `gorm:"type:bigint;not null;default:0"`
	MaxTries      *uint                `gorm:"type:bigint;default:null;default:0"`
	MaxExceptions *uint                `gorm:"type:bigint;default:null;default:0"`
	Exception     *string              `gorm:"type:text;default:null"`
	Backoff       uint                 `gorm:"type:bigint;not null;default:0"`     // A number of seconds to wait before retrying the job.
	Timeout       *uint                `gorm:"type:bigint;default:null;default:0"` // The number of seconds the job can run.
	TimeoutAt     *carbon.DateTime     `gorm:"column:timeout_at"`
	ReservedAt    *carbon.DateTime     `gorm:"column:reserved_at"`
	AvailableAt   *carbon.DateTime     `gorm:"column:available_at"`
	CreatedAt     *carbon.DateTime     `gorm:"column:created_at"`
	FailedAt      *carbon.DateTime     `gorm:"column:failed_at"`
}

func (j Job) Signature() string {
	return j.Job
}

func (j Job) Args() []contractsqueue.Arg {
	return j.Arg
}

func (j Job) Handle(args ...any) error {
	return queue.Call(j.Job, args...)
}
