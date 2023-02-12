package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/file"
)

func TestMysqlDocker(t *testing.T) {
	pool, resource, db, err := MysqlDocker()

	assert.NotNil(t, pool)
	assert.NotNil(t, resource)
	assert.NotNil(t, db)
	assert.Nil(t, err)
}

func TestPostgresqlDocker(t *testing.T) {
	pool, resource, db, err := PostgresqlDocker()

	assert.NotNil(t, pool)
	assert.NotNil(t, resource)
	assert.NotNil(t, db)
	assert.Nil(t, err)
}

func TestSqliteDocker(t *testing.T) {
	pool, resource, db, err := SqliteDocker(dbDatabase)

	assert.NotNil(t, pool)
	assert.NotNil(t, resource)
	assert.NotNil(t, db)
	assert.Nil(t, err)

	file.Remove("goravel")
}

func TestSqlserverDocker(t *testing.T) {
	pool, resource, db, err := SqlserverDocker()

	assert.NotNil(t, pool)
	assert.NotNil(t, resource)
	assert.NotNil(t, db)
	assert.Nil(t, err)
}
