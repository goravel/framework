package migration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
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
	postgresQuery := gorm.NewTestQuery(postgresDocker, true)
	s.driverToTestQuery = map[database.Driver]*gorm.TestQuery{
		database.DriverPostgres: postgresQuery,
	}
}

func (s *RepositoryTestSuite) TestCreate_Delete_Exists() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			repository, mockOrm := s.initRepository(testQuery)

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
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			repository, mockOrm := s.initRepository(testQuery)

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

			lastBatchNumber, err := repository.getLastBatchNumber()
			s.NoError(err)
			s.Equal(2, lastBatchNumber)

			nextBatchNumber, err := repository.GetNextBatchNumber()
			s.NoError(err)
			s.Equal(3, nextBatchNumber)

			ranMigrations, err := repository.GetRan()
			s.NoError(err)
			s.ElementsMatch([]string{"migration1", "migration2", "migration3"}, ranMigrations)

			migrations, err := repository.GetMigrations(2)
			s.NoError(err)
			s.Len(migrations, 2)
			s.Equal("migration3", migrations[0].Migration)
			s.Equal(2, migrations[0].Batch)
			s.Equal("migration2", migrations[1].Migration)
			s.Equal(1, migrations[1].Batch)

			migrations, err = repository.GetMigrationsByBatch(1)
			s.NoError(err)
			s.Len(migrations, 2)
			s.Equal("migration2", migrations[0].Migration)
			s.Equal(1, migrations[0].Batch)
			s.Equal("migration1", migrations[1].Migration)
			s.Equal(1, migrations[1].Batch)

			migrations, err = repository.GetLast()
			s.NoError(err)
			s.Len(migrations, 1)
			s.Equal("migration3", migrations[0].Migration)
			s.Equal(2, migrations[0].Batch)

			err = repository.Delete("migration1")
			s.NoError(err)

			ranMigrations, err = repository.GetRan()
			s.NoError(err)
			s.ElementsMatch([]string{"migration2", "migration3"}, ranMigrations)
		})
	}
}

func (s *RepositoryTestSuite) initRepository(testQuery *gorm.TestQuery) (*Repository, *mocksorm.Orm) {
	schema, mockOrm := initSchema(s.T(), testQuery)

	return NewRepository(testQuery.Query(), schema, "migrations"), mockOrm
}
