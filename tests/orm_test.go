package tests

import (
	"testing"

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
}
