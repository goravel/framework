package migration

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/database/orm"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/gorm"
	databaseschema "github.com/goravel/framework/database/schema"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type DefaultMigratorWithDBSuite struct {
	suite.Suite
	driverToTestQuery map[contractsdatabase.Driver]*gorm.TestQuery
}

func TestDefaultMigratorWithDBSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests that use Docker")
	}

	suite.Run(t, &DefaultMigratorWithDBSuite{})
}

func (s *DefaultMigratorWithDBSuite) SetupTest() {
	// TODO Add other drivers
	postgresDocker := docker.Postgres()
	postgresQuery := gorm.NewTestQuery(postgresDocker, true)
	s.driverToTestQuery = map[contractsdatabase.Driver]*gorm.TestQuery{
		contractsdatabase.DriverPostgres: postgresQuery,
	}
}

func (s *DefaultMigratorWithDBSuite) TestRun() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			schema := databaseschema.GetTestSchema(testQuery, s.driverToTestQuery)
			testMigration := NewTestMigration(schema)
			schema.Register([]contractsschema.Migration{
				testMigration,
			})

			migrator := NewDefaultMigrator(nil, schema, "migrations")

			s.NoError(migrator.Run())
			s.True(schema.HasTable("users"))
		})
	}
}

// TODO Add rollback test cases after implementing Sqlite driver, to test migrating different databases.
func (s *DefaultMigratorWithDBSuite) TestRollback() {

}

type DefaultMigratorSuite struct {
	suite.Suite
	value          int
	mockArtisan    *mocksconsole.Artisan
	mockRepository *mocksmigration.Repository
	mockSchema     *mocksschema.Schema
	driver         *DefaultMigrator
}

func TestDefaultMigratorSuite(t *testing.T) {
	suite.Run(t, &DefaultMigratorSuite{})
}

func (s *DefaultMigratorSuite) SetupTest() {
	s.value = 0
	s.mockArtisan = mocksconsole.NewArtisan(s.T())
	s.mockRepository = mocksmigration.NewRepository(s.T())
	s.mockSchema = mocksschema.NewSchema(s.T())

	s.driver = &DefaultMigrator{
		artisan:    s.mockArtisan,
		creator:    NewDefaultCreator(),
		repository: s.mockRepository,
		schema:     s.mockSchema,
	}
}

func (s *DefaultMigratorSuite) TestCreate() {
	now := carbon.FromDateTime(2024, 8, 17, 21, 45, 1)
	carbon.SetTestNow(now)

	pwd, err := os.Getwd()
	s.NoError(err)

	path := filepath.Join(pwd, "database", "migrations")
	name := "create_users_table"

	s.NoError(s.driver.Create(name))

	migrationFile := filepath.Join(path, "20240817214501_"+name+".go")
	s.True(file.Exists(migrationFile))

	defer func() {
		carbon.UnsetTestNow()
		s.NoError(file.Remove("database"))
	}()
}

func (s *DefaultMigratorSuite) TestFresh() {
	// Success
	s.mockArtisan.EXPECT().Call("db:wipe --force").Return(nil).Once()
	s.mockArtisan.EXPECT().Call("migrate").Return(nil).Once()

	s.NoError(s.driver.Fresh())

	// db:wipe returns error
	s.mockArtisan.EXPECT().Call("db:wipe --force").Return(assert.AnError).Once()

	s.EqualError(s.driver.Fresh(), assert.AnError.Error())

	// migrate returns error
	s.mockArtisan.EXPECT().Call("db:wipe --force").Return(nil).Once()
	s.mockArtisan.EXPECT().Call("migrate").Return(assert.AnError).Once()

	s.EqualError(s.driver.Fresh(), assert.AnError.Error())
}

