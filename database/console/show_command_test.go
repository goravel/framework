package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
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
	successCaseExpected := [][2]string{
		{"<fg=green;op=bold>MariaDB</>", "test-version"},
		{"Database", "db"},
		{"Host", "host"},
		{"Port", "port"},
		{"Username", "username"},
		{"Open Connections", "2"},
		{"Tables", "1"},
		{"Total Size", "0.000 MB"},
		{"<fg=green;op=bold>Tables</>", "<fg=yellow;op=bold>Size (MB)</>"},
		{"test", "0.000"},
		{"<fg=green;op=bold>Views</>", "<fg=yellow;op=bold>Rows</>"},
		{"test", "0"},
	}
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "invalid argument",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("test").Once()
				mockContext.EXPECT().Error("No arguments expected for 'db:show' command, got 'test'.").Once()
			},
		},
		{
			name: "get tables failed",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("").Once()
				mockContext.EXPECT().Option("database").Return("test").Once()
				mockSchema.EXPECT().Connection("test").Return(mockSchema).Once()
				mockSchema.EXPECT().GetConnection().Return("test").Once()
				mockConfig.EXPECT().GetString("database.connections.test.database").Return("db").Once()
				mockConfig.EXPECT().GetString("database.connections.test.host").Return("host").Once()
				mockConfig.EXPECT().GetString("database.connections.test.port").Return("port").Once()
				mockConfig.EXPECT().GetString("database.connections.test.username").Return("username").Once()
				mockQuery.EXPECT().Driver().Return(database.DriverMysql).Twice()
				mockOrm.EXPECT().Query().Return(mockQuery).Times(4)
				mockSchema.EXPECT().Orm().Return(mockOrm).Times(4)
				mockQuery.EXPECT().Raw("SELECT VERSION() AS value;").Return(mockQuery).Once()
				mockQuery.EXPECT().Raw("SHOW status WHERE variable_name = 'threads_connected';").Return(mockQuery).Once()
				mockQuery.EXPECT().Scan(&queryResult{}).Return(nil).Twice()
				mockSchema.EXPECT().GetTables().Return(nil, assert.AnError).Once()
				mockContext.EXPECT().Error(assert.AnError.Error()).Once()
			},
		},
		{
			name: "get views failed",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("").Once()
				mockContext.EXPECT().Option("database").Return("test").Once()
				mockSchema.EXPECT().Connection("test").Return(mockSchema).Once()
				mockSchema.EXPECT().GetConnection().Return("test").Once()
				mockConfig.EXPECT().GetString("database.connections.test.database").Return("db").Once()
				mockConfig.EXPECT().GetString("database.connections.test.host").Return("host").Once()
				mockConfig.EXPECT().GetString("database.connections.test.port").Return("port").Once()
				mockConfig.EXPECT().GetString("database.connections.test.username").Return("username").Once()
				mockQuery.EXPECT().Driver().Return(database.DriverMysql).Twice()
				mockOrm.EXPECT().Query().Return(mockQuery).Times(4)
				mockSchema.EXPECT().Orm().Return(mockOrm).Times(4)
				mockQuery.EXPECT().Raw("SELECT VERSION() AS value;").Return(mockQuery).Once()
				mockQuery.EXPECT().Raw("SHOW status WHERE variable_name = 'threads_connected';").Return(mockQuery).Once()
				mockQuery.EXPECT().Scan(&queryResult{}).Return(nil).Twice()
				mockSchema.EXPECT().GetTables().Return(nil, nil).Once()
				mockContext.EXPECT().OptionBool("views").Return(true).Once()
				mockSchema.EXPECT().GetViews().Return(nil, assert.AnError).Once()
				mockContext.EXPECT().Error(assert.AnError.Error()).Once()
			},
		},
		{
			name: "success",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("").Once()
				mockContext.EXPECT().Option("database").Return("test").Once()
				mockSchema.EXPECT().Connection("test").Return(mockSchema).Once()
				mockSchema.EXPECT().GetConnection().Return("test").Once()
				mockConfig.EXPECT().GetString("database.connections.test.database").Return("db").Once()
				mockConfig.EXPECT().GetString("database.connections.test.host").Return("host").Once()
				mockConfig.EXPECT().GetString("database.connections.test.port").Return("port").Once()
				mockConfig.EXPECT().GetString("database.connections.test.username").Return("username").Once()
				mockQuery.EXPECT().Driver().Return(database.DriverMysql).Twice()
				mockOrm.EXPECT().Query().Return(mockQuery).Times(5)
				mockSchema.EXPECT().Orm().Return(mockOrm).Times(5)
				mockQuery.EXPECT().Raw("SELECT VERSION() AS value;").Return(mockQuery).Once()
				mockQuery.EXPECT().Raw("SHOW status WHERE variable_name = 'threads_connected';").Return(mockQuery).Once()
				mockQuery.EXPECT().Scan(&queryResult{}).Run(func(dest interface{}) {
					if d, ok := dest.(*queryResult); ok {
						d.Value = "test-version-MariaDB"
					}
				}).Return(nil).Once()
				mockQuery.EXPECT().Scan(&queryResult{}).Run(func(dest interface{}) {
					if d, ok := dest.(*queryResult); ok {
						d.Value = "2"
					}
				}).Return(nil).Once()
				mockSchema.EXPECT().GetTables().Return([]schema.Table{
					{Name: "test", Size: 100},
				}, nil).Once()
				mockContext.EXPECT().OptionBool("views").Return(true).Once()
				mockSchema.EXPECT().GetViews().Return([]schema.View{
					{Name: "test"},
				}, nil).Once()
				mockQuery.EXPECT().Table("test").Return(mockQuery).Once()
				var rows int64
				mockQuery.EXPECT().Count(&rows).Return(nil).Once()
				mockContext.EXPECT().NewLine().Times(4)
				for i := range successCaseExpected {
					mockContext.EXPECT().TwoColumnDetail(successCaseExpected[i][0], successCaseExpected[i][1]).Once()
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			command := NewShowCommand(mockConfig, mockSchema)
			assert.NoError(t, command.Handle(mockContext))
		})
	}
}
