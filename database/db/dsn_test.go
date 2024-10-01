package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database"
)

const (
	testHost     = "127.0.0.1"
	testPort     = 3306
	testDatabase = "forge"
	testUsername = "root"
	testPassword = "123123"
)

var testConfig = database.Config{
	Host:     testHost,
	Port:     testPort,
	Database: testDatabase,
	Username: testUsername,
	Password: testPassword,
}

func TestDsn(t *testing.T) {
	tests := []struct {
		name      string
		config    database.FullConfig
		expectDsn string
	}{
		{
			name: "empty",
			config: database.FullConfig{
				Config: database.Config{},
			},
		},
		{
			name: "mysql",
			config: database.FullConfig{
				Config:  testConfig,
				Driver:  database.DriverMysql,
				Charset: "utf8mb4",
				Loc:     "Local",
			},
			expectDsn: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&multiStatements=true",
				testUsername, testPassword, testHost, testPort, testDatabase, "utf8mb4", true, "Local"),
		},
		{
			name: "postgres",
			config: database.FullConfig{
				Config:   testConfig,
				Driver:   database.DriverPostgres,
				Sslmode:  "disable",
				Timezone: "UTC",
			},
			expectDsn: fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=%s",
				testUsername, testPassword, testHost, testPort, testDatabase, "disable", "UTC"),
		},
		{
			name: "sqlite",
			config: database.FullConfig{
				Config: testConfig,
				Driver: database.DriverSqlite,
			},
			expectDsn: fmt.Sprintf("%s?multi_stmts=true", testDatabase),
		},
		{
			name: "sqlserver",
			config: database.FullConfig{
				Config:  testConfig,
				Driver:  database.DriverSqlserver,
				Charset: "utf8mb4",
			},
			expectDsn: fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&charset=%s&MultipleActiveResultSets=true",
				testUsername, testPassword, testHost, testPort, testDatabase, "utf8mb4"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dsn := Dsn(test.config)
			assert.Equal(t, test.expectDsn, dsn)
		})
	}
}
