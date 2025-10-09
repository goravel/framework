package migration

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/contracts/database/orm"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

type MigratorSuite struct {
	suite.Suite
	mockArtisan    *mocksconsole.Artisan
	mockRepository *mocksmigration.Repository
	mockSchema     *mocksschema.Schema
	migrator       *Migrator
}

func TestMigratorSuite(t *testing.T) {
	suite.Run(t, &MigratorSuite{})
}

func (s *MigratorSuite) SetupTest() {
	s.mockArtisan = mocksconsole.NewArtisan(s.T())
	s.mockRepository = mocksmigration.NewRepository(s.T())
	s.mockSchema = mocksschema.NewSchema(s.T())

	s.migrator = &Migrator{
		artisan:    s.mockArtisan,
		creator:    NewCreator(),
		repository: s.mockRepository,
		schema:     s.mockSchema,
	}
}

func (s *MigratorSuite) TestCreate() {
	now := carbon.FromDateTime(2024, 8, 17, 21, 45, 1)
	carbon.SetTestNow(now)

	pwd, err := os.Getwd()
	s.NoError(err)

	path := filepath.Join(pwd, "database", "migrations")
	name := "create_users_table"

	fileName, err := s.migrator.Create(name, "")
	s.NoError(err)
	s.Equal("20240817214501_"+name, fileName)

	migrationFile := filepath.Join(path, "20240817214501_"+name+".go")
	s.True(file.Exists(migrationFile))

	defer func() {
		carbon.ClearTestNow()
		s.NoError(file.Remove("database"))
	}()
}

func (s *MigratorSuite) TestFresh() {
	// Success
	s.mockArtisan.EXPECT().Call("db:wipe --force").Return(nil).Once()
	s.mockArtisan.EXPECT().Call("migrate").Return(nil).Once()

	s.NoError(s.migrator.Fresh())

	// db:wipe returns error
	s.mockArtisan.EXPECT().Call("db:wipe --force").Return(assert.AnError).Once()

	s.EqualError(s.migrator.Fresh(), assert.AnError.Error())

	// migrate returns error
	s.mockArtisan.EXPECT().Call("db:wipe --force").Return(nil).Once()
	s.mockArtisan.EXPECT().Call("migrate").Return(assert.AnError).Once()

	s.EqualError(s.migrator.Fresh(), assert.AnError.Error())
}

func (s *MigratorSuite) TestGetFilesForRollback() {
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
				s.mockRepository.EXPECT().GetMigrationsByStep(1).Return([]migration.File{{Migration: "20240817214501_create_users_table"}}, nil).Once()
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
			name:  "Returns error when GetMigrationsByStep fails",
			step:  1,
			batch: 0,
			setup: func() {
				s.mockRepository.EXPECT().GetMigrationsByStep(1).Return(nil, errors.New("error")).Once()
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

			files, err := s.migrator.getFilesForRollback(test.step, test.batch)
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

func (s *MigratorSuite) TestPendingMigrations() {
	testMigration := NewTestMigration(s.mockSchema)
	testConnectionMigration := NewTestConnectionMigration(s.mockSchema)
	testErrorMigration := NewTestErrorMigration()

	s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{
		testMigration,
		testConnectionMigration,
		testErrorMigration,
	}).Once()

	ran := []string{
		"20240817214501_create_users_table",
	}

	pendingMigrations := s.migrator.pendingMigrations(ran)
	s.Len(pendingMigrations, 2)
	s.Equal(testConnectionMigration, pendingMigrations[0])
	s.Equal(testErrorMigration, pendingMigrations[1])
}

func (s *MigratorSuite) TestPrepareDatabase() {
	s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
	s.NoError(s.migrator.prepareDatabase())

	s.mockRepository.EXPECT().RepositoryExists().Return(false).Once()
	s.mockRepository.EXPECT().CreateRepository().Return(nil).Once()
	s.NoError(s.migrator.prepareDatabase())
}

func (s *MigratorSuite) TestPrintTitle() {
	s.Equal("\x1b[39mMigration name      \x1b[0m\x1b[39m | Batch / Status\x1b[0m\n\x1b[39m\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m-\x1b[0m\x1b[39m\x1b[0m\n\x1b[39m\x1b[0m", color.CaptureOutput(func(w io.Writer) {
		s.migrator.printTitle(20)
	}))
}

func (s *MigratorSuite) TestReset() {
	tests := []struct {
		name      string
		setup     func()
		expectErr string
	}{
		{
			name: "Get ran failed",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(false).Once()
			},
		},
		{
			name: "failed to get ran",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetRan().Return(nil, assert.AnError).Once()
			},
			expectErr: assert.AnError.Error(),
		},
		{
			name: "happy path",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Twice()

				previousConnection := "postgres"
				testMigration := NewTestMigration(s.mockSchema)
				s.mockRepository.EXPECT().GetRan().Return([]string{testMigration.Signature()}, nil).Once()

				s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{
					testMigration,
				})
				s.mockRepository.EXPECT().GetMigrationsByStep(1).Return([]migration.File{{Migration: testMigration.Signature()}}, nil).Once()

				mockOrm := mocksorm.NewOrm(s.T())
				s.mockRunDown(mockOrm, previousConnection, testMigration.Signature(), "users", nil)
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()

			err := s.migrator.Reset()

			if test.expectErr == "" {
				s.NoError(err)
			} else {
				s.EqualError(err, test.expectErr)
			}
		})
	}
}

