package schema

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type CommonSchemaSuite struct {
	suite.Suite
	driverToTestQuery map[database.Driver]*gorm.TestQuery
}

func TestCommonSchemaSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	suite.Run(t, &CommonSchemaSuite{})
}

func (s *CommonSchemaSuite) SetupTest() {
	postgresDocker := docker.Postgres()
	s.Require().NoError(postgresDocker.Ready())

	postgresQuery := gorm.NewTestQuery(postgresDocker, true)
	s.driverToTestQuery = map[database.Driver]*gorm.TestQuery{
		database.DriverPostgres: postgresQuery,
	}
}

// TODO Implement this after implementing create view
func (s *CommonSchemaSuite) TestGetViews() {

}
