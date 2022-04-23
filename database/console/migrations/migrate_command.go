package migrations

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/facades"
	"github.com/urfave/cli/v2"
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

//Flags Set flags, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#flags
func (receiver *MigrateCommand) Flags() []cli.Flag {
	var flags []cli.Flag

	return flags
}

//Subcommands Set Subcommands, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#subcommands
func (receiver *MigrateCommand) Subcommands() []*cli.Command {
	var subcommands []*cli.Command

	return subcommands
}

//Handle Execute the console command.
func (receiver *MigrateCommand) Handle(c *cli.Context) error {
	config := support.Helpers{}.GetDatabaseConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		config["username"], config["password"], config["host"], config["port"], config["database"], config["charset"], true, "Local")

	flag.Parse()
	var migrationDir = flag.String("migration.files", "./database/migrations", "Directory where the migration files are located ?")
	var mysqlDSN = flag.String("mysql.dsn", dsn, "Mysql DSN")

	db, err := sql.Open("mysql", *mysqlDSN)
	if err != nil {
		return errors.New("Could not connect to database: " + err.Error())
	}

	if err := db.Ping(); err != nil {
		return errors.New("Could not ping to database: " + err.Error())
	}

	// Run migrations
	driver, err := mysql.WithInstance(db, &mysql.Config{
		MigrationsTable: facades.Config.GetString("database.migrations"),
	})
	if err != nil {
		return errors.New("Could not start sql migration: " + err.Error())
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", *migrationDir),
		"mysql", driver)

	if err != nil {
		return errors.New("Migration init failed: " + err.Error())
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return errors.New("Migration failed: " + err.Error())
	}

	fmt.Println("Migration success")

	return nil
}
