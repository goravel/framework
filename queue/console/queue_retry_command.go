package console

import (
	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/queue/models"
	"github.com/goravel/framework/queue/utils"
	"github.com/goravel/framework/support/carbon"
)

type QueueRetryCommand struct {
	db             db.DB
	failedJobQuery db.Query
	json           foundation.Json
	queue          contractsqueue.Queue
}

func NewQueueRetryCommand(config config.Config, db db.DB, queue contractsqueue.Queue, json foundation.Json) *QueueRetryCommand {
	failedDatabase := config.GetString("queue.failed.database")
	failedTable := config.GetString("queue.failed.table")

	return &QueueRetryCommand{
		db:             db,
		failedJobQuery: db.Connection(failedDatabase).Table(failedTable),
		json:           json,
		queue:          queue,
	}
}

// Signature The name and signature of the console command.
func (r *QueueRetryCommand) Signature() string {
	return "queue:retry"
}

// Description The console command description.
func (r *QueueRetryCommand) Description() string {
	return "Retry a failed queue job"
}

// Extend The console command extend.
func (r *QueueRetryCommand) Extend() command.Extend {
	return command.Extend{
		Category: "queue",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "queue",
				Aliases: []string{"q"},
				Usage:   "Retry all of the failed jobs for the specified queue",
			},
		},
	}
}

// Handle Execute the console command.
func (r *QueueRetryCommand) Handle(ctx console.Context) error {
	ids, err := r.getJobIDs(ctx)
	if err != nil {
		ctx.Error(err.Error())
		return err
	}

	if len(ids) == 0 {
		ctx.Info(errors.QueueNoRetryableJobsFound.Error())
		return nil
	}

	ctx.Info(errors.QueuePushingFailedJob.Error())
	ctx.Line("")

	for _, id := range ids {
		now := carbon.Now()

		var failedJob models.FailedJob
		if err := r.failedJobQuery.Where("id", id).First(&failedJob); err != nil {
			return err
		}

		if failedJob.ID == 0 {
			ctx.Error(errors.QueueFailedJobNotFound.Args(id).Error())
			continue
		}

		if err := r.retryJob(failedJob); err != nil {
			ctx.Error(errors.QueueFailedToRetryJob.Args(failedJob, err).Error())
			continue
		}

		if _, err := r.failedJobQuery.Where("id", id).Delete(); err != nil {
			ctx.Error(errors.QueueFailedToDeleteFailedJob.Args(failedJob, err).Error())
			continue
		}

		r.printSuccess(ctx, failedJob.UUID, now.DiffAbsInDuration().String())
	}

	ctx.Line("")

	return nil
}

func (r *QueueRetryCommand) getJobIDs(ctx console.Context) ([]string, error) {
	uuids := ctx.Arguments()

	var ids []string

	if len(uuids) == 1 && uuids[0] == "all" {
		if err := r.failedJobQuery.Pluck("id", &ids); err != nil {
			return nil, err
		}

		return ids, nil
	}

	if queue := ctx.Option("queue"); queue != "" {
		if err := r.failedJobQuery.Where("queue", queue).Pluck("id", &ids); err != nil {
			return nil, err
		}

		return ids, nil
	}

	if len(uuids) == 0 {
		return nil, nil
	}

	if err := r.failedJobQuery.WhereIn("uuid", cast.ToSlice(uuids)).Pluck("id", &ids); err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *QueueRetryCommand) printSuccess(ctx console.Context, uuid, duration string) {
	status := "<fg=green;op=bold>DONE</>"
	first := uuid
	second := duration + " " + status

	ctx.TwoColumnDetail(first, second)
}

func (r *QueueRetryCommand) retryJob(failedJob models.FailedJob) error {
	connection, err := r.queue.Connection(failedJob.Connection)
	if err != nil {
		return err
	}

	task, err := utils.JsonToTask(failedJob.Payload, r.queue.GetJobStorer(), r.json)
	if err != nil {
		return err
	}

	return connection.Push(task, failedJob.Queue)
}
