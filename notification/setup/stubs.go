package main

import "strings"

type Stubs struct{}

func (s Stubs) Config(pkg, facadesImport, facadesPackage string) string {
	content := `package DummyPackage

import (
	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
	config.Add("notification", map[string]any{
		// default is the fallback channel used when Via() is not implemented.
		// In practice Via() is always required, so this is advisory only.
		"default": config.Env("NOTIFICATION_CHANNEL", "mail"),

		"channels": map[string]any{
			// slack contains settings for the Slack incoming-webhook channel.
			"slack": map[string]any{
				// webhook is the default Slack webhook URL, used when a
				// notifiable does not implement RouteNotificationFor("slack").
				"webhook": config.Env("SLACK_WEBHOOK_URL", ""),
				// username overrides the bot name in Slack messages.
				"username": config.Env("SLACK_USERNAME", "Goravel"),
				// icon_emoji overrides the bot icon.
				"icon_emoji": config.Env("SLACK_ICON_EMOJI", ":bell:"),
			},

			// database contains settings for the database channel.
			"database": map[string]any{
				// connection is the DB connection name (empty = default connection).
				"connection": config.Env("NOTIFICATION_DB_CONNECTION", ""),
			},
		},
	})
}
`

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyFacadesImport", facadesImport)
	content = strings.ReplaceAll(content, "DummyFacadesPackage", facadesPackage)

	return content
}

func (s Stubs) NotificationFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/notification"
)

func Notification() notification.Manager {
	return App().MakeNotification()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