func (s *DefaultMigratorSuite) TestGetFilesForRollback() {
	tests := []struct {
		name        string
		step        int
		batch       int
		setup       func()
		expectFiles []migration.File
		expectError string
	}{
		{
			name:  "Returns migrations for step",
			step:  1,
			batch: 0,
			setup: func() {
				s.mockRepository.EXPECT().GetMigrations(1).Return([]migration.File{{Migration: "20240817214501_create_users_table"}}, nil).Once()
			},
			expectFiles: []migration.File{{Migration: "20240817214501_create_users_table"}},
		},
		{
			name:  "Returns migrations for batch",
			step:  0,
			batch: 1,
			setup: func() {
				s.mockRepository.EXPECT().GetMigrationsByBatch(1).Return([]migration.File{{Migration: "20240817214501_create_users_table"}}, nil).Once()
			},
			expectFiles: []migration.File{{Migration: "20240817214501_create_users_table"}},
		},
		{
			name:  "Returns last migrations",
			step:  0,
			batch: 0,
			setup: func() {
				s.mockRepository.EXPECT().GetLast().Return([]migration.File{{Migration: "20240817214501_create_users_table"}}, nil).Once()
			},
			expectFiles: []migration.File{{Migration: "20240817214501_create_users_table"}},
		},
		{
			name:  "Returns error when GetMigrations fails",
			step:  1,
			batch: 0,
			setup: func() {
				s.mockRepository.EXPECT().GetMigrations(1).Return(nil, errors.New("error")).Once()
			},
			expectError: "error",
		},
		{
			name:  "Returns error when GetMigrationsByBatch fails",
			step:  0,
			batch: 1,
			setup: func() {
				s.mockRepository.EXPECT().GetMigrationsByBatch(1).Return(nil, errors.New("error")).Once()
			},
			expectError: "error",
		},
		{
			name:  "Returns error when GetLast fails",
			step:  0,
			batch: 0,
			setup: func() {
				s.mockRepository.EXPECT().GetLast().Return(nil, errors.New("error")).Once()
			},
			expectError: "error",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()

			files, err := s.driver.getFilesForRollback(test.step, test.batch)
			if test.expectError == "" {
				s.NoError(err)
				s.Equal(test.expectFiles, files)
			} else {
				s.EqualError(err, test.expectError)
				s.Nil(files)
			}
		})
	}
}

func (s *DefaultMigratorSuite) TestRollback() {
	tests := []struct {
		name      string
		setup     func()
		expectErr string
	}{
		{
			name: "Rollback with no files",
			setup: func() {
				s.mockRepository.EXPECT().GetMigrations(1).Return(nil, nil).Once()
			},
		},
		{
			name: "Rollback with files",
			setup: func() {
				previousConnection := "postgres"
				testMigration := NewTestMigration(s.mockSchema)

				s.mockRepository.EXPECT().GetMigrations(1).Return([]migration.File{{Migration: testMigration.Signature()}}, nil).Once()
				s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{testMigration}).Once()

				mockOrm := mocksorm.NewOrm(s.T())
				s.mockRunDown(mockOrm, previousConnection, testMigration.Signature(), "users", nil)
			},
		},
		{
			name: "Rollback with missing migration",
			setup: func() {
				s.mockRepository.EXPECT().GetMigrations(1).Return([]migration.File{{Migration: "20240817214501_create_users_table"}}, nil).Once()
				s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{}).Once()
			},
		},
		{
			name: "Rollback with error",
			setup: func() {
				previousConnection := "postgres"
				testMigration := NewTestMigration(s.mockSchema)

				s.mockRepository.EXPECT().GetMigrations(1).Return([]migration.File{{Migration: testMigration.Signature()}}, nil).Once()
				s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{testMigration}).Once()

				mockOrm := mocksorm.NewOrm(s.T())
				s.mockRunDown(mockOrm, previousConnection, testMigration.Signature(), "users", assert.AnError)
			},
			expectErr: assert.AnError.Error(),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.value = 0
			test.setup()

			err := s.driver.Rollback(1, 0)

			if test.expectErr == "" {
				s.NoError(err)
			} else {
				s.EqualError(err, test.expectErr)
			}
		})
	}
}

