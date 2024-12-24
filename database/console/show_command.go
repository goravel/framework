package console

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
)

type ShowCommand struct {
	config config.Config
	schema schema.Schema
}

type databaseInfo struct {
	Platform struct {
		Name            string
		Version         string
		openConnections int
	}
	Tables []schema.Table
	Views  []schema.View
}

func NewShowCommand(config config.Config, schema schema.Schema) *ShowCommand {
	return &ShowCommand{
		config: config,
		schema: schema,
	}
}

// Signature The name and signature of the console command.
func (receiver *ShowCommand) Signature() string {
	return "db:show"
}

// Description The console command description.
func (receiver *ShowCommand) Description() string {
	return "Display information about the given database"
}

// Extend The console command extend.
func (receiver *ShowCommand) Extend() command.Extend {
	return command.Extend{
		Category: "db",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:  "database",
				Usage: "The database connection",
			},
			&command.BoolFlag{
				Name:  "views",
				Usage: "Show the database views</>",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *ShowCommand) Handle(ctx console.Context) error {
	if got := ctx.Argument(0); len(got) > 0 {
		ctx.Error(fmt.Sprintf("No arguments expected for '%s' command, got '%s'.", receiver.Signature(), got))
	}
	receiver.schema = receiver.schema.Connection(ctx.Option("database"))
	//platform, version := receiver.platformAndVersion()
	//ctx.TwoColumnDetail(fmt.Sprintf("<fg=green;op=bold>%s</>", platform), version)

	return nil
}

func (receiver *ShowCommand) getDatabaseInfo() (info databaseInfo) {
	driver := receiver.schema.Orm().Query().Driver()
	switch driver {
	case database.DriverSqlite:
		info.Platform.Name = "SQLite"
	case database.DriverMysql:
		info.Platform.Name = "MySQL"
		if err := receiver.schema.Orm().Query().Raw("SELECT VERSION() as version;").Scan(&info.Platform); err == nil {

		}
	case database.DriverPostgres:
		info.Platform.Name = "PostgresSQL"

	case database.DriverSqlserver:
		info.Platform.Name = "SQL Server"
	default:
		info.Platform.Name = driver.String()
	}

	return
}
