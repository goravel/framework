package migration

import (
	"log"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/migration"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	supportdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type RepositoryTestSuite struct {
	suite.Suite
	repository *Repository
	mockOrm    *mocksorm.Orm
}

func TestRepositoryTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, &RepositoryTestSuite{})
}

func (s *RepositoryTestSuite) SetupTest() {
	postgresDriver := supportdocker.Postgres()
	postgresDocker := gorm.NewPostgresDocker(postgresDriver)
	postgresQuery, err := postgresDocker.New()
	if err != nil {
		log.Fatalf("Init postgres docker error: %v", err)
	}

	schema, _, _, mockOrm := initSchema(s.T(), contractsorm.DriverPostgres)

	s.repository = NewRepository(postgresQuery, schema, "migrations")
	s.mockOrm = mockOrm
}

func (s *RepositoryTestSuite) TestCreate_Delete_Exists() {
	s.mockOrm.EXPECT().Connection(contractsorm.DriverPostgres.String()).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.repository.query).Once()

	err := s.repository.CreateRepository()
	s.NoError(err)

	s.mockOrm.EXPECT().Query().Return(s.repository.query).Once()

	s.True(s.repository.RepositoryExists())

	s.mockOrm.EXPECT().Connection(contractsorm.DriverPostgres.String()).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.repository.query).Once()

	err = s.repository.DeleteRepository()
	s.NoError(err)

	s.mockOrm.EXPECT().Query().Return(s.repository.query).Once()

	s.False(s.repository.RepositoryExists())
}

func (s *RepositoryTestSuite) TestRecord() {
	err := s.repository.Log("migration1", 1)
	s.NoError(err)

	err = s.repository.Log("migration2", 1)
	s.NoError(err)

	err = s.repository.Log("migration3", 2)
	s.NoError(err)

	lastBatchNumber := s.repository.getLastBatchNumber()
	s.Equal(2, lastBatchNumber)

	nextBatchNumber := s.repository.GetNextBatchNumber()
	s.Equal(3, nextBatchNumber)

	ranMigrations, err := s.repository.GetRan()
	s.NoError(err)
	s.ElementsMatch([]string{"migration1", "migration2", "migration3"}, ranMigrations)

	migrations, err := s.repository.GetMigrations(2)
	s.NoError(err)
	s.ElementsMatch([]migration.File{
		{Migration: "migration3", Batch: 2},
		{Migration: "migration2", Batch: 1},
	}, migrations)

	migrations, err = s.repository.GetMigrationsByBatch(1)
	s.NoError(err)
	s.ElementsMatch([]migration.File{
		{Migration: "migration2", Batch: 1},
		{Migration: "migration1", Batch: 1},
	}, migrations)

	migrations, err = s.repository.GetLast()
	s.NoError(err)
	s.ElementsMatch([]migration.File{
		{Migration: "migration3", Batch: 2},
	}, migrations)

	err = s.repository.Delete("migration1")
	s.NoError(err)

	ranMigrations, err = s.repository.GetRan()
	s.NoError(err)
	s.ElementsMatch([]string{"migration2", "migration3"}, ranMigrations)
}