func (s *MigratorSuite) TestRollback() {
	tests := []struct {
		name      string
		setup     func()
		expectErr string
	}{
		{
			name: "happy path - no files",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetMigrationsByStep(1).Return(nil, nil).Once()
			},
		},
		{
			name: "happy path",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()

				previousConnection := "postgres"
				testMigration := NewTestMigration(s.mockSchema)

				s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{
					testMigration,
				})
				s.mockRepository.EXPECT().GetMigrationsByStep(1).Return([]migration.File{{Migration: testMigration.Signature()}}, nil).Once()

				mockOrm := mocksorm.NewOrm(s.T())
				s.mockRunDown(mockOrm, previousConnection, testMigration.Signature(), "users", nil)
			},
		},
		{
			name: "happy path - missing migration",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetMigrationsByStep(1).Return([]migration.File{{Migration: "20240817214502_create_users_table"}}, nil).Once()
			},
		},
		{
			name: "failed to rollback",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()

				previousConnection := "postgres"
				testMigration := NewTestMigration(s.mockSchema)

				s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{
					testMigration,
				})
				s.mockRepository.EXPECT().GetMigrationsByStep(1).Return([]migration.File{{Migration: testMigration.Signature()}}, nil).Once()

				mockOrm := mocksorm.NewOrm(s.T())
				s.mockRunDown(mockOrm, previousConnection, testMigration.Signature(), "users", assert.AnError)
			},
			expectErr: assert.AnError.Error(),
		},
		{
			name: "failed to get migrations by step",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetMigrationsByStep(1).Return(nil, assert.AnError).Once()
			},
			expectErr: assert.AnError.Error(),
		},
		{
			name: "repository doesn't exist",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(false).Once()
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()

			err := s.migrator.Rollback(1, 0)

			if test.expectErr == "" {
				s.NoError(err)
			} else {
				s.EqualError(err, test.expectErr)
			}
		})
	}
}

func (s *MigratorSuite) TestRun() {
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
				s.mockRepository.EXPECT().GetNextBatchNumber().Return(1, nil).Once()
				s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{
					testMigration,
					testConnectionMigration,
				}).Once()

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

			err := s.migrator.Run()
			if test.expectError == "" {
				s.Nil(err)
			} else {
				s.EqualError(err, test.expectError)
			}
		})
	}
}

func (s *MigratorSuite) TestRunDown() {
	var (
		previousConnection      = "postgres"
		testMigration           = NewTestMigration(s.mockSchema)
		testErrorMigration      = NewTestErrorMigration()
		testConnectionMigration = NewTestConnectionMigration(s.mockSchema)

		mockOrm *mocksorm.Orm
	)

	beforeEach := func() {
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

				mockQuery := mocksorm.NewQuery(s.T())
				mockOrm.EXPECT().Query().Return(mockQuery).Once()

				mockOrm.EXPECT().Transaction(mock.Anything).RunAndReturn(func(f func(tx orm.Query) error) error {
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

			err := s.migrator.runDown(test.migration)

			if test.expectErr == "" {
				s.NoError(err)
			} else {
				s.EqualError(err, test.expectErr)
			}
		})
	}
}

func (s *MigratorSuite) TestRunPending() {
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

			err := s.migrator.runPending(test.migrations)
			if test.expectError == "" {
				s.Nil(err)
			} else {
				s.EqualError(err, test.expectError)
			}
		})
	}
}

func (s *MigratorSuite) TestRunUp() {
	var (
		batch                   = 1
		previousConnection      = "postgres"
		testMigration           = NewTestMigration(s.mockSchema)
		testErrorMigration      = NewTestErrorMigration()
		testConnectionMigration = NewTestConnectionMigration(s.mockSchema)

		mockOrm *mocksorm.Orm
	)

	beforeEach := func() {
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

				mockQuery := mocksorm.NewQuery(s.T())
				mockOrm.EXPECT().Query().Return(mockQuery).Once()

				mockOrm.EXPECT().Transaction(mock.Anything).RunAndReturn(func(f func(tx orm.Query) error) error {
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

			err := s.migrator.runUp(test.migration, batch)

			if test.expectErr == "" {
				s.NoError(err)
			} else {
				s.EqualError(err, test.expectErr)
			}
		})
	}
}

