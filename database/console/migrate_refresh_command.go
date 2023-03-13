package console

import (

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type MigrateRefreshCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *MigrateRefreshCommand) Signature() string {
	return "migrate:refresh"
}

//Description The console command description.
func (receiver *MigrateRefreshCommand) Description() string {
	return "Reset and re-run all migrations"
}

//Extend The console command extend.
func (receiver *MigrateRefreshCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
	}
}

//Handle Execute the console command.
func (receiver *MigrateRefreshCommand) Handle(ctx console.Context) error {
	
	MigrateResetCommand := &MigrateResetCommand{}
	MigrateResetCommand.Handle(ctx)

	MigrateCommand := &MigrateCommand{}
	MigrateCommand.Handle(ctx)
	
	return nil
}
