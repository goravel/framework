package console

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/database/schema"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
	"github.com/goravel/framework/support/color"
)

func TestTableCommand(t *testing.T) {
	var (
		mockContext *mocksconsole.Context
		mockConfig  *mocksconfig.Config
		mockSchema  *mocksschema.Schema
	)
	beforeEach := func() {
		mockContext = mocksconsole.NewContext(t)
		mockConfig = mocksconfig.NewConfig(t)
		mockSchema = mocksschema.NewSchema(t)
	}
	tests := []struct {
		name     string
		setup    func()
		expected string
	}{
		{
			name: "get tables failed",
			setup: func() {
				mockContext.EXPECT().NewLine()
				mockContext.EXPECT().Option("database").Return("")
				mockSchema.EXPECT().Connection(mock.Anything).Return(mockSchema)
				mockContext.EXPECT().Argument(0).Return("")
				mockSchema.EXPECT().GetTables().Return(nil, assert.AnError)
				mockContext.EXPECT().Error(mock.Anything).Run(func(message string) {
					color.Errorln(message)
				})
			},
			expected: assert.AnError.Error(),
		},
		{
			name: "table not found",
			setup: func() {
				mockContext.EXPECT().NewLine()
				mockContext.EXPECT().Option("database").Return("")
				mockSchema.EXPECT().Connection(mock.Anything).Return(mockSchema)
				mockContext.EXPECT().Argument(0).Return("test")
				mockSchema.EXPECT().GetTables().Return(nil, nil)
				mockContext.EXPECT().Warning(mock.Anything).Run(func(message string) {
					color.Warningln(message)
				})
			},
			expected: "Table 'test' doesn't exist",
		},
		{
			name: "choice table canceled",
			setup: func() {
				mockContext.EXPECT().NewLine()
				mockContext.EXPECT().Option("database").Return("")
				mockSchema.EXPECT().Connection(mock.Anything).Return(mockSchema)
				mockContext.EXPECT().Argument(0).Return("")
				mockSchema.EXPECT().GetTables().Return(nil, nil)
				mockContext.EXPECT().Choice(mock.Anything, mock.Anything).Return("", assert.AnError)
				mockContext.EXPECT().Line(mock.Anything).Run(func(message string) {
					color.Default().Println(message)
				})
			},
			expected: assert.AnError.Error(),
		},
		{
			name: "get columns failed",
			setup: func() {
				mockContext.EXPECT().NewLine()
				mockContext.EXPECT().Option("database").Return("")
				mockSchema.EXPECT().Connection(mock.Anything).Return(mockSchema)
				mockContext.EXPECT().Argument(0).Return("")
				mockSchema.EXPECT().GetTables().Return([]schema.Table{{Name: "test"}}, nil)
				mockContext.EXPECT().Choice(mock.Anything, mock.Anything).Return("test", nil)
				mockSchema.EXPECT().GetColumns("test").Return(nil, assert.AnError)
				mockContext.EXPECT().Error(mock.Anything).Run(func(message string) {
					color.Errorln(message)
				})
			},
			expected: assert.AnError.Error(),
		},
		{
			name: "get indexes failed",
			setup: func() {
				mockContext.EXPECT().NewLine()
				mockContext.EXPECT().Option("database").Return("")
				mockSchema.EXPECT().Connection(mock.Anything).Return(mockSchema)
				mockContext.EXPECT().Argument(0).Return("")
				mockSchema.EXPECT().GetTables().Return([]schema.Table{{Name: "test"}}, nil)
				mockContext.EXPECT().Choice(mock.Anything, mock.Anything).Return("test", nil)
				mockSchema.EXPECT().GetColumns("test").Return(nil, nil)
				mockSchema.EXPECT().GetIndexes("test").Return(nil, assert.AnError)
				mockContext.EXPECT().Error(mock.Anything).Run(func(message string) {
					color.Errorln(message)
				})
			},
			expected: assert.AnError.Error(),
		},
		{
			name: "get foreign keys failed",
			setup: func() {
				mockContext.EXPECT().NewLine()
				mockContext.EXPECT().Option("database").Return("")
				mockSchema.EXPECT().Connection(mock.Anything).Return(mockSchema)
				mockContext.EXPECT().Argument(0).Return("")
				mockSchema.EXPECT().GetTables().Return([]schema.Table{{Name: "test"}}, nil)
				mockContext.EXPECT().Choice(mock.Anything, mock.Anything).Return("test", nil)
				mockSchema.EXPECT().GetColumns("test").Return(nil, nil)
				mockSchema.EXPECT().GetIndexes("test").Return(nil, nil)
				mockSchema.EXPECT().GetForeignKeys("test").Return(nil, assert.AnError)
				mockContext.EXPECT().Error(mock.Anything).Run(func(message string) {
					color.Errorln(message)
				})
			},
			expected: assert.AnError.Error(),
		},
		{
			name: "success",
			setup: func() {
				mockContext.EXPECT().NewLine()
				mockContext.EXPECT().Option("database").Return("")
				mockSchema.EXPECT().Connection(mock.Anything).Return(mockSchema)
				mockContext.EXPECT().Argument(0).Return("")
				mockSchema.EXPECT().GetTables().Return([]schema.Table{{Name: "test"}}, nil)
				mockContext.EXPECT().Choice(mock.Anything, mock.Anything).Return("test", nil)
				mockSchema.EXPECT().GetColumns("test").Return([]schema.Column{
					{Name: "foo", Type: "int", Autoincrement: true, Nullable: true, Default: "bar"},
				}, nil)
				mockSchema.EXPECT().GetIndexes("test").Return([]schema.Index{
					{Name: "index_foo", Columns: []string{"foo", "bar"}, Unique: true, Primary: true},
				}, nil)
				mockSchema.EXPECT().GetForeignKeys("test").Return([]schema.ForeignKey{
					{Name: "fk_foo", Columns: []string{"foo"}, ForeignTable: "bar", ForeignColumns: []string{"baz"}},
				}, nil)
				mockContext.EXPECT().TwoColumnDetail(mock.Anything, mock.Anything)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			command := NewTableCommand(mockConfig, mockSchema)
			assert.Contains(t, color.CaptureOutput(func(_ io.Writer) {
				assert.NoError(t, command.Handle(mockContext))
			}), test.expected)
		})
	}

}
