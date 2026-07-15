package console

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
)

func TestNotificationTableCommand_Metadata(t *testing.T) {
	cmd := NewNotificationTableCommand()
	assert.Equal(t, "notification:table", cmd.Signature())
	assert.Equal(t, "Create a migration for the notifications table (database channel)", cmd.Description())
	assert.Equal(t, command.Extend{}, cmd.Extend())
}

func TestNotificationTableCommand_Handle_CreatesMigration_NoBootstrapFile(t *testing.T) {
	t.Chdir(t.TempDir())

	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Info(mock.MatchedBy(func(s string) bool {
		return strings.HasPrefix(s, "Migration created successfully:")
	})).Once()
	// No bootstrap/migrations.go present — registerMigration fails to
	// read it, so Handle should fall back to the manual-registration
	// warning path, not error out.
	ctx.EXPECT().Warning(mock.MatchedBy(func(s string) bool {
		return strings.HasPrefix(s, "Could not auto-register migration:")
	})).Once()
	ctx.EXPECT().Warning("Add manually to bootstrap/migrations.go:").Once()
	ctx.EXPECT().Info(mock.MatchedBy(func(s string) bool {
		return strings.HasPrefix(s, "  &migrations.M")
	})).Once()
	ctx.EXPECT().Info("Run `./artisan migrate` to apply it.").Once()

	cmd := NewNotificationTableCommand()
	err := cmd.Handle(ctx)
	assert.NoError(t, err)

	entries, err := os.ReadDir("database/migrations")
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0].Name(), "_create_notifications_table.go")
}

func TestNotificationTableCommand_Handle_RegistersInBootstrapFile_WhenPresent(t *testing.T) {
	t.Chdir(t.TempDir())

	assert.NoError(t, os.MkdirAll("bootstrap", 0o755))
	bootstrapContent := `package bootstrap

import "github.com/goravel/framework/contracts/database/schema"

func Migrations() []schema.Migration {
	return []schema.Migration{
	}
}
`
	assert.NoError(t, os.WriteFile("bootstrap/migrations.go", []byte(bootstrapContent), 0o644))

	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Info(mock.MatchedBy(func(s string) bool {
		return strings.HasPrefix(s, "Migration created successfully:")
	})).Once()
	ctx.EXPECT().Info("Migration registered in bootstrap/migrations.go").Once()
	ctx.EXPECT().Info("Run `./artisan migrate` to apply it.").Once()

	cmd := NewNotificationTableCommand()
	err := cmd.Handle(ctx)
	assert.NoError(t, err)

	updated, err := os.ReadFile("bootstrap/migrations.go")
	assert.NoError(t, err)
	assert.Contains(t, string(updated), "&migrations.M")
	assert.Contains(t, string(updated), "CreateNotificationsTable{},")
}

// TestNotificationTableCommand_Handle_WarnsAndSkips_WhenMigrationAlreadyExists
// relies on two Handle() calls executed back-to-back landing in the same
// wall-clock second (near-certain in practice — sub-millisecond apart —
// since table_command.go computes its timestamp via time.Now() with
// second-level precision and isn't refactored to accept an injected
// clock). Technically has a vanishingly small chance of flaking exactly
// at a second boundary; acceptable trade-off vs. sleeping across a
// second or refactoring Handle's signature just for testability.
func TestNotificationTableCommand_Handle_WarnsAndSkips_WhenMigrationAlreadyExists(t *testing.T) {
	t.Chdir(t.TempDir())

	firstCtx := mocksconsole.NewContext(t)
	firstCtx.EXPECT().Info(mock.Anything).Maybe()
	firstCtx.EXPECT().Warning(mock.Anything).Maybe()

	cmd := NewNotificationTableCommand()
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

func TestRegisterMigration_Success(t *testing.T) {
	t.Chdir(t.TempDir())
	assert.NoError(t, os.MkdirAll("bootstrap", 0o755))
	assert.NoError(t, os.WriteFile("bootstrap/migrations.go", []byte(`package bootstrap

func Migrations() []int {
	return []int{
	}
}
`), 0o644))

	err := registerMigration("20260101000000")
	assert.NoError(t, err)

	updated, err := os.ReadFile("bootstrap/migrations.go")
	assert.NoError(t, err)
	assert.Contains(t, string(updated), "&migrations.M20260101000000CreateNotificationsTable{},")
}

func TestRegisterMigration_ReturnsError_WhenBootstrapFileMissing(t *testing.T) {
	t.Chdir(t.TempDir())
	err := registerMigration("20260101000000")
	assert.Error(t, err)
}

func TestRegisterMigration_ReturnsError_WhenNoClosingBrace(t *testing.T) {
	t.Chdir(t.TempDir())
	assert.NoError(t, os.MkdirAll("bootstrap", 0o755))
	assert.NoError(t, os.WriteFile("bootstrap/migrations.go", []byte("package bootstrap"), 0o644))

	err := registerMigration("20260101000000")
	assert.ErrorContains(t, err, "insertion point")
}

func TestRegisterMigration_ReturnsError_WhenOnlyOneClosingBrace(t *testing.T) {
	t.Chdir(t.TempDir())
	assert.NoError(t, os.MkdirAll("bootstrap", 0o755))
	// Exactly one '}' in the whole file: insertAt finds it, but
	// src[:insertAt] then has none, hitting the second LastIndex miss.
	assert.NoError(t, os.WriteFile("bootstrap/migrations.go", []byte("package bootstrap\n}"), 0o644))

	err := registerMigration("20260101000000")
	assert.ErrorContains(t, err, "slice closing brace")
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
