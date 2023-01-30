package gorm

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/testing/mock"
)

func TestGetGormConfig(t *testing.T) {
	var mockConfig *mocks.Config

	tests := []struct {
		name            string
		connection      orm.Driver
		setup           func()
		expectDialector gorm.Dialector
		expectErr       error
	}{
		{
			name:       "mysql",
			connection: orm.DriverMysql,
			setup: func() {
				mockConfig.On("GetString", "database.connections.mysql.driver").
					Return(orm.DriverMysql.String()).Once()
				mockConfig.On("GetString", "database.connections.mysql.host").
					Return("127.0.0.1").Once()
				mockConfig.On("GetString", "database.connections.mysql.port").
					Return("3306").Once()
				mockConfig.On("GetString", "database.connections.mysql.database").
					Return("goravel").Once()
				mockConfig.On("GetString", "database.connections.mysql.username").
					Return("root").Once()
				mockConfig.On("GetString", "database.connections.mysql.password").
					Return("123123").Once()
				mockConfig.On("GetString", "database.connections.mysql.charset").
					Return("utf8mb4").Once()
				mockConfig.On("GetString", "database.connections.mysql.loc").
					Return("Local").Once()
			},
			expectDialector: mysql.New(mysql.Config{
				DSN: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
					"root", "123123", "127.0.0.1", "3306", "goravel", "utf8mb4", true, "Local"),
			}),
		},
		{
			name:       "postgresql",
			connection: orm.DriverPostgresql,
			setup: func() {
				mockConfig.On("GetString", "database.connections.postgresql.driver").
					Return(orm.DriverPostgresql.String()).Once()
				mockConfig.On("GetString", "database.connections.postgresql.host").
					Return("127.0.0.1").Once()
				mockConfig.On("GetString", "database.connections.postgresql.port").
					Return("3306").Once()
				mockConfig.On("GetString", "database.connections.postgresql.database").
					Return("goravel").Once()
				mockConfig.On("GetString", "database.connections.postgresql.username").
					Return("root").Once()
				mockConfig.On("GetString", "database.connections.postgresql.password").
					Return("123123").Once()
				mockConfig.On("GetString", "database.connections.postgresql.sslmode").
					Return("disable").Once()
				mockConfig.On("GetString", "database.connections.postgresql.timezone").
					Return("UTC").Once()
			},
			expectDialector: postgres.New(postgres.Config{
				DSN: fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
					"127.0.0.1", "root", "123123", "goravel", "3306", "disable", "UTC"),
			}),
		},
		{
			name:       "sqlite",
			connection: orm.DriverSqlite,
			setup: func() {
				mockConfig.On("GetString", "database.connections.sqlite.driver").
					Return(orm.DriverSqlite.String()).Once()
				mockConfig.On("GetString", "database.connections.sqlite.database").
					Return("goravel").Once()
			},
			expectDialector: sqlite.Open("goravel"),
		},
		{
			name:       "sqlserver",
			connection: orm.DriverSqlserver,
			setup: func() {
				mockConfig.On("GetString", "database.connections.sqlserver.driver").
					Return(orm.DriverSqlserver.String()).Once()
				mockConfig.On("GetString", "database.connections.sqlserver.host").
					Return("127.0.0.1").Once()
				mockConfig.On("GetString", "database.connections.sqlserver.port").
					Return("5432").Once()
				mockConfig.On("GetString", "database.connections.sqlserver.database").
					Return("goravel").Once()
				mockConfig.On("GetString", "database.connections.sqlserver.username").
					Return("root").Once()
				mockConfig.On("GetString", "database.connections.sqlserver.password").
					Return("123123").Once()
			},
			expectDialector: sqlserver.New(sqlserver.Config{
				DSN: fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
					"root", "123123", "127.0.0.1", "5432", "goravel"),
			}),
		},
		{
			name:       "error driver",
			connection: "goravel",
			setup: func() {
				mockConfig.On("GetString", "database.connections.goravel.driver").
					Return("goravel").Once()
			},
			expectErr: errors.New(fmt.Sprintf("err database driver: %s, only support mysql, postgresql, sqlite and sqlserver", "goravel")),
		},
	}

	for _, test := range tests {
		mockConfig = mock.Config()
		test.setup()
		dialector, err := config(test.connection.String())
		assert.Equal(t, test.expectDialector, dialector)
		assert.Equal(t, test.expectErr, err)
	}
}
