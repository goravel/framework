package docker

import (
	"testing"

	"github.com/stretchr/testify/suite"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

type DockerTestSuite struct {
	suite.Suite
	mockApp *mocksfoundation.Application
	docker  *Docker
}

func TestDockerTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTestSuite))
}

func (s *DockerTestSuite) SetupTest() {
	s.mockApp = mocksfoundation.NewApplication(s.T())
	s.docker = NewDocker(s.mockApp)
}

func (s *DockerTestSuite) TestDatabase() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("database.default").Return("mysql").Once()
	mockConfig.EXPECT().GetString("database.connections.mysql.driver").Return("mysql").Once()
	mockConfig.EXPECT().GetString("database.connections.mysql.database").Return("goravel").Once()
	mockConfig.EXPECT().GetString("database.connections.mysql.username").Return("goravel").Once()
	mockConfig.EXPECT().GetString("database.connections.mysql.password").Return("goravel").Once()

	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockOrm := mocksorm.NewOrm(s.T())

	s.mockApp.EXPECT().MakeArtisan().Return(mockArtisan).Once()
	s.mockApp.EXPECT().MakeConfig().Return(mockConfig).Once()
	s.mockApp.EXPECT().MakeOrm().Return(mockOrm).Once()

	database, err := s.docker.Database()
	s.Nil(err)
	s.NotNil(database)

	databaseImpl := database.(*Database)
	s.Equal("mysql", databaseImpl.connection)

	mockConfig = mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("database.connections.postgres.driver").Return("postgres").Once()
	mockConfig.EXPECT().GetString("database.connections.postgres.database").Return("goravel").Once()
	mockConfig.EXPECT().GetString("database.connections.postgres.username").Return("goravel").Once()
	mockConfig.EXPECT().GetString("database.connections.postgres.password").Return("goravel").Once()

	mockArtisan = mocksconsole.NewArtisan(s.T())
	mockOrm = mocksorm.NewOrm(s.T())

	s.mockApp.EXPECT().MakeArtisan().Return(mockArtisan).Once()
	s.mockApp.On("MakeConfig").Return(mockConfig).Once()
	s.mockApp.EXPECT().MakeOrm().Return(mockOrm).Once()

	database, err = s.docker.Database("postgres")
	s.Nil(err)
	s.NotNil(database)

	databaseImpl = database.(*Database)
	s.Equal("postgres", databaseImpl.connection)
}
