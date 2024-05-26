package console

import (
	"fmt"
	"os"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

type MigrateMakeCommand struct {
	config config.Config
}

func NewMigrateMakeCommand(config config.Config) *MigrateMakeCommand {
	return &MigrateMakeCommand{config: config}
}

// Signature The name and signature of the console command.
func (receiver *MigrateMakeCommand) Signature() string {
	return "queue:failed-table"
}

// Description The console command description.
func (receiver *MigrateMakeCommand) Description() string {
	return "Create a migration for the failed queue jobs database table"
}

// Extend The console command extend.
func (receiver *MigrateMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "queue",
	}
}

// Handle Execute the console command.
func (receiver *MigrateMakeCommand) Handle(ctx console.Context) error {
	up, down := receiver.getStub()
	if err := file.Create(receiver.getPath("up"), up); err != nil {
		return err
	}
	if err := file.Create(receiver.getPath("down"), down); err != nil {
		return err
	}

	color.Green().Println("Created Migration: create_failed_jobs_table")
	return nil
}

func (receiver *MigrateMakeCommand) getStub() (string, string) {
	driver := receiver.config.GetString("database.connections." + receiver.config.GetString("database.default") + ".driver")
	switch orm.Driver(driver) {
	case orm.DriverPostgresql:
		return PostgresqlStubs{}.FailedJobsUp(), PostgresqlStubs{}.FailedJobsDown()
	case orm.DriverSqlite:
		return SqliteStubs{}.FailedJobsUp(), SqliteStubs{}.FailedJobsDown()
	case orm.DriverSqlserver:
		return SqlserverStubs{}.FailedJobsUp(), SqlserverStubs{}.FailedJobsDown()
	default:
		return MysqlStubs{}.FailedJobsUp(), MysqlStubs{}.FailedJobsDown()
	}
}

// getPath Get the full path to the command.
func (receiver *MigrateMakeCommand) getPath(category string) string {
	pwd, _ := os.Getwd()

	return fmt.Sprintf("%s/database/migrations/%s_%s.%s.sql", pwd, carbon.Now().ToShortDateTimeString(), "create_failed_jobs_table", category)
}
