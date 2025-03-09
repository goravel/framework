package tests

import (
	"fmt"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/driver"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/testing/utils"
	"github.com/goravel/mysql"
	"github.com/goravel/postgres"
	"github.com/goravel/sqlite"
	"github.com/goravel/sqlserver"
)

func mockDatabaseConfig(mockConfig *mocksconfig.Config, config database.Config, connection string, prefix string, singular bool) {
	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.write", connection)).Return(nil)
	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.read", connection)).Return(nil)

	mockDatabaseConfigWithoutWriteAndRead(mockConfig, config, connection, prefix, singular)
}

func mockDatabaseConfigWithoutWriteAndRead(mockConfig *mocksconfig.Config, config database.Config, connection string, prefix string, singular bool) {
	mockConfig.EXPECT().GetBool("app.debug").Return(true)
	mockConfig.EXPECT().GetInt("database.slow_threshold", 200).Return(200)
	mockConfig.EXPECT().GetInt("database.pool.max_idle_conns", 10).Return(10)
	mockConfig.EXPECT().GetInt("database.pool.max_open_conns", 100).Return(100)
	mockConfig.EXPECT().GetInt("database.pool.conn_max_idletime", 3600).Return(3600)
	mockConfig.EXPECT().GetInt("database.pool.conn_max_lifetime", 3600).Return(3600)

	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.host", connection)).Return(config.Host)
	mockConfig.EXPECT().GetInt(fmt.Sprintf("database.connections.%s.port", connection)).Return(config.Port)
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.username", connection)).Return(config.Username)
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.password", connection)).Return(config.Password)
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.database", connection)).Return(config.Database)
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", connection)).Return(prefix)
	mockConfig.EXPECT().GetBool(fmt.Sprintf("database.connections.%s.singular", connection)).Return(singular)
	mockConfig.EXPECT().GetBool(fmt.Sprintf("database.connections.%s.no_lower_case", connection)).Return(false)
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.dsn", connection)).Return("")
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.schema", connection), "public").Return("public")
	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.name_replacer", connection)).Return(nil)

	if config.Driver == postgres.Name {
		mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.sslmode", connection)).Return("disable")
		mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.timezone", connection)).Return("UTC")
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.via", connection)).Return(func() (driver.Driver, error) {
			return postgres.NewPostgres(mockConfig, utils.NewTestLog(), connection), nil
		})
	}
	if config.Driver == mysql.Name {
		mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.loc", connection)).Return("UTC")
		mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.charset", connection)).Return("utf8mb4")
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.via", connection)).Return(func() (driver.Driver, error) {
			return mysql.NewMysql(mockConfig, utils.NewTestLog(), connection), nil
		})
	}
	if config.Driver == sqlserver.Name {
		mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.charset", connection)).Return("utf8mb4")
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.via", connection)).Return(func() (driver.Driver, error) {
			return sqlserver.NewSqlserver(mockConfig, utils.NewTestLog(), connection), nil
		})
	}
	if config.Driver == sqlite.Name {
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.via", connection)).Return(func() (driver.Driver, error) {
			return sqlite.NewSqlite(mockConfig, utils.NewTestLog(), connection), nil
		})
	}
}
