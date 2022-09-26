package console

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gookit/color"
	"github.com/urfave/cli/v2"

	"github.com/goravel/framework/contracts/console"
)

type MigrateCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *MigrateCommand) Signature() string {
	return "migrate"
}

//Description The console command description.
func (receiver *MigrateCommand) Description() string {
	return "Run the database migrations"
}

//Extend The console command extend.
func (receiver *MigrateCommand) Extend() console.CommandExtend {
	return console.CommandExtend{
		Category: "migrate",
	}
}

//Handle Execute the console command.
func (receiver *MigrateCommand) Handle(c *cli.Context) error {
	m, err := getMigrate()
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		color.Redln("Migration failed:", err.Error())

		return nil
	}

	color.Greenln("Migration success")

	return nil
}
