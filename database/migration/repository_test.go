package migration

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/schema"
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
		t.Skip("Skipping tests that use Docker")
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
			mockTransaction(mockOrm, testQuery)

			s.NoError(repository.CreateRepository())

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			s.True(repository.RepositoryExists())

			mockTransaction(mockOrm, testQuery)

			s.NoError(repository.DeleteRepository())

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			s.False(repository.RepositoryExists())
		})
	}
}

func (s *RepositoryTestSuite) TestRecord() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			repository, mockOrm := s.initRepository(testQuery)

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			if !repository.RepositoryExists() {
				mockTransaction(mockOrm, testQuery)

				s.NoError(repository.CreateRepository())
			}

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			err := repository.Log("migration1", 1)
			s.NoError(err)

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			err = repository.Log("migration2", 1)
			s.NoError(err)

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			err = repository.Log("migration3", 2)
			s.NoError(err)

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			lastBatchNumber, err := repository.getLastBatchNumber()
			s.NoError(err)
			s.Equal(2, lastBatchNumber)

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			nextBatchNumber, err := repository.GetNextBatchNumber()
			s.NoError(err)
			s.Equal(3, nextBatchNumber)

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			ranMigrations, err := repository.GetRan()
			s.NoError(err)
			s.ElementsMatch([]string{"migration1", "migration2", "migration3"}, ranMigrations)

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			migrations, err := repository.GetMigrations(2)

			s.NoError(err)
			s.Len(migrations, 2)
			s.Equal("migration3", migrations[0].Migration)
			s.Equal(2, migrations[0].Batch)
			s.Equal("migration2", migrations[1].Migration)
			s.Equal(1, migrations[1].Batch)

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			migrations, err = repository.GetMigrationsByBatch(1)

			s.NoError(err)
			s.Len(migrations, 2)
			s.Equal("migration2", migrations[0].Migration)
			s.Equal(1, migrations[0].Batch)
			s.Equal("migration1", migrations[1].Migration)
			s.Equal(1, migrations[1].Batch)

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Twice()

			migrations, err = repository.GetLast()

			s.NoError(err)
			s.Len(migrations, 1)
			s.Equal("migration3", migrations[0].Migration)
			s.Equal(2, migrations[0].Batch)

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			err = repository.Delete("migration1")
			s.NoError(err)

			mockOrm.EXPECT().Query().Return(testQuery.Query()).Once()

			ranMigrations, err = repository.GetRan()
			s.NoError(err)
			s.ElementsMatch([]string{"migration2", "migration3"}, ranMigrations)
		})
	}
}

func (s *RepositoryTestSuite) initRepository(testQuery *gorm.TestQuery) (*Repository, *mocksorm.Orm) {
	testSchema, mockOrm := schema.GetTestSchema(s.T(), testQuery)

	return NewRepository(testSchema, "migrations"), mockOrm
}

func mockTransaction(mockOrm *mocksorm.Orm, testQuery *gorm.TestQuery) {
	mockOrm.EXPECT().Transaction(mock.Anything).RunAndReturn(func(txFunc func(contractsorm.Query) error) error {
		return txFunc(testQuery.Query())
	}).Once()
}
