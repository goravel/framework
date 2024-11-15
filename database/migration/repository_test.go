package migration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/schema"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type RepositoryTestSuite struct {
	suite.Suite
	driverToTestQuery map[database.Driver]*gorm.TestQuery
}

func TestRepositoryTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	suite.Run(t, &RepositoryTestSuite{})
}

func (s *RepositoryTestSuite) SetupTest() {
	postgresDocker := docker.Postgres()
	s.Require().NoError(postgresDocker.Ready())

	postgresQuery := gorm.NewTestQuery(postgresDocker, true)
	s.driverToTestQuery = map[database.Driver]*gorm.TestQuery{
		database.DriverPostgres: postgresQuery,
	}
}

func (s *RepositoryTestSuite) TestCreate_Delete_Exists() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			repository := s.initRepository(testQuery)

			s.NoError(repository.CreateRepository())
			s.True(repository.RepositoryExists())
			s.NoError(repository.DeleteRepository())
			s.False(repository.RepositoryExists())
		})
	}
}

func (s *RepositoryTestSuite) TestRecord() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			repository := s.initRepository(testQuery)

			if !repository.RepositoryExists() {
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

			migrations, err := repository.GetMigrationsByStep(2)

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

			migrations, err = repository.GetMigrations()

			s.NoError(err)
			s.Len(migrations, 3)
			s.Equal("migration3", migrations[0].Migration)
			s.Equal(2, migrations[0].Batch)
			s.Equal("migration2", migrations[1].Migration)
			s.Equal(1, migrations[1].Batch)
			s.Equal("migration1", migrations[2].Migration)
			s.Equal(1, migrations[2].Batch)

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

func (s *RepositoryTestSuite) initRepository(testQuery *gorm.TestQuery) *Repository {
	testSchema := schema.GetTestSchema(testQuery, s.driverToTestQuery)

	return NewRepository(testSchema, "migrations")
}
