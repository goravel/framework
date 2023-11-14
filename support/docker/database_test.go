package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/env"
)

func TestInitDatabase(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	database1, err := InitDatabase()
	assert.Nil(t, err)
	assert.NotNil(t, database1)
	assert.True(t, database1.Mysql.Config().Port > 0)
	assert.True(t, database1.Postgresql.Config().Port > 0)
	assert.True(t, database1.Sqlserver.Config().Port > 0)

	database2, err := InitDatabase()
	assert.Nil(t, err)
	assert.NotNil(t, database2)
	assert.True(t, database2.Mysql.Config().Port > 0)
	assert.True(t, database2.Postgresql.Config().Port > 0)
	assert.True(t, database2.Sqlserver.Config().Port > 0)

	assert.Nil(t, database1.Fresh())
	assert.Nil(t, database2.Fresh())

	assert.Nil(t, database1.Stop())
	assert.Nil(t, database2.Stop())
}