func (s *DefaultMigratorSuite) TestRun() {
	testMigration := NewTestMigration(s.mockSchema)
	testConnectionMigration := NewTestConnectionMigration(s.mockSchema)

	tests := []struct {
		name        string
		setup       func()
		expectError string
	}{
		{
			name: "Happy path",
			setup: func() {
				previousConnection := "postgres"

				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetRan().Return([]string{testConnectionMigration.Signature()}, nil).Once()
				s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{
					testMigration,
					testConnectionMigration,
				}).Once()
				s.mockRepository.EXPECT().GetNextBatchNumber().Return(1, nil).Once()

				mockOrm := mocksorm.NewOrm(s.T())
				s.mockRunUp(mockOrm, previousConnection, testMigration.Signature(), "users", 1, nil)
			},
		},
		{
			name: "Sad path - Log returns error",
			setup: func() {
				previousConnection := "postgres"

				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetRan().Return([]string{testConnectionMigration.Signature()}, nil).Once()
				s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{
					testMigration,
					testConnectionMigration,
				}).Once()
				s.mockRepository.EXPECT().GetNextBatchNumber().Return(1, nil).Once()

				mockOrm := mocksorm.NewOrm(s.T())
				s.mockRunUp(mockOrm, previousConnection, testMigration.Signature(), "users", 1, assert.AnError)
			},
			expectError: assert.AnError.Error(),
		},
		{
			name: "Sad path - GetNextBatchNumber returns error",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetRan().Return([]string{testConnectionMigration.Signature()}, nil).Once()
				s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{
					testMigration,
					testConnectionMigration,
				}).Once()
				s.mockRepository.EXPECT().GetNextBatchNumber().Return(0, assert.AnError).Once()
			},
			expectError: assert.AnError.Error(),
		},
		{
			name: "Sad path - GetRan returns error",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetRan().Return(nil, assert.AnError).Once()
			},
			expectError: assert.AnError.Error(),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()

			err := s.driver.Run()
			if test.expectError == "" {
				s.Nil(err)
			} else {
				s.EqualError(err, test.expectError)
			}
		})
	}
}

func (s *DefaultMigratorSuite) TestPendingMigrations() {
	migrations := []contractsschema.Migration{
		NewTestMigration(s.mockSchema),
		NewTestConnectionMigration(s.mockSchema),
	}
	ran := []string{
		"20240817214501_create_users_table",
	}

	pendingMigrations := s.driver.pendingMigrations(migrations, ran)
	s.Len(pendingMigrations, 1)
	s.Equal(NewTestConnectionMigration(s.mockSchema), pendingMigrations[0])
}

func (s *DefaultMigratorSuite) TestPrepareDatabase() {
	s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
	s.NoError(s.driver.prepareDatabase())

	s.mockRepository.EXPECT().RepositoryExists().Return(false).Once()
	s.mockRepository.EXPECT().CreateRepository().Return(nil).Once()
	s.NoError(s.driver.prepareDatabase())
}

func (s *DefaultMigratorSuite) TestRunPending() {
	testMigration := NewTestMigration(s.mockSchema)

	tests := []struct {
		name        string
		migrations  []contractsschema.Migration
		setup       func()
		expectError string
	}{
		{
			name: "Happy path",
			migrations: []contractsschema.Migration{
				testMigration,
			},
			setup: func() {
				previousConnection := "postgres"
				mockOrm := mocksorm.NewOrm(s.T())

				s.mockRepository.EXPECT().GetNextBatchNumber().Return(1, nil).Once()
				s.mockRunUp(mockOrm, previousConnection, testMigration.Signature(), "users", 1, nil)
			},
		},
		{
			name:       "Happy path - no migrations",
			migrations: []contractsschema.Migration{},
			setup:      func() {},
		},
		{
			name: "Sad path - GetNextBatchNumber returns error",
			migrations: []contractsschema.Migration{
				testMigration,
			},
			setup: func() {
				s.mockRepository.EXPECT().GetNextBatchNumber().Return(0, assert.AnError).Once()
			},
			expectError: assert.AnError.Error(),
		},
		{
			name: "Sad path - runUp returns error",
			migrations: []contractsschema.Migration{
				testMigration,
			},
			setup: func() {
				previousConnection := "postgres"
				mockOrm := mocksorm.NewOrm(s.T())

				s.mockRepository.EXPECT().GetNextBatchNumber().Return(1, nil).Once()
				s.mockRunUp(mockOrm, previousConnection, testMigration.Signature(), "users", 1, assert.AnError)
			},
			expectError: assert.AnError.Error(),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()

			err := s.driver.runPending(test.migrations)
			if test.expectError == "" {
				s.Nil(err)
			} else {
				s.EqualError(err, test.expectError)
			}
		})
	}
}

