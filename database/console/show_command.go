package console

import (
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/str"
)

type ShowCommand struct {
	config config.Config
	schema schema.Schema
}

type databaseInfo struct {
	Name            string
	Version         string
	Database        string
	Host            string
	Port            string
	Username        string
	OpenConnections string
	Tables          []schema.Table `gorm:"-"`
	Views           []schema.View  `gorm:"-"`
}

type queryResult struct{ Value string }

func NewShowCommand(config config.Config, schema schema.Schema) *ShowCommand {
	return &ShowCommand{
		config: config,
		schema: schema,
	}
}

// Signature The name and signature of the console command.
func (r *ShowCommand) Signature() string {
	return "db:show"
}

// Description The console command description.
func (r *ShowCommand) Description() string {
	return "Display information about the given database"
}

// Extend The console command extend.
func (r *ShowCommand) Extend() command.Extend {
	return command.Extend{
		Category: "db",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "database",
				Aliases: []string{"d"},
				Usage:   "The database connection",
			},
			&command.BoolFlag{
				Name:    "views",
				Aliases: []string{"v"},
				Usage:   "Show the database views",
			},
		},
	}
}

// Handle Execute the console command.
func (r *ShowCommand) Handle(ctx console.Context) error {
	if got := ctx.Argument(0); len(got) > 0 {
		ctx.Error(fmt.Sprintf("No arguments expected for '%s' command, got '%s'.", r.Signature(), got))
		return nil
	}
	r.schema = r.schema.Connection(ctx.Option("database"))
	connection := r.schema.GetConnection()
	getConfigValue := func(k string) string {
		return r.config.GetString("database.connections." + connection + "." + k)
	}
	info := databaseInfo{
		Database: getConfigValue("database"),
		Host:     getConfigValue("host"),
		Port:     getConfigValue("port"),
		Username: getConfigValue("username"),
	}
	var err error
	info.Name, info.Version, info.OpenConnections, err = r.getDataBaseInfo()
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}
	if info.Tables, err = r.schema.GetTables(); err != nil {
		ctx.Error(err.Error())
		return nil
	}
	if ctx.OptionBool("views") {
		if info.Views, err = r.schema.GetViews(); err != nil {
			ctx.Error(err.Error())
			return nil
		}
	}
	r.display(ctx, info)
	return nil
}

func (r *ShowCommand) getDataBaseInfo() (name, version, openConnections string, err error) {
	var (
		drivers = map[database.Driver]struct {
			name                 string
			versionQuery         string
			openConnectionsQuery string
		}{
			database.DriverSqlite: {
				name:         "SQLite",
				versionQuery: "SELECT sqlite_version() AS value;",
			},
			database.DriverMysql: {
				name:                 "MySQL",
				versionQuery:         "SELECT VERSION() AS value;",
				openConnectionsQuery: "SHOW status WHERE variable_name = 'threads_connected';",
			},
			database.DriverPostgres: {
				name:                 "PostgresSQL",
				versionQuery:         "SELECT current_setting('server_version') AS value;",
				openConnectionsQuery: "SELECT COUNT(*) AS value FROM pg_stat_activity;",
			},
			database.DriverSqlserver: {
				name:                 "SQL Server",
				versionQuery:         "SELECT SERVERPROPERTY('productversion') AS value;",
				openConnectionsQuery: "SELECT COUNT(*) Value FROM sys.dm_exec_sessions WHERE status = 'running';",
			},
		}
	)
	name = string(r.schema.Orm().Query().Driver())
	if driver, ok := drivers[r.schema.Orm().Query().Driver()]; ok {
		name = driver.name
		var versionResult queryResult
		if err = r.schema.Orm().Query().Raw(driver.versionQuery).Scan(&versionResult); err == nil {
			version = versionResult.Value
			if strings.Contains(version, "MariaDB") {
				name = "MariaDB"
			}
			if len(driver.openConnectionsQuery) > 0 {
				var openConnectionsResult queryResult
				if err = r.schema.Orm().Query().Raw(driver.openConnectionsQuery).Scan(&openConnectionsResult); err == nil {
					openConnections = openConnectionsResult.Value
				}
			}
		}
	}
	return
}

func (r *ShowCommand) display(ctx console.Context, info databaseInfo) {
	ctx.NewLine()
	ctx.TwoColumnDetail(fmt.Sprintf("<fg=green;op=bold>%s</>", info.Name), info.Version)
	ctx.TwoColumnDetail("Database", info.Database)
	ctx.TwoColumnDetail("Host", info.Host)
	ctx.TwoColumnDetail("Port", info.Port)
	ctx.TwoColumnDetail("Username", info.Username)
	ctx.TwoColumnDetail("Open Connections", info.OpenConnections)
	ctx.TwoColumnDetail("Tables", fmt.Sprintf("%d", len(info.Tables)))
	if size := func() (size int) {
		for i := range info.Tables {
			size += info.Tables[i].Size
		}
		return
	}(); size > 0 {
		ctx.TwoColumnDetail("Total Size", fmt.Sprintf("%.3fMiB", float64(size)/1024/1024))
	}
	ctx.NewLine()
	if len(info.Tables) > 0 {
		ctx.TwoColumnDetail("<fg=green;op=bold>Tables</>", "<fg=yellow;op=bold>Size (MiB)</>")
		for i := range info.Tables {
			ctx.TwoColumnDetail(info.Tables[i].Name, fmt.Sprintf("%.3f", float64(info.Tables[i].Size)/1024/1024))
		}
		ctx.NewLine()
	}
	if len(info.Views) > 0 {
		ctx.TwoColumnDetail("<fg=green;op=bold>Views</>", "<fg=yellow;op=bold>Rows</>")
		for i := range info.Views {
			if !str.Of(info.Views[i].Name).StartsWith("pg_catalog", "information_schema", "spt_") {
				var rows int64
				if err := r.schema.Orm().Query().Table(info.Views[i].Name).Count(&rows); err != nil {
					ctx.Error(err.Error())
					return
				}
				ctx.TwoColumnDetail(info.Views[i].Name, fmt.Sprintf("%d", rows))
			}
		}
		ctx.NewLine()
	}
}
