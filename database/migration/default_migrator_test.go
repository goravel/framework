package migration

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/schema"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

type DefaultDriverSuite struct {
	suite.Suite
	value          int
	mockRepository *mocksmigration.Repository
	mockSchema     *mocksmigration.Schema
	driver         *DefaultMigrator
}

func TestDefaultDriverSuite(t *testing.T) {
	suite.Run(t, &DefaultDriverSuite{})
}

func (s *DefaultDriverSuite) SetupTest() {
	s.value = 0
	s.mockRepository = mocksmigration.NewRepository(s.T())
	s.mockSchema = mocksmigration.NewSchema(s.T())

	s.driver = &DefaultMigrator{
		creator:    NewDefaultCreator(),
		repository: s.mockRepository,
		schema:     s.mockSchema,
	}
}

func (s *DefaultDriverSuite) TestCreate() {
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

func (s *DefaultDriverSuite) TestRun() {
	tests := []struct {
		name        string
		setup       func()
		expectError string
	}{
		{
			name: "Happy path",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetRan().Return([]string{"20240817214501_create_agents_table"}, nil).Once()
				s.mockSchema.EXPECT().Migrations().Return([]schema.Migration{
					&TestMigration{suite: s},
					&TestConnectionMigration{suite: s},
				}).Once()
				s.mockRepository.EXPECT().GetNextBatchNumber().Return(1, nil).Once()
				s.mockRepository.EXPECT().Log("20240817214501_create_users_table", 1).Return(nil).Once()
			},
		},
		{
			name: "Sad path - Log returns error",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetRan().Return([]string{"20240817214501_create_agents_table"}, nil).Once()
				s.mockSchema.EXPECT().Migrations().Return([]schema.Migration{
					&TestMigration{suite: s},
					&TestConnectionMigration{suite: s},
				}).Once()
				s.mockRepository.EXPECT().GetNextBatchNumber().Return(1, nil).Once()
				s.mockRepository.EXPECT().Log("20240817214501_create_users_table", 1).Return(errors.New("error")).Once()
			},
			expectError: "error",
		},
		{
			name: "Sad path - GetNextBatchNumber returns error",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetRan().Return([]string{"20240817214501_create_agents_table"}, nil).Once()
				s.mockSchema.EXPECT().Migrations().Return([]schema.Migration{
					&TestMigration{suite: s},
					&TestConnectionMigration{suite: s},
				}).Once()
				s.mockRepository.EXPECT().GetNextBatchNumber().Return(0, errors.New("error")).Once()
			},
			expectError: "error",
		},
		{
			name: "Sad path - GetRan returns error",
			setup: func() {
				s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
				s.mockRepository.EXPECT().GetRan().Return(nil, errors.New("error")).Once()
			},
			expectError: "error",
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

func (s *DefaultDriverSuite) TestPendingMigrations() {
	migrations := []schema.Migration{
		&TestMigration{suite: s},
		&TestConnectionMigration{suite: s},
	}
	ran := []string{
		"20240817214501_create_users_table",
	}

	pendingMigrations := s.driver.pendingMigrations(migrations, ran)
	s.Len(pendingMigrations, 1)
	s.Equal(&TestConnectionMigration{suite: s}, pendingMigrations[0])
}

func (s *DefaultDriverSuite) TestPrepareDatabase() {
	s.mockRepository.EXPECT().RepositoryExists().Return(true).Once()
	s.driver.prepareDatabase()

	s.mockRepository.EXPECT().RepositoryExists().Return(false).Once()
	s.mockRepository.EXPECT().CreateRepository().Once()
	s.driver.prepareDatabase()
}

func (s *DefaultDriverSuite) TestRunPending() {
	tests := []struct {
		name        string
		migrations  []schema.Migration
		setup       func()
		expectError string
	}{
		{
			name: "Happy path",
			migrations: []schema.Migration{
				&TestMigration{suite: s},
			},
			setup: func() {
				s.mockRepository.EXPECT().GetNextBatchNumber().Return(1, nil).Once()
				s.mockRepository.EXPECT().Log("20240817214501_create_users_table", 1).Return(nil).Once()
			},
		},
		{
			name:       "Happy path - no migrations",
			migrations: []schema.Migration{},
			setup:      func() {},
		},
		{
			name: "Sad path - GetNextBatchNumber returns error",
			migrations: []schema.Migration{
				&TestMigration{suite: s},
			},
			setup: func() {
				s.mockRepository.EXPECT().GetNextBatchNumber().Return(0, errors.New("error")).Once()
			},
			expectError: "error",
		},
		{
			name: "Sad path - runUp returns error",
			migrations: []schema.Migration{
				&TestMigration{suite: s},
			},
			setup: func() {
				s.mockRepository.EXPECT().GetNextBatchNumber().Return(1, nil).Once()
				s.mockRepository.EXPECT().Log("20240817214501_create_users_table", 1).Return(errors.New("error")).Once()
			},
			expectError: "error",
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

func (s *DefaultDriverSuite) TestRunUp() {
	batch := 1
	s.mockRepository.EXPECT().Log("20240817214501_create_users_table", batch).Return(nil).Once()
	s.NoError(s.driver.runUp(&TestMigration{
		suite: s,
	}, batch))
	s.Equal(1, s.value)

	previousConnection := "postgres"
	s.mockSchema.EXPECT().GetConnection().Return(previousConnection).Once()
	s.mockSchema.EXPECT().SetConnection("mysql").Once()
	s.mockSchema.EXPECT().SetConnection(previousConnection).Once()
	s.mockRepository.EXPECT().Log("20240817214501_create_agents_table", batch).Return(nil).Once()
	s.NoError(s.driver.runUp(&TestConnectionMigration{
		suite: s,
	}, batch))
	s.Equal(2, s.value)
}

type TestMigration struct {
	suite *DefaultDriverSuite
}

func (s *TestMigration) Signature() string {
	return "20240817214501_create_users_table"
}

func (s *TestMigration) Up() error {
	s.suite.value++

	return nil
}

func (s *TestMigration) Down() error {
	return nil
}

type TestConnectionMigration struct {
	suite *DefaultDriverSuite
}

func (s *TestConnectionMigration) Signature() string {
	return "20240817214501_create_agents_table"
}

func (s *TestConnectionMigration) Connection() string {
	return "mysql"
}

func (s *TestConnectionMigration) Up() error {
	s.suite.value++

	return nil
}

func (s *TestConnectionMigration) Down() error {
	return nil
}
