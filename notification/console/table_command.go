package console

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
)

// NotificationTableCommand generates a migration file that creates the
// notifications table used by the database channel. Ported as-is from
// the standalone package — proven working — with one change: the two
// inline fmt.Errorf calls in registerMigration are replaced with named
// errors per AGENTS.md's "no inline errors" rule.
//
// Usage:
//
//	./artisan notification:table
//	./artisan migrate
//
// Open question for maintainers: registerMigration below hand-splices the
// migration entry into bootstrap/migrations.go as a string. Since v1.16,
// make:migration has its own auto-registration mechanism — worth
// confirming whether this command should call into that instead of
// duplicating the string-splicing logic here.
type NotificationTableCommand struct{}

func NewNotificationTableCommand() *NotificationTableCommand {
	return &NotificationTableCommand{}
}

func (c *NotificationTableCommand) Signature() string {
	return "notification:table"
}

func (c *NotificationTableCommand) Description() string {
	return "Create a migration for the notifications table (database channel)"
}

func (c *NotificationTableCommand) Extend() command.Extend {
	return command.Extend{}
}

func (c *NotificationTableCommand) Handle(ctx console.Context) error {
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

	if err := registerMigration(timestamp); err != nil {
		ctx.Warning("Could not auto-register migration: " + err.Error())
		ctx.Warning("Add manually to bootstrap/migrations.go:")
		ctx.Info("  &migrations.M" + timestamp + "CreateNotificationsTable{},")
	} else {
		ctx.Info("Migration registered in bootstrap/migrations.go")
	}

	ctx.Info("Run `./artisan migrate` to apply it.")
	return nil
}

func registerMigration(timestamp string) error {
	bootstrapFile := filepath.Join("bootstrap", "migrations.go")

	content, err := os.ReadFile(bootstrapFile)
	if err != nil {
		return err
	}

	entry := "\t\t&migrations.M" + timestamp + "CreateNotificationsTable{},"
	newLine := "\n" + entry

	src := string(content)
	insertAt := strings.LastIndex(src, "}")
	if insertAt == -1 {
		return errors.NotificationMigrationInsertionPointNotFound
	}

	sliceClose := strings.LastIndex(src[:insertAt], "}")
	if sliceClose == -1 {
		return errors.NotificationMigrationSliceCloseNotFound
	}

	updated := src[:sliceClose] + newLine + "\n\t" + src[sliceClose:]
	return os.WriteFile(bootstrapFile, []byte(updated), 0o644)
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
