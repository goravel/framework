package console

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
	"github.com/goravel/framework/support/color"
)

func TestShowCommand(t *testing.T) {
	var (
		mockContext *mocksconsole.Context
		mockConfig  *mocksconfig.Config
		mockSchema  *mocksschema.Schema
		mockOrm     *mocksorm.Orm
		mockQuery   *mocksorm.Query
	)

	beforeEach := func() {
		mockContext = mocksconsole.NewContext(t)
		mockConfig = mocksconfig.NewConfig(t)
		mockSchema = mocksschema.NewSchema(t)
		mockOrm = mocksorm.NewOrm(t)
		mockQuery = mocksorm.NewQuery(t)
	}

	tests := []struct {
		name     string
		setup    func()
		expected string
	}{
		{
			name: "invalid argument",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("test")
				mockContext.EXPECT().Error(mock.Anything).Run(func(message string) {
					color.Errorln(message)
				})
			},
			expected: "No arguments expected for 'db:show' command, got 'test'.",
		},
		{
			name: "get tables failed",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("")
				mockContext.EXPECT().Option("database").Return("")
				mockSchema.EXPECT().Connection(mock.Anything).Return(mockSchema)
				mockSchema.EXPECT().GetConnection().Return("test")
				mockConfig.EXPECT().GetString(mock.Anything).Return("test")
				mockQuery.EXPECT().Driver().Return(database.DriverMysql)
				mockOrm.EXPECT().Query().Return(mockQuery)
				mockSchema.EXPECT().Orm().Return(mockOrm)
				mockQuery.EXPECT().Raw(mock.Anything).Return(mockQuery)
				mockQuery.EXPECT().Scan(mock.Anything).Return(nil)
				mockSchema.EXPECT().GetTables().Return(nil, assert.AnError)
				mockContext.EXPECT().Error(mock.Anything).Run(func(message string) {
					color.Errorln(message)
				})
			},
			expected: assert.AnError.Error(),
		}, {
			name: "get views failed",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("")
				mockContext.EXPECT().Option("database").Return("")
				mockSchema.EXPECT().Connection(mock.Anything).Return(mockSchema)
				mockSchema.EXPECT().GetConnection().Return("test")
				mockConfig.EXPECT().GetString(mock.Anything).Return("test")
				mockQuery.EXPECT().Driver().Return(database.DriverMysql)
				mockOrm.EXPECT().Query().Return(mockQuery)
				mockSchema.EXPECT().Orm().Return(mockOrm)
				mockQuery.EXPECT().Raw(mock.Anything).Return(mockQuery)
				mockQuery.EXPECT().Scan(mock.Anything).Return(nil)
				mockSchema.EXPECT().GetTables().Return(nil, nil)
				mockContext.EXPECT().OptionBool("views").Return(true)
				mockSchema.EXPECT().GetViews().Return(nil, assert.AnError)
				mockContext.EXPECT().Error(mock.Anything).Run(func(message string) {
					color.Errorln(message)
				})
			},
			expected: assert.AnError.Error(),
		}, {
			name: "success",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("")
				mockContext.EXPECT().Option("database").Return("")
				mockSchema.EXPECT().Connection(mock.Anything).Return(mockSchema)
				mockSchema.EXPECT().GetConnection().Return("test")
				mockConfig.EXPECT().GetString(mock.Anything).Return("test")
				mockQuery.EXPECT().Driver().Return(database.DriverMysql)
				mockOrm.EXPECT().Query().Return(mockQuery)
				mockSchema.EXPECT().Orm().Return(mockOrm)
				mockQuery.EXPECT().Raw(mock.Anything).Return(mockQuery)
				mockQuery.EXPECT().Scan(mock.Anything).RunAndReturn(func(dest interface{}) error {
					if d, ok := dest.(*queryResult); ok {
						d.Value = "MariaDB"
					}
					return nil
				})
				mockSchema.EXPECT().GetTables().Return([]schema.Table{
					{Name: "test", Size: 100},
				}, nil)
				mockContext.EXPECT().OptionBool("views").Return(true)
				mockSchema.EXPECT().GetViews().Return([]schema.View{
					{Name: "test"},
				}, nil)
				mockQuery.EXPECT().Table(mock.Anything).Return(mockQuery)
				mockQuery.EXPECT().Count(mock.Anything).Return(nil)
				mockContext.EXPECT().NewLine()
				mockContext.EXPECT().TwoColumnDetail(mock.Anything, mock.Anything)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			command := NewShowCommand(mockConfig, mockSchema)
			assert.Contains(t, color.CaptureOutput(func(_ io.Writer) {
				assert.NoError(t, command.Handle(mockContext))
			}), test.expected)
		})
	}

}
