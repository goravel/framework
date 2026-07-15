package console

// Stubs holds the raw templates used by make: commands in this package.
type Stubs struct{}

func (r Stubs) Notification() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/notification/mailmessage"
)

type DummyNotification struct {
}

func NewDummyNotification() *DummyNotification {
	return &DummyNotification{}
}

// Via returns the channels this notification should be sent through.
func (r *DummyNotification) Via(notifiable notification.Notifiable) []string {
	return []string{"mail"}
}

// ToMail builds the mail channel's payload. Implement
// contracts/notification.MailableNotification; omit this method entirely
// to fall back to the channel's default plain-text message.
func (r *DummyNotification) ToMail(notifiable notification.Notifiable) notification.MailMessage {
	return mailmessage.NewMailMessage().
		Subject("DummyNotification").
		Html("<p>Hello!</p>").
		Build()
}
`
}

func (r Stubs) NotificationDatabase() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/notification"
)

type DummyNotification struct {
}

func NewDummyNotification() *DummyNotification {
	return &DummyNotification{}
}

// Via returns the channels this notification should be sent through.
func (r *DummyNotification) Via(notifiable notification.Notifiable) []string {
	return []string{"database"}
}

// ToDatabase builds the data persisted to the notifications table.
// Implement contracts/notification.DatabaseNotification; omit this
// method entirely to fall back to a minimal default payload.
func (r *DummyNotification) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{
		"message": "DummyNotification",
	}
}
`
}