func (s *DefaultMigratorSuite) TestRunDown() {
	var (
		previousConnection      = "postgres"
		testMigration           = NewTestMigration(s.mockSchema)
		testErrorMigration      = NewTestErrorMigration()
		testConnectionMigration = NewTestConnectionMigration(s.mockSchema)

		mockOrm *mocksorm.Orm
	)

	beforeEach := func() {
		s.value = 0
		mockOrm = mocksorm.NewOrm(s.T())
	}

	tests := []struct {
		name      string
		migration contractsschema.Migration
		setup     func()
		expectErr string
	}{
		{
			name:      "Happy path",
			migration: testMigration,
			setup: func() {
				s.mockRunDown(mockOrm, previousConnection, testMigration.Signature(), "users", nil)
			},
		},
		{
			name:      "Happy path - with connection",
			migration: testConnectionMigration,
			setup: func() {
				s.mockRunDown(mockOrm, previousConnection, testConnectionMigration.Signature(), "agents", nil)
			},
		},
		{
			name:      "Sad path - Down returns error",
			migration: testErrorMigration,
			setup: func() {
				s.mockSchema.EXPECT().GetConnection().Return(previousConnection).Once()
				s.mockSchema.EXPECT().Orm().Return(mockOrm).Times(4)
				mockOrm.EXPECT().Transaction(mock.Anything).RunAndReturn(func(f func(tx orm.Query) error) error {
					mockQuery := mocksorm.NewQuery(s.T())
					mockOrm.EXPECT().Query().Return(mockQuery).Once()
					mockOrm.EXPECT().SetQuery(mockQuery).Once()
					s.mockSchema.EXPECT().SetConnection(previousConnection).Once()
					mockOrm.EXPECT().SetQuery(mockQuery).Once()

					return f(mockQuery)
				}).Once()
			},
			expectErr: assert.AnError.Error(),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			beforeEach()
			test.setup()

			err := s.driver.runDown(test.migration)

			if test.expectErr == "" {
				s.NoError(err)
			} else {
				s.EqualError(err, test.expectErr)
			}
		})
	}
}

func (s *DefaultMigratorSuite) TestRunUp() {
	var (
		batch                   = 1
		previousConnection      = "postgres"
		testMigration           = NewTestMigration(s.mockSchema)
		testErrorMigration      = NewTestErrorMigration()
		testConnectionMigration = NewTestConnectionMigration(s.mockSchema)

		mockOrm *mocksorm.Orm
	)

	beforeEach := func() {
		s.value = 0
		mockOrm = mocksorm.NewOrm(s.T())
	}

	tests := []struct {
		name      string
		migration contractsschema.Migration
		setup     func()
		expectErr string
	}{
		{
			name:      "Happy path",
			migration: testMigration,
			setup: func() {
				s.mockRunUp(mockOrm, previousConnection, testMigration.Signature(), "users", batch, nil)
			},
		},
		{
			name:      "Happy path - with connection",
			migration: testConnectionMigration,
			setup: func() {
				s.mockRunUp(mockOrm, previousConnection, testConnectionMigration.Signature(), "agents", batch, nil)
			},
		},
		{
			name:      "Sad path - Up returns error",
			migration: testErrorMigration,
			setup: func() {
				s.mockSchema.EXPECT().GetConnection().Return(previousConnection).Once()
				s.mockSchema.EXPECT().Orm().Return(mockOrm).Times(4)
				mockOrm.EXPECT().Transaction(mock.Anything).RunAndReturn(func(f func(tx orm.Query) error) error {
					mockQuery := mocksorm.NewQuery(s.T())
					mockOrm.EXPECT().Query().Return(mockQuery).Once()
					mockOrm.EXPECT().SetQuery(mockQuery).Once()
					s.mockSchema.EXPECT().SetConnection(previousConnection).Once()
					mockOrm.EXPECT().SetQuery(mockQuery).Once()

					return f(mockQuery)
				}).Once()
			},
			expectErr: assert.AnError.Error(),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			beforeEach()
			test.setup()

			err := s.driver.runUp(test.migration, batch)

			if test.expectErr == "" {
				s.NoError(err)
			} else {
				s.EqualError(err, test.expectErr)
			}
		})
	}
}

