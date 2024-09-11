package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

func TestMysqlDocker(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	mysqlDocker := NewMysqlDocker(docker.Mysql())
	query, err := mysqlDocker.New()

	assert.NotNil(t, query)
	assert.Nil(t, err)
}

func TestPostgresDocker(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	postgresDocker := NewPostgresDocker(docker.Postgres())
	query, err := postgresDocker.New()

	assert.NotNil(t, query)
	assert.Nil(t, err)
}

func TestSqliteDocker(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	sqliteDocker := NewSqliteDocker(docker.Sqlite())
	db, err := sqliteDocker.New()

	assert.NotNil(t, db)
	assert.Nil(t, err)
	assert.Nil(t, file.Remove("goravel"))
}

func TestSqlserverDocker(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	sqlserverDocker := NewSqlserverDocker(docker.Sqlserver())
	db, err := sqlserverDocker.New()

	assert.NotNil(t, db)
	assert.Nil(t, err)
}
