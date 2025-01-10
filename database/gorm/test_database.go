package gorm

import (
	"testing"

	"github.com/goravel/framework/contracts/database/orm"
	supportdocker "github.com/goravel/framework/support/docker"
	"github.com/stretchr/testify/assert"
)

func TestCount(t *testing.T, driver orm.Driver) {
	docker, err := driver.Docker()
	if err != nil {
		panic(err)
	}

	if err := supportdocker.Ready(docker); err != nil {
		panic(err)
	}

	query := NewTestQuery(docker)

	user := User{Name: "count_user", Avatar: "count_avatar"}
	assert.Nil(t, query.Query().Create(&user))
	assert.True(t, user.ID > 0)

	user1 := User{Name: "count_user", Avatar: "count_avatar1"}
	assert.Nil(t, query.Query().Create(&user1))
	assert.True(t, user1.ID > 0)

	var count int64
	assert.Nil(t, query.Query().Model(&User{}).Where("name = ?", "count_user").Count(&count))
	assert.True(t, count > 0)

	var count1 int64
	assert.Nil(t, query.Query().Table("users").Where("name = ?", "count_user").Count(&count1))
	assert.True(t, count1 > 0)
}
