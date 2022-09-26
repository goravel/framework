package console

import (
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gookit/color"
	"github.com/urfave/cli/v2"

	"github.com/goravel/framework/contracts/console"
)

type MigrateRollbackCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *MigrateRollbackCommand) Signature() string {
	return "migrate:rollback"
}

//Description The console command description.
func (receiver *MigrateRollbackCommand) Description() string {
	return "Rollback the database migrations"
}

//Extend The console command extend.
func (receiver *MigrateRollbackCommand) Extend() console.CommandExtend {
	return console.CommandExtend{
		Category: "migrate",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "step",
				Value: "1",
				Usage: "rollback steps",
			},
		},
	}
}

//Handle Execute the console command.
func (receiver *MigrateRollbackCommand) Handle(c *cli.Context) error {
	m, err := getMigrate()
	if err != nil {
		return err
	}

	stepString := "-" + c.String("step")
	step, err := strconv.Atoi(stepString)
	if err != nil {
		color.Redln("Migration failed: invalid step", c.String("step"))

		return nil
	}

	if err := m.Steps(step); err != nil && err != migrate.ErrNoChange && err != migrate.ErrNilVersion {
		switch err.(type) {
		case migrate.ErrShortLimit:
		default:
			color.Redln("Migration failed:", err.Error())

			return nil
		}
	}

	color.Greenln("Migration rollback success")

	return nil
}
