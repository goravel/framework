package console

// Stubs holds the raw templates used by make: commands in this package.
type Stubs struct{}

func (r Stubs) Notification() string {
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
	return []string{"mail"}
}

// ToMail builds the mail channel's payload. Implement
// contracts/notification.MailableNotification; omit this method entirely
// to fall back to the channel's default plain-text message.
func (r *DummyNotification) ToMail(notifiable notification.Notifiable) notification.MailMessage {
	return notification.MailMessage{
		Subject: "DummyNotification",
		Content: notification.MailContent{
			Html: "<p>Hello!</p>",
		},
	}
}
`
}
