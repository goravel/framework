package migration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type RepositoryTestSuite struct {
	suite.Suite
	driverToTestQuery map[database.Driver]*gorm.TestQuery
}

func TestRepositoryTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, &RepositoryTestSuite{})
}

func (s *RepositoryTestSuite) SetupTest() {
	postgresDocker := docker.Postgres()
	postgresQuery := gorm.NewTestQuery(postgresDocker)
	s.driverToTestQuery = map[database.Driver]*gorm.TestQuery{
		database.DriverPostgres: postgresQuery,
	}
}

func (s *RepositoryTestSuite) TestCreate_Delete_Exists() {
	for driver, query := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			repository, mockOrm := s.initRepository(s.T(), driver, query.Query())

			mockOrm.EXPECT().Connection(driver.String()).Return(mockOrm).Once()
			mockOrm.EXPECT().Query().Return(repository.query).Once()

			err := repository.CreateRepository()
			s.NoError(err)

			mockOrm.EXPECT().Query().Return(repository.query).Once()

			s.True(repository.RepositoryExists())

			mockOrm.EXPECT().Connection(driver.String()).Return(mockOrm).Once()
			mockOrm.EXPECT().Query().Return(repository.query).Once()

			err = repository.DeleteRepository()
			s.NoError(err)

			mockOrm.EXPECT().Query().Return(repository.query).Once()

			s.False(repository.RepositoryExists())
		})
	}
}

func (s *RepositoryTestSuite) TestRecord() {
	for driver, query := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			repository, mockOrm := s.initRepository(s.T(), driver, query.Query())

			mockOrm.EXPECT().Query().Return(repository.query).Once()

			if !repository.RepositoryExists() {
				mockOrm.EXPECT().Connection(driver.String()).Return(mockOrm).Once()
				mockOrm.EXPECT().Query().Return(repository.query).Once()

				s.NoError(repository.CreateRepository())
			}

			err := repository.Log("migration1", 1)
			s.NoError(err)

			err = repository.Log("migration2", 1)
			s.NoError(err)

			err = repository.Log("migration3", 2)
			s.NoError(err)

			lastBatchNumber := repository.getLastBatchNumber()
			s.Equal(2, lastBatchNumber)

			nextBatchNumber := repository.GetNextBatchNumber()
			s.Equal(3, nextBatchNumber)

			ranMigrations, err := repository.GetRan()
			s.NoError(err)
			s.ElementsMatch([]string{"migration1", "migration2", "migration3"}, ranMigrations)

			migrations, err := repository.GetMigrations(2)
			s.NoError(err)
			s.ElementsMatch([]migration.File{
				{Migration: "migration3", Batch: 2},
				{Migration: "migration2", Batch: 1},
			}, migrations)

			migrations, err = repository.GetMigrationsByBatch(1)
			s.NoError(err)
			s.ElementsMatch([]migration.File{
				{Migration: "migration2", Batch: 1},
				{Migration: "migration1", Batch: 1},
			}, migrations)

			migrations, err = repository.GetLast()
			s.NoError(err)
			s.ElementsMatch([]migration.File{
				{Migration: "migration3", Batch: 2},
			}, migrations)

			err = repository.Delete("migration1")
			s.NoError(err)

			ranMigrations, err = repository.GetRan()
			s.NoError(err)
			s.ElementsMatch([]string{"migration2", "migration3"}, ranMigrations)
		})
	}
}

func (s *RepositoryTestSuite) initRepository(t *testing.T, driver database.Driver, query orm.Query) (*Repository, *mocksorm.Orm) {
	schema, _, _, mockOrm := initSchema(s.T(), driver)

	return NewRepository(query, schema, schema.prefix+"migrations"), mockOrm
}
