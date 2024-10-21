package schema

import (
	"testing"

	"github.com/goravel/framework/database/gorm"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
)

func GetTestSchema(t *testing.T, testQuery *gorm.TestQuery) (*Schema, *mocksorm.Orm) {
	mockOrm := mocksorm.NewOrm(t)
	mockOrm.EXPECT().Name().Return(testQuery.Docker().Driver().String()).Twice()
	schema := NewSchema(testQuery.MockConfig(), nil, mockOrm, nil)

	return schema, mockOrm
}
