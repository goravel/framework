package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

type SqlDriverSuite struct {
	suite.Suite
	mockConfig *mocksconfig.Config
	mockSchema *mocksmigration.Schema
	driver     *SqlDriver
}

func TestSqlDriverSuite(t *testing.T) {
	suite.Run(t, &SqlDriverSuite{})
}

func (s *SqlDriverSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockConfig.EXPECT().GetString("database.default").Return("postgres").Once()
	s.mockConfig.EXPECT().GetString("database.connections.postgres.driver").Return("postgres").Once()
	s.mockConfig.EXPECT().GetString("database.connections.postgres.charset").Return("utf8mb4").Once()
	s.mockConfig.EXPECT().GetString("database.migrations.table").Return("migrations").Once()
	s.mockSchema = mocksmigration.NewSchema(s.T())

	s.driver = NewSqlDriver(s.mockConfig)
}

func (s *SqlDriverSuite) TestCreate() {
	now := carbon.FromDateTime(2024, 8, 17, 21, 45, 1)
	carbon.SetTestNow(now)

	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "database", "migrations")
	name := "create_users_table"

	s.NoError(s.driver.Create(name))

	upFile := filepath.Join(path, "20240817214501_"+name+".up.sql")
	downFile := filepath.Join(path, "20240817214501_"+name+".down.sql")

	s.True(file.Exists(upFile))
	s.True(file.Exists(downFile))

	defer func() {
		carbon.UnsetTestNow()
		s.NoError(file.Remove("database"))
	}()
}
