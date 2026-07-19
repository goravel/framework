package console

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
)

func TestNotificationsTableCommand_Metadata(t *testing.T) {
	cmd := NewNotificationsTableCommand()
	assert.Equal(t, "notifications:table", cmd.Signature())
	assert.Equal(t, "Create a migration for the notifications table (database channel)", cmd.Description())
	assert.Equal(t, command.Extend{}, cmd.Extend())
}

func TestNotificationsTableCommand_Handle_CreatesMigrationFile(t *testing.T) {
	t.Chdir(t.TempDir())

	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Info(mock.MatchedBy(func(s string) bool {
		return strings.HasPrefix(s, "Migration created successfully:")
	})).Once()
	ctx.EXPECT().Warning(mock.MatchedBy(func(s string) bool {
		return strings.HasPrefix(s, "Could not auto-register migration:")
	})).Once()
	ctx.EXPECT().Warning("Add manually to your migrations registration:").Once()
	ctx.EXPECT().Info(mock.MatchedBy(func(s string) bool {
		return strings.HasPrefix(s, "  &migrations.")
	})).Once()
	ctx.EXPECT().Info("Run `./artisan migrate` to apply it.").Once()

	cmd := NewNotificationsTableCommand()
	err := cmd.Handle(ctx)
	assert.NoError(t, err)

	entries, err := os.ReadDir("database/migrations")
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0].Name(), "_create_notifications_table.go")
}

func TestNotificationsTableCommand_Handle_WarnsAndSkips_WhenMigrationAlreadyExists(t *testing.T) {
	t.Chdir(t.TempDir())

	firstCtx := mocksconsole.NewContext(t)
	firstCtx.EXPECT().Info(mock.Anything).Maybe()
	firstCtx.EXPECT().Warning(mock.Anything).Maybe()

	cmd := NewNotificationsTableCommand()
	assert.NoError(t, cmd.Handle(firstCtx))

	entries, err := os.ReadDir("database/migrations")
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	secondCtx := mocksconsole.NewContext(t)
	secondCtx.EXPECT().Warning(mock.MatchedBy(func(s string) bool {
		return strings.HasPrefix(s, "Migration already exists:")
	})).Once()

	assert.NoError(t, cmd.Handle(secondCtx))

	// Still only one file — the second call returned early.
	entries, err = os.ReadDir("database/migrations")
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestNotificationsTableCommand_RegisterMigration_ReturnsError_WhenNotBootstrapSetup(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := NewNotificationsTableCommand()
	err := cmd.registerMigration("M20260101000000CreateNotificationsTable")
	assert.ErrorIs(t, err, errors.NotificationTableRequiresBootstrapSetup)
}

func TestMigrationStub_ContainsExpectedSchema(t *testing.T) {
	stub := migrationStub("20260101000000")

	assert.Contains(t, stub, "type M20260101000000CreateNotificationsTable struct{}")
	assert.Contains(t, stub, `return "20260101000000_create_notifications_table"`)
	assert.Contains(t, stub, `facades.Schema().HasTable("notifications")`)
	assert.Contains(t, stub, `table.String("notifiable_type")`)
	assert.Contains(t, stub, `table.Index("notifiable_type", "notifiable_id")`)
	assert.Contains(t, stub, `facades.Schema().DropIfExists("notifications")`)
}
