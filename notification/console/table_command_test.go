package console

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

// newTestApp returns an app mock permissive enough for either
// registration branch registerMigration might take — DatabasePath is
// only actually called on the deprecated kernel.go fallback, but a bare
// t.TempDir() test environment can't control which branch
// env.IsBootstrapSetup() picks, so this stays generic (.Maybe()) rather
// than asserting a specific call count.
func newTestApp(t *testing.T) *mocksfoundation.Application {
	app := mocksfoundation.NewApplication(t)
	app.EXPECT().DatabasePath("kernel.go").Return("database/kernel.go").Maybe()
	return app
}

func TestNotificationsTableCommand_Metadata(t *testing.T) {
	cmd := NewNotificationsTableCommand(newTestApp(t))
	assert.Equal(t, "notifications:table", cmd.Signature())
	assert.Equal(t, "Create a migration for the notifications table (database channel)", cmd.Description())
	assert.Equal(t, command.Extend{}, cmd.Extend())
}

// TestNotificationsTableCommand_Handle_CreatesMigrationFile exercises
// the file-creation half of Handle directly. It does NOT assert on
// which registration branch fires — registerMigration now calls real,
// environment-dependent functions (env.IsBootstrapSetup(),
// debug.ReadBuildInfo(), modify.AddMigration / modify.GoFile(...).Apply())
// that behave differently depending on whether a real Go module/bootstrap
// setup is present, which a bare t.TempDir() isn't. Both the success and
// failure paths end in a graceful ctx.Warning/ctx.Info fallback, so this
// just asserts the file itself is written correctly and Handle doesn't
// error either way.
func TestNotificationsTableCommand_Handle_CreatesMigrationFile(t *testing.T) {
	t.Chdir(t.TempDir())

	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Info(mock.MatchedBy(func(s string) bool {
		return strings.HasPrefix(s, "Migration created successfully:")
	})).Once()
	// registerMigration will fail in this bare temp dir (no real module
	// context) — accept either outcome message rather than asserting one:
	ctx.EXPECT().Warning(mock.Anything).Maybe()
	ctx.EXPECT().Info(mock.Anything).Maybe()

	cmd := NewNotificationsTableCommand(newTestApp(t))
	err := cmd.Handle(ctx)
	assert.NoError(t, err)

	entries, err := os.ReadDir("database/migrations")
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0].Name(), "_create_notifications_table.go")
}

// TestNotificationsTableCommand_Handle_WarnsAndSkips_WhenMigrationAlreadyExists
// relies on two Handle() calls executed back-to-back landing in the same
// wall-clock second (near-certain in practice — sub-millisecond apart —
// since table_command.go computes its timestamp via time.Now() with
// second-level precision and isn't refactored to accept an injected
// clock). Technically has a vanishingly small chance of flaking exactly
// at a second boundary; acceptable trade-off vs. sleeping across a
// second or refactoring Handle's signature just for testability.
func TestNotificationsTableCommand_Handle_WarnsAndSkips_WhenMigrationAlreadyExists(t *testing.T) {
	t.Chdir(t.TempDir())

	firstCtx := mocksconsole.NewContext(t)
	firstCtx.EXPECT().Info(mock.Anything).Maybe()
	firstCtx.EXPECT().Warning(mock.Anything).Maybe()

	cmd := NewNotificationsTableCommand(newTestApp(t))
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

// NOTE: no direct unit tests for registerMigration itself. It now calls
// env.IsBootstrapSetup(), debug.ReadBuildInfo(), and
// modify.AddMigration / modify.GoFile(...).Apply() — real functions with
// filesystem/environment dependencies this package doesn't have seams to
// mock. The same is true of MigrateMakeCommand's own registration logic,
// which this mirrors — worth checking how that command's tests handle
// this (likely integration-style, against a real scaffolded module) and
// matching that approach rather than inventing a different one here.

func TestMigrationStub_ContainsExpectedSchema(t *testing.T) {
	stub := migrationStub("20260101000000")

	assert.Contains(t, stub, "type M20260101000000CreateNotificationsTable struct{}")
	assert.Contains(t, stub, `return "20260101000000_create_notifications_table"`)
	assert.Contains(t, stub, `facades.Schema().HasTable("notifications")`)
	assert.Contains(t, stub, `table.String("notifiable_type")`)
	assert.Contains(t, stub, `table.Index("notifiable_type", "notifiable_id")`)
	assert.Contains(t, stub, `facades.Schema().DropIfExists("notifications")`)
}
