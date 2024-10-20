package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	databasedb "github.com/goravel/framework/database/db"
	"github.com/goravel/framework/database/gorm"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type SqlDriverSuite struct {
	suite.Suite
	mockConfig        *mocksconfig.Config
	driverToTestQuery map[contractsdatabase.Driver]*gorm.TestQuery
}

func TestSqlDriverSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, &SqlDriverSuite{})
}

func (s *SqlDriverSuite) SetupSuite() {
	s.driverToTestQuery = gorm.NewTestQueries().Queries()
}

func (s *SqlDriverSuite) SetupTest() {

}

func (s *SqlDriverSuite) TearDownTest() {
	s.NoError(file.Remove("database"))
}

func (s *SqlDriverSuite) TestCreate() {
	now := carbon.FromDateTime(2024, 8, 17, 21, 45, 1)
	carbon.SetTestNow(now)

	pwd, err := os.Getwd()
	s.NoError(err)

	path := filepath.Join(pwd, "database", "migrations")
	name := "create_users_table"

	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockConfig.EXPECT().GetString("database.default").Return("postgres").Once()
	s.mockConfig.EXPECT().GetString("database.connections.postgres.driver").Return("postgres").Once()
	s.mockConfig.EXPECT().GetString("database.connections.postgres.charset").Return("utf8mb4").Once()
	s.mockConfig.EXPECT().GetString("database.migrations.table").Return("migrations").Once()

	driver := NewSqlMigrator(s.mockConfig)

	s.NoError(driver.Create(name))

	upFile := filepath.Join(path, "20240817214501_"+name+".up.sql")
	downFile := filepath.Join(path, "20240817214501_"+name+".down.sql")

	s.True(file.Exists(upFile))
	s.True(file.Exists(downFile))

	defer carbon.UnsetTestNow()
}

func (s *SqlDriverSuite) TestFresh() {

}

func (s *SqlDriverSuite) TestRun() {
	testQueries := gorm.NewTestQueries().Queries()
	for driver, testQuery := range testQueries {
		query := testQuery.Query()
		mockConfig := testQuery.MockConfig()
		CreateTestMigrations(driver)

		sqlDriver := &SqlMigrator{
			configBuilder: databasedb.NewConfigBuilder(mockConfig, driver.String()),
			creator:       NewSqlCreator(driver, "utf8mb4"),
			table:         "migrations",
		}
		err := sqlDriver.Run()
		s.NoError(err)

		var agent Agent
		s.NoError(query.Where("name", "goravel").First(&agent))
		s.True(agent.ID > 0)

		err = sqlDriver.Run()
		s.NoError(err)
	}
}
