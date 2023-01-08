package database

import (
	"log"
	"testing"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/database/support"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/testing/mock"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"
)

type GormQueryTestSuite struct {
	suite.Suite
	query ormcontract.Query
}

func TestGormQueryTestSuite(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	mysql, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "5.7",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=123123",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	_ = mysql.Expire(30)

	mockConfig := mock.Config()
	mockConfig.On("GetBool", "app.debug").Return(true).Once()
	mockConfig.On("GetString", "database.connections.mysql.driver").Return(support.Mysql).Once()
	mockConfig.On("GetString", "database.connections.mysql.host").Return("localhost").Once()
	mockConfig.On("GetString", "database.connections.mysql.port").Return("49154").Once()
	mockConfig.On("GetString", "database.connections.mysql.database").Return("goravel").Once()
	mockConfig.On("GetString", "database.connections.mysql.username").Return("root").Once()
	mockConfig.On("GetString", "database.connections.mysql.password").Return("123123").Once()
	mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4").Once()
	mockConfig.On("GetString", "database.connections.mysql.loc").Return("Local").Once()
	db, err := NewGormInstance("mysql")
	if err != nil {
		log.Fatalf("Get gorm instance error: %s", err)
	}
	suite.Run(t, &GormQueryTestSuite{
		query: NewGormQuery(db),
	})

	if err := pool.Purge(mysql); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	mockConfig.AssertExpectations(t)
}

func (s *GormQueryTestSuite) SetupTest() {
}

func (s *GormQueryTestSuite) TestSelect() {
	type User struct {
		orm.Model
		Name   string
		Avatar string
	}

	user := User{Name: "user"}
	s.Nil(facades.Orm.Query().Create(&user))
	s.Equal(uint(1), user.ID)

	//var user1 User
	//assert.Nil(t, facades.Orm.Query().Where("name = ?", "user").First(&user1))
	//assert.Equal(t, uint(1), user1.ID)
	//
	//var user2 User
	//assert.Nil(t, facades.Orm.Query().Find(&user2, user.ID))
	//assert.Equal(t, uint(1), user2.ID)
	//
	//var user3 []User
	//assert.Nil(t, facades.Orm.Query().Find(&user3, []uint{user.ID}))
	//assert.Equal(t, 1, len(user3))
	//
	//var user4 []User
	//assert.Nil(t, facades.Orm.Query().Where("id in ?", []uint{user.ID}).Find(&user4))
	//assert.Equal(t, 1, len(user4))
	//
	//var user5 []User
	//assert.Nil(t, facades.Orm.Query().Where("id in ?", []uint{user.ID}).Get(&user5))
	//assert.Equal(t, 1, len(user5))
}
