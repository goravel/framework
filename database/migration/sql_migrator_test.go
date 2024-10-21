package migration

import (
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
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type SqlMigratorSuite struct {
	suite.Suite
	driverToTestQuery map[contractsdatabase.Driver]*gorm.TestQuery
}

func TestSqlMigratorSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, &SqlMigratorSuite{})
}

func (s *SqlMigratorSuite) SetupSuite() {
	s.driverToTestQuery = gorm.NewTestQueries().Queries()
}

func (s *SqlMigratorSuite) SetupTest() {

}

func (s *SqlMigratorSuite) TearDownTest() {
	s.NoError(file.Remove("database"))
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
		migrator, query := getTestSqlMigrator(s.T(), driver, testQuery)

		err := migrator.Run()
		s.NoError(err)

		err = migrator.Fresh()
		s.NoError(err)

		var agent Agent
		err = query.Where("name", "goravel").First(&agent)
		s.NoError(err)
		s.True(agent.ID > 0)
	}
}

func (s *SqlMigratorSuite) TestRun() {
	for driver, testQuery := range s.driverToTestQuery {
		migrator, query := getTestSqlMigrator(s.T(), driver, testQuery)

		err := migrator.Run()
		s.NoError(err)

		var agent Agent
		s.NoError(query.Where("name", "goravel").First(&agent))
		s.True(agent.ID > 0)

		err = migrator.Run()
		s.NoError(err)
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
