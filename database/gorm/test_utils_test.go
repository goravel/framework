package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

func TestMysqlDocker(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	if err := testDatabaseDocker.Fresh(); err != nil {
		t.Fatal(err)
	}

	docker := NewMysqlDocker(testDatabaseDocker)
	query, err := docker.New()

	assert.NotNil(t, query)
	assert.Nil(t, err)
}

func TestPostgresqlDocker(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	if err := testDatabaseDocker.Fresh(); err != nil {
		t.Fatal(err)
	}

	docker := NewPostgresqlDocker(testDatabaseDocker)
	query, err := docker.New()

	assert.NotNil(t, query)
	assert.Nil(t, err)
}

func TestSqliteDocker(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	docker := NewSqliteDocker(dbDatabase)
	db, err := docker.New()

	assert.NotNil(t, db)
	assert.Nil(t, err)
	assert.Nil(t, file.Remove("goravel"))
}

func TestSqlserverDocker(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	if err := testDatabaseDocker.Fresh(); err != nil {
		t.Fatal(err)
	}

	docker := NewSqlserverDocker(testDatabaseDocker)
	db, err := docker.New()

	assert.NotNil(t, db)
	assert.Nil(t, err)
}
