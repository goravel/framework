package console

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
)

type QueueFailedCommand struct {
	failedJobQuery contractsdb.Query
	queue          contractsqueue.Queue
}

func NewQueueFailedCommand(config config.Config, db contractsdb.DB, queue contractsqueue.Queue) *QueueFailedCommand {
	failedDatabase := config.GetString("queue.failed.database")
	failedTable := config.GetString("queue.failed.table")

	var failedJobQuery contractsdb.Query
	if db != nil {
		failedJobQuery = db.Connection(failedDatabase).Table(failedTable)
	}

	return &QueueFailedCommand{
		failedJobQuery: failedJobQuery,
		queue:          queue,
	}
}

// Signature The name and signature of the console command.
func (r *QueueFailedCommand) Signature() string {
	return "queue:failed"
}

// Description The console command description.
func (r *QueueFailedCommand) Description() string {
	return "List all of the failed queue jobs"
}

// Extend The console command extend.
func (r *QueueFailedCommand) Extend() command.Extend {
	return command.Extend{
		Category: "queue",
	}
}

// Handle Execute the console command.
func (r *QueueFailedCommand) Handle(ctx console.Context) error {
	if r.failedJobQuery == nil {
		ctx.Error(errors.DBFacadeNotSet.Error())
		return nil
	}

	failedJobs, err := r.queue.Failer().All()
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if len(failedJobs) == 0 {
		ctx.Info(errors.QueueNoFailedJobsFound.Error())
		return nil
	}

	ctx.Info(errors.QueuePushingFailedJob.Error())
	ctx.Line("")

	for _, failedJob := range failedJobs {
		r.printJob(ctx, failedJob.UUID(), failedJob.Connection(), failedJob.Queue())
	}

	ctx.Line("")

	return nil
}

func (r *QueueFailedCommand) printJob(ctx console.Context, uuid, connection, queue string) {
	datetime := color.Gray().Sprint(carbon.Now().ToDateTimeString())
	status := connection + "@" + queue
	first := datetime + " " + uuid
	second := status

	ctx.TwoColumnDetail(first, second)
}
