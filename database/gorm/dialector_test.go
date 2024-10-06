package gorm

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/errors"
)

func TestGetDialectors(t *testing.T) {
	config := database.Config{
		Host: "localhost",
	}

	tests := []struct {
		name             string
		configs          []database.FullConfig
		expectDialectors func(dialector gorm.Dialector) bool
		expectError      error
	}{
		{
			name: "Sad path - dsn is empty",
			configs: []database.FullConfig{
				{
					Connection: "postgres",
				},
			},
			expectError: errors.OrmFailedToGenerateDNS.Args("postgres"),
		},
		{
			name: "Happy path - mysql",
			configs: []database.FullConfig{
				{
					Connection: "mysql",
					Driver:     database.DriverMysql,
					Config:     config,
				},
			},
			expectDialectors: func(dialector gorm.Dialector) bool {
				_, ok := dialector.(*mysql.Dialector)

				return ok
			},
		},
		{
			name: "Happy path - postgres",
			configs: []database.FullConfig{
				{
					Connection: "postgres",
					Driver:     database.DriverPostgres,
					Config:     config,
				},
			},
			expectDialectors: func(dialector gorm.Dialector) bool {
				_, ok := dialector.(*postgres.Dialector)

				return ok
			},
		},
		{
			name: "Happy path - sqlserver",
			configs: []database.FullConfig{
				{
					Connection: "sqlserver",
					Driver:     database.DriverSqlserver,
					Config:     config,
				},
			},
			expectDialectors: func(dialector gorm.Dialector) bool {
				_, ok := dialector.(*sqlserver.Dialector)

				return ok
			},
		},
		{
			name: "Happy path - sqlite",
			configs: []database.FullConfig{
				{
					Connection: "sqlite",
					Driver:     database.DriverSqlite,
					Config:     config,
				},
			},
			expectDialectors: func(dialector gorm.Dialector) bool {
				_, ok := dialector.(*sqlite.Dialector)

				return ok
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dialectors, err := getDialectors(test.configs)
			if test.expectError != nil {
				assert.EqualError(t, err, test.expectError.Error())
				assert.Nil(t, dialectors)
			} else {
				assert.NoError(t, err)
				assert.Len(t, dialectors, 1)
				assert.True(t, test.expectDialectors(dialectors[0]))
			}
		})
	}
}
