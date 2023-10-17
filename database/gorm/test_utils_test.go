package gorm

import (
	"os"
	"testing"

	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/file"
)

func TestMysqlDocker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}
	if len(os.Getenv("GORAVEL_DATABASE_TEST")) == 0 {
		color.Redln("Skip tests because not set GORAVEL_DATABASE_TEST environment variable")
		return
	}

	docker := NewMysqlDocker()
	pool, resource, query, err := docker.New()

	assert.NotNil(t, pool)
	assert.NotNil(t, resource)
	assert.NotNil(t, query)
	assert.Nil(t, err)
}

func TestPostgresqlDocker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}
	if len(os.Getenv("GORAVEL_DATABASE_TEST")) == 0 {
		color.Redln("Skip tests because not set GORAVEL_DATABASE_TEST environment variable")
		return
	}

	docker := NewPostgresqlDocker()
	pool, resource, query, err := docker.New()

	assert.NotNil(t, pool)
	assert.NotNil(t, resource)
	assert.NotNil(t, query)
	assert.Nil(t, err)
}

func TestSqliteDocker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}
	if len(os.Getenv("GORAVEL_DATABASE_TEST")) == 0 {
		color.Redln("Skip tests because not set GORAVEL_DATABASE_TEST environment variable")
		return
	}

	docker := NewSqliteDocker(dbDatabase)
	pool, resource, db, err := docker.New()

	assert.NotNil(t, pool)
	assert.NotNil(t, resource)
	assert.NotNil(t, db)
	assert.Nil(t, err)
	assert.Nil(t, file.Remove("goravel"))
}

func TestSqlserverDocker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}
	if len(os.Getenv("GORAVEL_DATABASE_TEST")) == 0 {
		color.Redln("Skip tests because not set GORAVEL_DATABASE_TEST environment variable")
		return
	}

	docker := NewSqlserverDocker()
	pool, resource, db, err := docker.New()

	assert.NotNil(t, pool)
	assert.NotNil(t, resource)
	assert.NotNil(t, db)
	assert.Nil(t, err)
}
