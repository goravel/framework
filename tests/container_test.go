package tests

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/testing/docker"
	mocksdocker "github.com/goravel/framework/mocks/testing/docker"
	supportdocker "github.com/goravel/framework/support/docker"
	supportfile "github.com/goravel/framework/support/file"
)

type ContainerTestSuite struct {
	suite.Suite
	mockDatabaseDriver *mocksdocker.DatabaseDriver
	container          *Container
}

func TestContainerTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerTestSuite))
}

func (s *ContainerTestSuite) SetupTest() {
	supportdocker.TestPortUsing = false
	s.container, s.mockDatabaseDriver = s.newMockContainer()
}

// newMockContainer keeps the registry files in a temp dir, so cleanup here leaves the machine's records alone.
func (s *ContainerTestSuite) newMockContainer() (*Container, *mocksdocker.DatabaseDriver) {
	mockDatabaseDriver := mocksdocker.NewDatabaseDriver(s.T())
	mockDatabaseDriver.EXPECT().Driver().Return("test").Once()

	container := NewContainer(mockDatabaseDriver)
	tempDir := s.T().TempDir()
	container.file = filepath.Join(tempDir, "goravel_docker.txt")
	container.lockFile = filepath.Join(tempDir, "goravel_docker.lock")

	return container, mockDatabaseDriver
}

func (s *ContainerTestSuite) TestAddAndAll() {
	s.mockDatabaseDriver.EXPECT().Config().Return(docker.DatabaseConfig{
		ContainerID: "test-container",
		Port:        5432,
		Database:    "test",
		Username:    "test",
		Password:    "test",
	}).Once()

	s.NoError(s.container.add())

	containers, err := s.container.all()
	s.NoError(err)
	s.NotEmpty(containers)
	s.Equal(docker.DatabaseConfig{
		ContainerID: "test-container",
		Port:        5432,
		Database:    "test",
		Username:    "test",
		Password:    "test",
	}, containers["test"])
}

func (s *ContainerTestSuite) TestBuild() {
	s.Run("Test reusing existing container", func() {
		supportdocker.TestPortUsing = true

		s.mockDatabaseDriver.EXPECT().Config().Return(docker.DatabaseConfig{
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
		s.mockDatabaseDriver.EXPECT().Config().Return(docker.DatabaseConfig{
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
	s.mockDatabaseDriver.EXPECT().Config().Return(docker.DatabaseConfig{
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

func (s *ContainerTestSuite) TestAddRewritesNullFile() {
	s.Require().NoError(supportfile.PutContent(s.container.file, "null"))
	s.mockDatabaseDriver.EXPECT().Config().Return(docker.DatabaseConfig{
		ContainerID: "test-container",
		Port:        5432,
	}).Once()

	s.NoError(s.container.add())

	containers, err := s.container.all()
	s.Require().NoError(err)
	s.Equal(map[string]docker.DatabaseConfig{
		"test": {ContainerID: "test-container", Port: 5432},
	}, containers)
}

func (s *ContainerTestSuite) TestCleanupContainers() {
	s.Require().NoError(s.container.write(map[string]docker.DatabaseConfig{
		"test": {ContainerID: "this-process-container"},
	}))
	s.mockDatabaseDriver.EXPECT().Shutdown().Return(nil).Once()
	s.container.Track()

	s.NoError(CleanupContainers())
	s.Empty(trackedContainers)
	s.NoFileExists(s.container.file)

	s.NoError(CleanupContainers())
}

func (s *ContainerTestSuite) TestCleanupContainersReturnsError() {
	s.mockDatabaseDriver.EXPECT().Shutdown().Return(assert.AnError).Once()
	s.container.Track()

	s.ErrorIs(CleanupContainers(), assert.AnError)
	s.Empty(trackedContainers)
}

func (s *ContainerTestSuite) TestCleanupContainersStopsEveryContainer() {
	secondContainer, secondMockDatabaseDriver := s.newMockContainer()
	secondMockDatabaseDriver.EXPECT().Shutdown().Return(assert.AnError).Once()

	s.mockDatabaseDriver.EXPECT().Shutdown().Return(nil).Once()
	s.container.Track()
	secondContainer.Track()

	// The failing container is tracked last, so the successful one only stops if cleanup keeps going.
	s.ErrorIs(CleanupContainers(), assert.AnError)
	s.Empty(trackedContainers)
}

func (s *ContainerTestSuite) TestCleanupContainersKeepsOtherRecords() {
	s.Require().NoError(s.container.write(map[string]docker.DatabaseConfig{
		"test":            {ContainerID: "this-process-container"},
		"other-container": {ContainerID: "other-process-container"},
	}))
	s.mockDatabaseDriver.EXPECT().Shutdown().Return(nil).Once()
	s.container.Track()

	s.NoError(CleanupContainers())

	containers, err := s.container.all()
	s.Require().NoError(err)
	s.Equal(map[string]docker.DatabaseConfig{
		"other-container": {ContainerID: "other-process-container"},
	}, containers)
}