func (s *DefaultMigratorSuite) mockRunDown(
	mockOrm *mocksorm.Orm,
	previousConnection, migrationSignature, table string,
	err error,
) {
	s.mockSchema.EXPECT().GetConnection().Return(previousConnection).Once()
	s.mockSchema.EXPECT().Orm().Return(mockOrm).Times(4)

	testConnectionMigration := &TestConnectionMigration{}
	if testConnectionMigration.Signature() == migrationSignature {
		s.mockSchema.EXPECT().SetConnection(testConnectionMigration.Connection()).Once()
	}

	mockOrm.EXPECT().Transaction(mock.Anything).RunAndReturn(func(f func(tx orm.Query) error) error {
		mockQuery := mocksorm.NewQuery(s.T())
		mockOrm.EXPECT().Query().Return(mockQuery).Once()
		mockOrm.EXPECT().SetQuery(mockQuery).Once()
		s.mockSchema.EXPECT().DropIfExists(table).Return(nil).Once()
		s.mockSchema.EXPECT().SetConnection(previousConnection).Once()
		mockOrm.EXPECT().SetQuery(mockQuery).Once()
		s.mockRepository.EXPECT().Delete(migrationSignature).Return(err).Once()

		return f(mockQuery)
	}).Once()
}

func (s *DefaultMigratorSuite) mockRunUp(
	mockOrm *mocksorm.Orm,
	previousConnection, migrationSignature, table string,
	batch int,
	err error,
) {
	s.mockSchema.EXPECT().GetConnection().Return(previousConnection).Once()
	s.mockSchema.EXPECT().Orm().Return(mockOrm).Times(4)

	testConnectionMigration := &TestConnectionMigration{}
	if testConnectionMigration.Signature() == migrationSignature {
		s.mockSchema.EXPECT().SetConnection(testConnectionMigration.Connection()).Once()
	}

	mockOrm.EXPECT().Transaction(mock.Anything).RunAndReturn(func(f func(tx orm.Query) error) error {
		mockQuery := mocksorm.NewQuery(s.T())
		mockOrm.EXPECT().Query().Return(mockQuery).Once()
		mockOrm.EXPECT().SetQuery(mockQuery).Once()
		s.mockSchema.EXPECT().Create(table, mock.Anything).Return(nil).Once()
		s.mockSchema.EXPECT().SetConnection(previousConnection).Once()
		mockOrm.EXPECT().SetQuery(mockQuery).Once()
		s.mockRepository.EXPECT().Log(migrationSignature, batch).Return(err).Once()

		return f(mockQuery)
	}).Once()
}

type TestMigration struct {
	schema contractsschema.Schema
}

func NewTestMigration(schema contractsschema.Schema) *TestMigration {
	return &TestMigration{schema: schema}
}

func (r *TestMigration) Signature() string {
	return "20240817214501_create_users_table"
}

func (r *TestMigration) Up() error {
	return r.schema.Create("users", func(table contractsschema.Blueprint) {
		table.String("name")
	})
}

func (r *TestMigration) Down() error {
	return r.schema.DropIfExists("users")
}

type TestConnectionMigration struct {
	schema contractsschema.Schema
}

func NewTestConnectionMigration(schema contractsschema.Schema) *TestConnectionMigration {
	return &TestConnectionMigration{schema: schema}
}

func (r *TestConnectionMigration) Signature() string {
	return "20240817214501_create_agents_table"
}

func (r *TestConnectionMigration) Connection() string {
	return "sqlite"
}

func (r *TestConnectionMigration) Up() error {
	return r.schema.Create("agents", func(table contractsschema.Blueprint) {
		table.String("name")
	})
}

func (r *TestConnectionMigration) Down() error {
	return r.schema.DropIfExists("agents")
}

type TestErrorMigration struct {
}

func NewTestErrorMigration() *TestErrorMigration {
	return &TestErrorMigration{}
}

func (r *TestErrorMigration) Signature() string {
	return "20240817214501_create_companies_table"
}

func (r *TestErrorMigration) Up() error {
	return assert.AnError
}

func (r *TestErrorMigration) Down() error {
	return assert.AnError
}
