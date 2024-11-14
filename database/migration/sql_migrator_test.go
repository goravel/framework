package migration

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
	databasedb "github.com/goravel/framework/database/db"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type SqlMigratorSuite struct {
	suite.Suite
	driverToTestQuery map[contractsdatabase.Driver]*gorm.TestQuery
}

func TestSqlMigratorSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	suite.Run(t, &SqlMigratorSuite{})
}

func (s *SqlMigratorSuite) SetupTest() {
	s.driverToTestQuery = gorm.NewTestQueries().Queries()
}

func (s *SqlMigratorSuite) TearDownTest() {
	s.NoError(file.Remove("database"))
	if s.driverToTestQuery[contractsdatabase.DriverSqlite] != nil {
		s.NoError(s.driverToTestQuery[contractsdatabase.DriverSqlite].Docker().Stop())
	}
}

func (s *SqlMigratorSuite) TestCreate() {
	now := carbon.FromDateTime(2024, 8, 17, 21, 45, 1)
	carbon.SetTestNow(now)
	defer carbon.UnsetTestNow()

	pwd, err := os.Getwd()
	s.NoError(err)

	path := filepath.Join(pwd, "database", "migrations")
	name := "create_users_table"

	for driver, testQuery := range s.driverToTestQuery {
		migrator, _ := getTestSqlMigrator(s.T(), driver, testQuery)

		s.NoError(migrator.Create(name))

		upFile := filepath.Join(path, "20240817214501_"+name+".up.sql")
		downFile := filepath.Join(path, "20240817214501_"+name+".down.sql")

		s.True(file.Exists(upFile))
		s.True(file.Exists(downFile))
	}
}

func (s *SqlMigratorSuite) TestFresh() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			migrator, query := getTestSqlMigrator(s.T(), driver, testQuery)

			err := migrator.Run()
			s.NoError(err)

			err = migrator.Fresh()
			s.NoError(err)

			var agent Agent
			err = query.Where("name", "goravel").First(&agent)
			s.NoError(err)
			s.True(agent.ID > 0)
		})
	}
}

func (s *SqlMigratorSuite) TestRollback() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			migrator, query := getTestSqlMigrator(s.T(), driver, testQuery)

			err := migrator.Run()
			s.NoError(err)

			var agent Agent
			err = query.Where("name", "goravel").First(&agent)
			s.NoError(err)
			s.True(agent.ID > 0)

			err = migrator.Rollback(1, 0)
			s.NoError(err)

			var agent1 Agent
			err = query.Where("name", "goravel").First(&agent1)
			s.NotNil(err)
		})
	}
}

func (s *SqlMigratorSuite) TestRun() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			migrator, query := getTestSqlMigrator(s.T(), driver, testQuery)

			err := migrator.Run()
			s.NoError(err)

			var agent Agent
			s.NoError(query.Where("name", "goravel").First(&agent))
			s.True(agent.ID > 0)

			err = migrator.Run()
			s.NoError(err)
		})
	}
}

func (s *SqlMigratorSuite) TestStatus() {
	for driver, testQuery := range s.driverToTestQuery {
		s.Run(driver.String(), func() {
			migrator, _ := getTestSqlMigrator(s.T(), driver, testQuery)

			s.Equal("\x1b[30;43m\x1b[30;43m WARNING \x1b[0m\x1b[0m \x1b[33m\x1b[33mNo migrations found\x1b[0m\x1b[0m\n", color.CaptureOutput(func(w io.Writer) {
				err := migrator.Status()
				s.NoError(err)
			}))

			err := migrator.Run()
			s.NoError(err)

			s.Equal("\x1b[30;42m\x1b[30;42m SUCCESS \x1b[0m\x1b[0m \x1b[32m\x1b[32mMigration version: 20230311160527\x1b[0m\x1b[0m\n", color.CaptureOutput(func(w io.Writer) {
				err := migrator.Status()
				s.NoError(err)
			}))
		})
	}
}

func getTestSqlMigrator(t *testing.T, driver contractsdatabase.Driver, testQuery *gorm.TestQuery) (*SqlMigrator, orm.Query) {
	query := testQuery.Query()
	mockConfig := testQuery.MockConfig()
	CreateTestMigrations(driver)

	table := "migrations"
	configBuilder := databasedb.NewConfigBuilder(mockConfig, driver.String())
	migrator, err := getMigrator(configBuilder, table)
	require.NoError(t, err)

	return &SqlMigrator{
		configBuilder: databasedb.NewConfigBuilder(mockConfig, driver.String()),
		creator:       NewSqlCreator(driver, "utf8mb4"),
		migrator:      migrator,
		table:         table,
	}, query
}
