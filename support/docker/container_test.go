package docker

import (
	"strings"
	"testing"

	contractstesting "github.com/goravel/framework/contracts/testing"
	mockstesting "github.com/goravel/framework/mocks/testing"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ContainerTestSuite struct {
	suite.Suite
	mockDatabaseDriver *mockstesting.DatabaseDriver
	container          *Container
}

func TestContainerTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerTestSuite))
}

func (s *ContainerTestSuite) SetupTest() {
	testPortUsing = false
	s.mockDatabaseDriver = mockstesting.NewDatabaseDriver(s.T())
	s.mockDatabaseDriver.EXPECT().Driver().Return("test").Once()
	s.container = NewContainer(s.mockDatabaseDriver)
}

func (s *ContainerTestSuite) TestAddAndAll() {
	s.mockDatabaseDriver.EXPECT().Config().Return(contractstesting.DatabaseConfig{
		ContainerID: "test-container",
		Port:        5432,
		Database:    "test",
		Username:    "test",
		Password:    "test",
	}).Once()

	s.NoError(s.container.add())

	containers, err := s.container.all()
	s.NoError(err)
	s.Len(containers, 1)
	s.Equal(contractstesting.DatabaseConfig{
		ContainerID: "test-container",
		Port:        5432,
		Database:    "test",
		Username:    "test",
		Password:    "test",
	}, containers["test"])
}

func (s *ContainerTestSuite) TestBuild() {
	s.Run("Test reusing existing container", func() {
		testPortUsing = true

		s.mockDatabaseDriver.EXPECT().Config().Return(contractstesting.DatabaseConfig{
			ContainerID: "test-container",
			Port:        5432,
		}).Once()
		s.mockDatabaseDriver.EXPECT().Reuse("test-container", 5432).Return(nil).Once()
		s.mockDatabaseDriver.EXPECT().Database(mock.MatchedBy(func(database string) bool {
			return strings.HasPrefix(database, "goravel_")
		})).Return(s.mockDatabaseDriver, nil).Once()

		// Add existing container config
		s.NoError(s.container.add())

		// Build should reuse existing container
		result, err := s.container.Build()
		s.NoError(err)
		s.NotNil(result)
	})

	s.Run("Test creating new container", func() {
		s.SetupTest()

		s.mockDatabaseDriver.EXPECT().Build().Return(nil).Once()
		s.mockDatabaseDriver.EXPECT().Config().Return(contractstesting.DatabaseConfig{
			ContainerID: "test-container",
			Port:        5432,
		}).Once()
		s.mockDatabaseDriver.EXPECT().Database(mock.MatchedBy(func(database string) bool {
			return strings.HasPrefix(database, "goravel_")
		})).Return(s.mockDatabaseDriver, nil).Once()

		result, err := s.container.Build()
		s.NoError(err)
		s.NotNil(result)
	})
}

func (s *ContainerTestSuite) TestBuilds() {
	s.mockDatabaseDriver.EXPECT().Build().Return(nil).Times(3)
	s.mockDatabaseDriver.EXPECT().Config().Return(contractstesting.DatabaseConfig{
		ContainerID: "test-container",
		Port:        5432,
	}).Times(3)
	s.mockDatabaseDriver.EXPECT().Database(mock.MatchedBy(func(database string) bool {
		return strings.HasPrefix(database, "goravel_")
	})).Return(s.mockDatabaseDriver, nil).Times(3)

	result, err := s.container.Builds(3)
	s.NoError(err)
	s.Len(result, 3)
}
