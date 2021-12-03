package migrations

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/goravel/framework/database/helpers"
	"github.com/goravel/framework/support/facades"
	"github.com/urfave/cli/v2"
	"log"
)

type MigrateCommand struct {
}

//Signature The name and signature of the console command.
func (receiver MigrateCommand) Signature() string {
	return "migrate"
}

//Description The console command description.
func (receiver MigrateCommand) Description() string {
	return "Run the database migrations"
}

//Flags Set flags, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#flags
func (receiver MigrateCommand) Flags() []cli.Flag {
	var flags []cli.Flag

	return flags
}

//Subcommands Set Subcommands, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#subcommands
func (receiver MigrateCommand) Subcommands() []*cli.Command {
	var subcommands []*cli.Command

	return subcommands
}

//Handle Execute the console command.
func (receiver MigrateCommand) Handle(c *cli.Context) error {
	config := helpers.GetDatabaseConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		config["username"], config["password"], config["host"], config["port"], config["database"], config["charset"], true, "Local")

	flag.Parse()
	var migrationDir = flag.String("migration.files", "./database/migrations", "Directory where the migration files are located ?")
	var mysqlDSN = flag.String("mysql.dsn", dsn, "Mysql DSN")

	db, err := sql.Open("mysql", *mysqlDSN)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not ping to database: %v", err)
	}

	// Run migrations
	driver, err := mysql.WithInstance(db, &mysql.Config{
		MigrationsTable: facades.Config.GetString("database.migrations"),
	})
	if err != nil {
		log.Fatalf("Could not start sql migration: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", *migrationDir),
		"mysql", driver)

	if err != nil {
		log.Fatalf("Migration init failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration success")

	return nil
}
