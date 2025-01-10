package tests

import (
	"testing"

<<<<<<< HEAD
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/database/gorm"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/testing/utils"
	"github.com/goravel/postgres"
	"github.com/stretchr/testify/suite"
)

type OrmTestSuite struct {
	suite.Suite
	queries map[database.Driver]*gorm.TestQuery1
}

func TestOrmTestSuite(t *testing.T) {
	suite.Run(t, &OrmTestSuite{
		queries: make(map[database.Driver]*gorm.TestQuery1),
	})
}

func (s *OrmTestSuite) SetupSuite() {
	mockConfig := mockPostgres(5432)
	driver := postgres.NewPostgres(postgres.NewConfigBuilder(mockConfig, "postgres"), utils.NewTestLog())
=======
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/foundation"
	supportdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T) {
	app := foundation.NewApplication()
	app.Boot()
	app.MakeConfig().Add("app", map[string]any{
		"name": "goravel",
		"env":  "testing",
		"key":  "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456",
	})
	app.MakeConfig().Add("database", map[string]any{
		"default": "postgres",
		"connections": map[string]any{
			"postgres": map[string]any{
				"driver":   "postgres",
				"host":     "127.0.0.1",
				"port":     5432,
				"database": "goravel_test",
				"username": "goravel",
				"password": "Framework!123",
				"via": func() (orm.Driver, error) {
					return postgres.NewPostgres(postgres.NewConfigBuilder(app.MakeConfig(), "postgres"), app.MakeLog()), nil
				},
			},
		},
	})

	driver := postgres.NewPostgres(postgres.NewConfigBuilder(app.MakeConfig(), "postgres"), app.MakeLog())
>>>>>>> 98751f8 (test)
	docker, err := driver.Docker()
	if err != nil {
		panic(err)
	}
<<<<<<< HEAD
	if err := docker.Build(); err != nil {
		panic(err)
	}
	if err := docker.Ready(); err != nil {
		panic(err)
	}

	mockConfig = mockPostgres(docker.Config().Port)
	driver = postgres.NewPostgres(postgres.NewConfigBuilder(mockConfig, "postgres"), utils.NewTestLog())

	testQuery, err := gorm.NewTestQuery1(driver, mockConfig)
	if err != nil {
		panic(err)
	}

	testQuery.CreateTable()
	s.queries[database.DriverPostgres] = testQuery
}

func (s *OrmTestSuite) SetupTest() {
}

func (s *OrmTestSuite) TestCount() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := gorm.User{Name: "count_user", Avatar: "count_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := gorm.User{Name: "count_user", Avatar: "count_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var count int64
			s.Nil(query.Query().Model(&gorm.User{}).Where("name = ?", "count_user").Count(&count))
			s.True(count > 0)

			var count1 int64
			s.Nil(query.Query().Table("users").Where("name = ?", "count_user").Count(&count1))
			s.True(count1 > 0)
		})
	}
}

func mockPostgres(port int) *mocksconfig.Config {
	username := "postgres"
	password := "Framework!123"
	database := "postgres"
	host := "localhost"
	mockConfig := &mocksconfig.Config{}

	mockConfig.EXPECT().GetBool("app.debug").Return(true)
	mockConfig.EXPECT().GetInt("database.slow_threshold", 200).Return(200)
	mockConfig.EXPECT().GetInt("database.pool.max_idle_conns", 10).Return(10)
	mockConfig.EXPECT().GetInt("database.pool.max_open_conns", 100).Return(100)
	mockConfig.EXPECT().GetInt("database.pool.conn_max_idletime", 3600).Return(3600)
	mockConfig.EXPECT().GetInt("database.pool.conn_max_lifetime", 3600).Return(3600)

	mockConfig.EXPECT().Get("database.connections.postgres.read").Return(nil)
	mockConfig.EXPECT().Get("database.connections.postgres.write").Return(nil)
	mockConfig.EXPECT().GetString("database.connections.postgres.host").Return(host)
	mockConfig.EXPECT().GetInt("database.connections.postgres.port").Return(port)
	mockConfig.EXPECT().GetString("database.connections.postgres.username").Return(username)
	mockConfig.EXPECT().GetString("database.connections.postgres.password").Return(password)
	mockConfig.EXPECT().GetString("database.connections.postgres.database").Return(database)
	mockConfig.EXPECT().GetString("database.connections.postgres.sslmode").Return("disable")
	mockConfig.EXPECT().GetString("database.connections.postgres.timezone").Return("UTC")
	mockConfig.EXPECT().GetString("database.connections.postgres.prefix").Return("")
	mockConfig.EXPECT().GetBool("database.connections.postgres.singular").Return(false)
	mockConfig.EXPECT().GetBool("database.connections.postgres.no_lower_case").Return(false)
	mockConfig.EXPECT().GetString("database.connections.postgres.dsn").Return("")
	mockConfig.EXPECT().GetString("database.connections.postgres.schema", "public").Return("public")
	mockConfig.EXPECT().Get("database.connections.postgres.name_replacer").Return(nil)

	return mockConfig
=======

	if err := docker.Build(); err != nil {
		panic(err)
	}

	if err := supportdocker.Ready(docker); err != nil {
		panic(err)
	}

	query := gorm.NewTestQuery(docker)

	user := gorm.User{Name: "count_user", Avatar: "count_avatar"}
	assert.Nil(t, query.Query().Create(&user))
	assert.True(t, user.ID > 0)

	user1 := gorm.User{Name: "count_user", Avatar: "count_avatar1"}
	assert.Nil(t, query.Query().Create(&user1))
	assert.True(t, user1.ID > 0)

	var count int64
	assert.Nil(t, query.Query().Model(&gorm.User{}).Where("name = ?", "count_user").Count(&count))
	assert.True(t, count > 0)

	var count1 int64
	assert.Nil(t, query.Query().Table("users").Where("name = ?", "count_user").Count(&count1))
	assert.True(t, count1 > 0)
>>>>>>> 98751f8 (test)
}
