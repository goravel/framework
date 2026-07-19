package console

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/env"
)

// Usage:
//
//	./artisan notifications:table
//	./artisan migrate
type NotificationsTableCommand struct{}

func NewNotificationsTableCommand() *NotificationsTableCommand {
	return &NotificationsTableCommand{}
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

func (c *NotificationsTableCommand) registerMigration(structName string) error {
	if !env.IsBootstrapSetup() {
		return errors.NotificationTableRequiresBootstrapSetup
	}

	modulePath := "goravel"
	if info, ok := debug.ReadBuildInfo(); ok {
		modulePath = info.Main.Path
	}
	pkgImportPath := modulePath + "/database/migrations"
	entry := fmt.Sprintf("&migrations.%s{}", structName)

	return modify.AddMigration(pkgImportPath, entry)
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
