package docker

import (
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/database/orm"
)

type SqlserverTestSuite struct {
	suite.Suite
	mockConfig *configmocks.Config
	sqlserver  *Sqlserver
}

func TestSqlserverTestSuite(t *testing.T) {
	suite.Run(t, new(SqlserverTestSuite))
}

func (s *SqlserverTestSuite) SetupTest() {
	s.mockConfig = configmocks.NewConfig(s.T())
	s.sqlserver = &Sqlserver{
		config:     s.mockConfig,
		connection: "sqlserver",
	}
}

func (s *SqlserverTestSuite) TestName() {
	s.Equal(orm.DriverSqlserver, s.sqlserver.Name())
}

func (s *SqlserverTestSuite) TestImage() {
	s.mockConfig.On("GetString", "database.connections.sqlserver.password").Return("123123").Once()

	s.Equal(&dockertest.RunOptions{
		Repository: "mcr.microsoft.com/mssql/server",
		Tag:        "latest",
		Env: []string{
			"MSSQL_SA_PASSWORD=123123",
			"ACCEPT_EULA=Y",
		},
	}, s.sqlserver.Image())
}
