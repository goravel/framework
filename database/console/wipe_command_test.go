package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
)

func TestWipeCommand(t *testing.T) {
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
		name  string
		setup func()
	}{
		{
			name: "Happy path - local",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("database").Return("postgres").Once()
				mockContext.EXPECT().OptionBool("drop-views").Return(true).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllViews().Return(nil).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllTables().Return(nil).Once()
				mockContext.EXPECT().OptionBool("drop-types").Return(true).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllTypes().Return(nil).Once()
			},
		},
		{
			name: "Happy path - local, no drop views and types",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("database").Return("postgres").Once()
				mockContext.EXPECT().OptionBool("drop-views").Return(false).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllTables().Return(nil).Once()
				mockContext.EXPECT().OptionBool("drop-types").Return(false).Once()
			},
		},
		{
			name: "Happy path - force run in production",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("production").Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Option("database").Return("postgres").Once()
				mockContext.EXPECT().OptionBool("drop-views").Return(true).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllViews().Return(nil).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllTables().Return(nil).Once()
				mockContext.EXPECT().OptionBool("drop-types").Return(true).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllTypes().Return(nil).Once()
			},
		},
		{
			name: "Happy path - confirm in production",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("production").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Confirm("Are you sure you want to run this command?").Return(true, nil).Once()
				mockContext.EXPECT().Option("database").Return("postgres").Once()
				mockContext.EXPECT().OptionBool("drop-views").Return(true).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllViews().Return(nil).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllTables().Return(nil).Once()
				mockContext.EXPECT().OptionBool("drop-types").Return(true).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllTypes().Return(nil).Once()
			},
		},
		{
			name: "Sad path - confirm false in production",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("production").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Confirm("Are you sure you want to run this command?").Return(false, assert.AnError).Once()
			},
		},
		{
			name: "Sad path - failed to confirm false in production",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("production").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Confirm("Are you sure you want to run this command?").Return(false, nil).Once()
			},
		},
		{
			name: "Sad path - drop all views failed",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("database").Return("postgres").Once()
				mockContext.EXPECT().OptionBool("drop-views").Return(true).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllViews().Return(assert.AnError).Once()
			},
		},
		{
			name: "Sad path - drop all tables failed",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("database").Return("postgres").Once()
				mockContext.EXPECT().OptionBool("drop-views").Return(false).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllTables().Return(assert.AnError).Once()
			},
		},
		{
			name: "Sad path - drop all types failed",
			setup: func() {
				mockConfig.EXPECT().GetString("app.env").Return("local").Once()
				mockContext.EXPECT().Option("database").Return("postgres").Once()
				mockContext.EXPECT().OptionBool("drop-views").Return(false).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllTables().Return(nil).Once()
				mockContext.EXPECT().OptionBool("drop-types").Return(true).Once()
				mockSchema.EXPECT().Connection("postgres").Return(mockSchema).Once()
				mockSchema.EXPECT().DropAllTypes().Return(assert.AnError).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()

			command := NewWipeCommand(mockConfig, mockSchema)
			assert.NoError(t, command.Handle(mockContext))
		})
	}
}
