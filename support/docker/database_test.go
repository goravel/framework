package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support/env"
)

type DatabaseTestSuite struct {
	suite.Suite
}

func TestDatabaseTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, new(DatabaseTestSuite))
}

func (s *DatabaseTestSuite) SetupTest() {
}

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

	mysql1, err := database1.Mysql.connect()
	assert.Nil(t, err)
	assert.NotNil(t, mysql1)

	mysql2, err := database2.Mysql.connect()
	assert.Nil(t, err)
	assert.NotNil(t, mysql2)

	assert.Nil(t, database1.Stop())
	assert.Nil(t, database2.Stop())
}

func TestGetValidPort(t *testing.T) {
	assert.True(t, GetValidPort() > 0)
}
