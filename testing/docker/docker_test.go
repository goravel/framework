package docker

import (
	"testing"

	"github.com/stretchr/testify/suite"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	foundationmocks "github.com/goravel/framework/contracts/foundation/mocks"
)

type DockerTestSuite struct {
	suite.Suite
	mockApp *foundationmocks.Application
	docker  *Docker
}

func TestDockerTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTestSuite))
}

func (s *DockerTestSuite) SetupTest() {
	s.mockApp = &foundationmocks.Application{}
	s.docker = NewDocker(s.mockApp)
}

func (s *DockerTestSuite) TestDatabase() {
	mockConfig := &configmocks.Config{}
	mockConfig.On("GetString", "database.default").Return("mysql").Once()
	mockConfig.On("GetString", "database.connections.mysql.driver").Return("mysql").Once()
	s.mockApp.On("MakeConfig").Return(mockConfig).Once()

	database, err := s.docker.Database()
	s.Nil(err)
	s.NotNil(database)
	databaseImpl := database.(*Database)
	s.Equal("mysql", databaseImpl.connection)

	mockConfig = &configmocks.Config{}
	mockConfig.On("GetString", "database.connections.postgresql.driver").Return("postgresql").Once()
	s.mockApp.On("MakeConfig").Return(mockConfig).Once()

	database, err = s.docker.Database("postgresql")
	s.Nil(err)
	s.NotNil(database)
	databaseImpl = database.(*Database)
	s.Equal("postgresql", databaseImpl.connection)
}
