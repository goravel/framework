package schema

import (
	"context"

	"github.com/stretchr/testify/mock"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	databaseorm "github.com/goravel/framework/database/orm"
	mockslog "github.com/goravel/framework/mocks/log"
)

func GetTestSchema(testQuery *gorm.TestQuery, driverToTestQuery map[contractsdatabase.Driver]*gorm.TestQuery) *Schema {
	queries := make(map[string]contractsorm.Query)
	for driver, testQuery := range driverToTestQuery {
		queries[driver.String()] = testQuery.Query()
	}

	mockLog := &mockslog.Log{}
	mockLog.EXPECT().Errorf(mock.Anything).Maybe()
	orm := databaseorm.NewOrm(context.Background(), testQuery.MockConfig(), testQuery.Docker().Driver().String(), testQuery.Query(), queries, mockLog, nil, nil)
	schema := NewSchema(testQuery.MockConfig(), mockLog, orm, nil)

	return schema
}
