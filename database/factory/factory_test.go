package factory

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/database/factory"
	"github.com/goravel/framework/support/carbon"
)

type BaseModel struct {
	ID uint `gorm:"primaryKey" json:"id"`
	NullableTimestamps
}

type NullableSoftDeletes struct {
	DeletedAt *gormio.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

type NullableTimestamps struct {
	CreatedAt *carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt *carbon.DateTime `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}

type User struct {
	BaseModel
	NullableSoftDeletes
	Name   string
	Avatar string
}

func (u *User) Factory() factory.Factory {
	return &UserFactory{}
}

type UserFactory struct {
}

func (u *UserFactory) Definition() map[string]any {
	faker := gofakeit.New(0)
	return map[string]any{
		"Name":      faker.Name(),
		"Avatar":    faker.Email(),
		"CreatedAt": carbon.NewDateTime(carbon.Now()),
		"UpdatedAt": carbon.NewDateTime(carbon.Now()),
		"DeletedAt": gormio.DeletedAt{Time: time.Now(), Valid: true},
	}
}

type House struct {
	BaseModel
	Name          string
	HouseableID   uint
	HouseableType string
}

func TestGetRawAttributes(t *testing.T) {
	var house House
	attributes, err := getRawAttributes(&house)
	assert.NotNil(t, err)
	assert.Nil(t, attributes)

	var user User
	attributes, err = getRawAttributes(&user)
	assert.Nil(t, err)
	assert.NotNil(t, attributes)

	var user1 User
	attributes, err = getRawAttributes(&user1, map[string]any{
		"Avatar": "avatar",
	})
	assert.Nil(t, err)
	assert.NotNil(t, attributes)
	assert.True(t, len(attributes["Name"].(string)) > 0)
	assert.Equal(t, "avatar", attributes["Avatar"].(string))
	assert.NotNil(t, attributes["CreatedAt"])
	assert.NotNil(t, attributes["UpdatedAt"])
	assert.NotNil(t, attributes["DeletedAt"])
}
