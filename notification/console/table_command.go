package console

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/env"
)

// NotificationsTableCommand generates a migration file that creates the
// notifications table used by the database channel.
//
// Unlike make:migration, this writes a pre-filled schema directly rather
// than delegating to migration.Migrator — mirrors Laravel's
// notifications:table, which does the same (its own stub-copy, not the
// generic migration creator), since a pre-filled notifications-table
// schema isn't something the generic migration creator supports without
// its own custom stub mechanism this command doesn't have visibility
// into. Registration, however, now uses the exact same real mechanism
// make:migration itself uses (confirmed against
// database/console/migration/migrate_make_command.go) — see
// registerMigration below.
//
// Usage:
//
//	./artisan notifications:table
//	./artisan migrate
type NotificationsTableCommand struct {
	app foundation.Application
}

func NewNotificationsTableCommand(app foundation.Application) *NotificationsTableCommand {
	return &NotificationsTableCommand{app: app}
}

func (c *NotificationsTableCommand) Signature() string {
	return "notifications:table"
}

func (c *NotificationsTableCommand) Description() string {
	return "Create a migration for the notifications table (database channel)"
}

func (c *NotificationsTableCommand) Extend() command.Extend {
	return command.Extend{}
}

func (c *NotificationsTableCommand) Handle(ctx console.Context) error {
	timestamp := time.Now().Format("20060102150405")
	filename := timestamp + "_create_notifications_table.go"
	dest := filepath.Join("database", "migrations", filename)

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	if _, err := os.Stat(dest); err == nil {
		ctx.Warning("Migration already exists: " + dest)
		return nil
	}

	if err := os.WriteFile(dest, []byte(migrationStub(timestamp)), 0o644); err != nil {
		return err
	}

	ctx.Info("Migration created successfully: " + dest)

	structName := "M" + timestamp + "CreateNotificationsTable"
	if err := c.registerMigration(structName); err != nil {
		ctx.Warning("Could not auto-register migration: " + err.Error())
		ctx.Warning("Add manually to your migrations registration:")
		ctx.Info("  &migrations." + structName + "{},")
	} else {
		ctx.Info("Migration registered successfully")
	}

	ctx.Info("Run `./artisan migrate` to apply it.")
	return nil
}

// registerMigration mirrors migration.MigrateMakeCommand.Handle's own
// registration branch exactly: modify.AddMigration for apps on the
// v1.16+ bootstrap setup, falling back to the deprecated kernel.go
// codemod otherwise — replaces the previous hand-rolled
// bootstrap/migrations.go string-splicing entirely.
//
// pkgImportPath is built the same way support/console.Make.GetPackageImportPath
// does it (confirmed from that file directly) — debug.ReadBuildInfo().Main.Path
// plus the migrations folder, with the same "goravel" fallback
// Make.GetModuleName() uses when build info isn't available (e.g. under
// `go test`, where ReadBuildInfo can return ok=false). This command
// doesn't call supportconsole.NewMake itself, since it hand-writes the
// migration file directly for the pre-filled-schema reason explained in
// the type doc comment above — but the import-path logic is copied
// exactly rather than reinvented.
//
// The kernel.go fallback path uses c.app.DatabasePath("kernel.go") —
// not a hardcoded "database/kernel.go" string — matching
// MigrateMakeCommand.registerInKernel exactly, which is why this is now
// a method (needs c.app) rather than the standalone function it was
// before this fix.
func (c *NotificationsTableCommand) registerMigration(structName string) error {
	modulePath := "goravel"
	if info, ok := debug.ReadBuildInfo(); ok {
		modulePath = info.Main.Path
	}
	pkgImportPath := modulePath + "/database/migrations"
	entry := fmt.Sprintf("&migrations.%s{}", structName)

	if env.IsBootstrapSetup() {
		return modify.AddMigration(pkgImportPath, entry)
	}

	// DEPRECATED path, mirrors migrate_make_command.go's
	// registerInKernel exactly, targeting database/kernel.go — kept for
	// apps that haven't migrated to the bootstrap setup yet.
	return modify.GoFile(c.app.DatabasePath("kernel.go")).
		Find(match.Imports()).Modify(modify.AddImport(pkgImportPath)).
		Find(match.Migrations()).Modify(modify.Register(entry)).
		Apply()
}

// Column shape here MUST stay in sync with DatabaseNotificationModel in
// notification/channels/database.go.
func migrationStub(timestamp string) string {
	return `package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M` + timestamp + `CreateNotificationsTable struct{}

func (r *M` + timestamp + `CreateNotificationsTable) Signature() string {
	return "` + timestamp + `_create_notifications_table"
}

func (r *M` + timestamp + `CreateNotificationsTable) Up() error {
	if facades.Schema().HasTable("notifications") {
		return nil
	}

	return facades.Schema().Create("notifications", func(table schema.Blueprint) {
		table.String("id", 36)
		table.Primary("id")
		table.String("type")
		table.String("notifiable_type")
		table.String("notifiable_id")
		table.Text("data")
		table.Timestamp("read_at").Nullable()
		table.Timestamps()
		table.Index("notifiable_type", "notifiable_id")
	})
}

func (r *M` + timestamp + `CreateNotificationsTable) Down() error {
	return facades.Schema().DropIfExists("notifications")
}
`
}
