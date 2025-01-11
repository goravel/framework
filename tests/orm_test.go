package tests

import (
	"testing"

	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/foundation"
	supportdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/postgres"
	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/tests/config"
)

func TestCount(t *testing.T) {
	app := foundation.NewApplication()
	app.Boot()
	config.Boot()

	configFacade := app.MakeConfig()
	driver := postgres.NewPostgres(postgres.NewConfigBuilder(configFacade, "postgres"), app.MakeLog())
	docker, err := driver.Docker()
	if err != nil {
		panic(err)
	}

	if err := docker.Build(); err != nil {
		panic(err)
	}

	if err := supportdocker.Ready(docker); err != nil {
		panic(err)
	}

	configFacade.Add("database.connections.postgres.port", docker.Config().Port)
	app.MakeOrm().Refresh()
	query := gorm.NewTestQuery1(docker, configFacade)
	query.CreateTable(gorm.TestTableUsers)

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
}
