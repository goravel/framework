package console

import (
	"database/sql"
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
		{"<fg=green;op=bold>test</>", "version"},
		{"Database", "db"},
		{"Host", "host"},
		{"Port", "1234"},
		{"Username", "username"},
		{"Open Connections", "0"},
		{"Tables", "1"},
		{"Total Size", "0.001 MB"},
		{"<fg=green;op=bold>Tables</>", "<fg=yellow;op=bold>Size (MB)</>"},
		{"test", "0.001"},
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
				// Handle
				mockContext.EXPECT().Argument(0).Return("").Once()
				mockContext.EXPECT().Option("database").Return("test").Once()
				mockSchema.EXPECT().Connection("test").Return(mockSchema).Once()
				mockSchema.EXPECT().Orm().Return(mockOrm).Once()
				mockOrm.EXPECT().Config().Return(database.Config{
					Database: "db",
					Host:     "host",
					Port:     1234,
					Username: "username",
				}).Once()

				// getDataBaseInfo
				mockSchema.EXPECT().Orm().Return(mockOrm).Times(3)
				mockOrm.EXPECT().Name().Return("test").Once()
				mockOrm.EXPECT().Version().Return("version").Once()
				mockOrm.EXPECT().DB().Return(&sql.DB{}, nil).Once()

				// Handle
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
				mockSchema.EXPECT().Orm().Return(mockOrm).Once()
				mockOrm.EXPECT().Config().Return(database.Config{
					Database: "db",
					Host:     "host",
					Port:     1234,
					Username: "username",
				}).Once()

				// getDataBaseInfo
				mockSchema.EXPECT().Orm().Return(mockOrm).Times(3)
				mockOrm.EXPECT().Name().Return("test").Once()
				mockOrm.EXPECT().Version().Return("version").Once()
				mockOrm.EXPECT().DB().Return(&sql.DB{}, nil).Once()

				// Handle
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
				mockSchema.EXPECT().Orm().Return(mockOrm).Once()
				mockOrm.EXPECT().Config().Return(database.Config{
					Database: "db",
					Host:     "host",
					Port:     1234,
					Username: "username",
				}).Once()

				// getDataBaseInfo
				mockSchema.EXPECT().Orm().Return(mockOrm).Times(3)
				mockOrm.EXPECT().Name().Return("test").Once()
				mockOrm.EXPECT().Version().Return("version").Once()
				mockOrm.EXPECT().DB().Return(&sql.DB{}, nil).Once()

				// Handle
				mockSchema.EXPECT().GetTables().Return([]schema.Table{
					{Name: "test", Size: 1024},
				}, nil).Once()
				mockContext.EXPECT().OptionBool("views").Return(true).Once()
				mockSchema.EXPECT().GetViews().Return([]schema.View{
					{Name: "test"},
				}, nil).Once()
				mockSchema.EXPECT().Orm().Return(mockOrm).Once()
				mockOrm.EXPECT().Query().Return(mockQuery).Once()
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
