package console

// Stubs holds the raw templates used by make: commands in this package.
// Mirrors mail/console's Stubs{}.Mail() call-site pattern. The template
// body is written against the real, confirmed MailMessage/MailContent
// shape from contracts/notification.go (Content is a nested MailContent
// struct with Html/Text/HtmlView/TextView/With — not a flat Html field,
// and no single View field). Still worth diffing against
// mail/console/stubs.go's actual mechanism (plain string vs go:embed)
// before merging.
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

// ID uniquely identifies this notification instance. Return an empty
// string to let the database channel generate one.
func (r *DummyNotification) ID() string {
	return ""
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
