package factory

import (
	"github.com/goravel/framework/database/orm"
	"github.com/jaswdr/faker"
	"testing"
)

var testfake = faker.Faker{}

func init() {
	testfake = faker.New()
}

type TestUser struct {
	orm.Model
	UserName string `gorm:"column:user_name;type:varchar(255);not null" json:"user_name" form:"user_name"`
	Password string `gorm:"column:password;type:varchar(255);not null" json:"password" form:"password"`
	orm.SoftDeletes
}

type TestFactory struct {
	Factory
}

func (t *TestFactory) NewFactory() *TestFactory {
	return &TestFactory{}
}

func (t *TestFactory) Definition() map[string]interface{} {
	mapData := make(map[string]interface{})
	mapData["UserName"] = testfake.Person().Name()
	mapData["Password"] = testfake.Internet().Password()
	mapData["mobile"] = testfake.App().Version()
	return mapData
}

func TestCreate(t *testing.T) {
	var test TestUser
	var testFactory TestFactory
	//以下代码可以在XXXseeder.go中使用
	testFactory.UseFactory(test).Count(3).Create(&testFactory)
}