func (s *MigratorSuite) TestStatus() {
	tests := []struct {
		name        string
		setup       func()
		assert      func()
		expectError string
	}{
		{
			name: "The migration table doesn't exist",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(false).Once()
			},
			assert: func() {
				s.Equal("\x1b[30;43m\x1b[30;43m WARNING \x1b[0m\x1b[0m \x1b[33m\x1b[33mMigration table not found\x1b[0m\x1b[0m\n", color.CaptureOutput(func(w io.Writer) {
					status, err := s.migrator.Status()
					s.NoError(err)
					s.Nil(status)
				}))
			},
		},
		{
			name: "Get migration batches failed",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetMigrations().Return(nil, assert.AnError).Once()
			},
			assert: func() {
				status, err := s.migrator.Status()
				s.EqualError(err, assert.AnError.Error())
				s.Nil(status)
			},
		},
		{
			name: "No migrations found",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockSchema.EXPECT().Migrations().Return(nil).Once()
				s.mockRepository.EXPECT().GetMigrations().Return(nil, nil).Once()
			},
			assert: func() {
				s.Equal("\x1b[30;43m\x1b[30;43m WARNING \x1b[0m\x1b[0m \x1b[33m\x1b[33mNo migrations found\x1b[0m\x1b[0m\n", color.CaptureOutput(func(w io.Writer) {
					status, err := s.migrator.Status()
					s.NoError(err)
					s.Len(status, 0)
				}))
			},
		},
		{
			name: "Success",
			setup: func() {
				testMigration := NewTestMigration(s.mockSchema)
				testConnectionMigration := NewTestConnectionMigration(s.mockSchema)

				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockSchema.EXPECT().Migrations().Return([]contractsschema.Migration{
					testMigration,
					testConnectionMigration,
				}).Once()
				s.mockRepository.EXPECT().GetMigrations().Return([]migration.File{
					{ID: 1, Migration: testMigration.Signature(), Batch: 1},
				}, nil).Once()
			},
			assert: func() {
				status, err := s.migrator.Status()
				s.NoError(err)
				s.Len(status, 2)
				s.ElementsMatch(status, []migration.Status{
					{
						Name:  NewTestMigration(s.mockSchema).Signature(),
						Batch: 1,
						Ran:   true,
					},
					{
						Name: NewTestConnectionMigration(s.mockSchema).Signature(),
					},
				})
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			test.setup()
			test.assert()
		})
	}
}

func (s *MigratorSuite) mockRunDown(
	mockOrm *mocksorm.Orm,
	previousConnection, migrationSignature, table string,
	err error,
) {
	s.mockSchema.EXPECT().GetConnection().Return(previousConnection).Once()
	s.mockSchema.EXPECT().Orm().Return(mockOrm).Times(5)

	mockQuery := mocksorm.NewQuery(s.T())
	mockOrm.EXPECT().Query().Return(mockQuery).Once()

	testConnectionMigration := &TestConnectionMigration{}
	if testConnectionMigration.Signature() == migrationSignature {
		s.mockSchema.EXPECT().SetConnection(testConnectionMigration.Connection()).Once()
	}

	mockOrm.EXPECT().Transaction(mock.Anything).RunAndReturn(func(f func(tx orm.Query) error) error {
		mockOrm.EXPECT().SetQuery(mockQuery).Once()
		s.mockSchema.EXPECT().DropIfExists(table).Return(nil).Once()
		s.mockSchema.EXPECT().SetConnection(previousConnection).Twice()
		mockOrm.EXPECT().SetQuery(mockQuery).Twice()
		s.mockRepository.EXPECT().Delete(migrationSignature).Return(err).Once()

		return f(mockQuery)
	}).Once()
}

func (s *MigratorSuite) mockRunUp(
	mockOrm *mocksorm.Orm,
	previousConnection, migrationSignature, table string,
	batch int,
	err error,
) {
	s.mockSchema.EXPECT().GetConnection().Return(previousConnection).Once()
	s.mockSchema.EXPECT().Orm().Return(mockOrm).Times(5)

	mockQuery := mocksorm.NewQuery(s.T())
	mockOrm.EXPECT().Query().Return(mockQuery).Once()

	testConnectionMigration := &TestConnectionMigration{}
	if testConnectionMigration.Signature() == migrationSignature {
		s.mockSchema.EXPECT().SetConnection(testConnectionMigration.Connection()).Once()
	}

	mockOrm.EXPECT().Transaction(mock.Anything).RunAndReturn(func(f func(tx orm.Query) error) error {
		mockOrm.EXPECT().SetQuery(mockQuery).Once()
		s.mockSchema.EXPECT().Create(table, mock.Anything).Return(nil).Once()
		s.mockSchema.EXPECT().SetConnection(previousConnection).Twice()
		mockOrm.EXPECT().SetQuery(mockQuery).Twice()
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
	return "20240817214502_create_agents_table"
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
	return "20240817214503_create_companies_table"
}

func (r *TestErrorMigration) Up() error {
	return assert.AnError
}

func (r *TestErrorMigration) Down() error {
	return assert.AnError
}
