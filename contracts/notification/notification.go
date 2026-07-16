package notification

import (
	"time"

	contractsmail "github.com/goravel/framework/contracts/mail"
)

type Notification interface {
	Via(notifiable Notifiable) []string
}

type NotificationWithID interface {
	Notification
	ID() string
}

type NotificationWithShouldSend interface {
	Notification
	ShouldSend(notifiable Notifiable, channel string) bool
}

type NotificationWithAfterSending interface {
	Notification
	AfterSending(notifiable Notifiable, channel string) error
}

// NOT CURRENTLY WIRED — implementing this on a notification has no
// effect today.
type NotificationWithBackoff interface {
	Notification
	// Backoff returns the number of seconds to wait before retrying
	// after channel is the channel that failed.
	Backoff(channel string) int
}

// NOT CURRENTLY WIRED — same root cause as NotificationWithBackoff:
type NotificationWithRetryUntil interface {
	Notification
	// RetryUntil returns the time after which delivery attempts should
	// stop being retried.
	RetryUntil() time.Time
}

type Notifiable interface {
	// RouteNotificationFor returns the delivery address for channel.
	// The concrete type is channel-specific: built-in channels expect a
	// string (mail: an email address, or []string/map[string]string for
	// multiple addresses; database: the model's primary key, cast to a
	// string via cast.ToString so numeric IDs work too). Custom channels
	// may expect other types entirely — any type not understood by a
	// given channel is treated the same as no route for that channel.
	RouteNotificationFor(channel string) any
}

type MailRoutable interface {
	RouteNotificationForMail(notification Notification) map[string]string
}

// Channel is the interface every delivery driver must satisfy. Register
// custom channels via Manager.Extend.
type Channel interface {
	// Name returns the unique identifier for this channel, e.g. "mail", "database".
	Name() string

	Send(notifiable Notifiable, notification Notification) error
}

type ResolvableChannel interface {
	Channel

	// Resolve computes what Deliver will need. route is whatever
	// RouteNotificationFor returned for this channel; payload is this
	// channel's message type, JSON-marshaled.
	Resolve(notifiable Notifiable, notification Notification) (route string, payload []byte, err error)

	// Deliver sends using only the plain data Resolve produced.
	Deliver(route string, payload []byte) error
}

type Manager interface {
	Send(notifiable Notifiable, notification Notification) error

	SendNow(notifiable Notifiable, notification Notification) error

	Extend(channel Channel)

	Channel(name string) Channel

	Route(channel string, route any) OnDemandNotifiable
}

type OnDemandNotifiable interface {
	Notifiable

	Route(channel string, route any) OnDemandNotifiable

	Notify(notification Notification) error

	NotifyNow(notification Notification) error
}

type MailableNotification interface {
	Notification
	// ToMail returns the MailMessage used to build the outgoing email.
	ToMail(notifiable Notifiable) MailMessage
}

type DatabaseNotification interface {
	Notification
	// ToDatabase returns the map that will be JSON-encoded into the data column.
	ToDatabase(notifiable Notifiable) map[string]any
}

type DatabaseRoutable interface {
	Notification
	// DatabaseConnection returns the connection name to use. Return ""
	// for the default connection.
	DatabaseConnection() string
}

type ShouldQueue interface {
	// OnQueue returns the queue name to use. Return "" for the default queue.
	OnQueue() string
	// OnConnection returns the connection name to use. Return "" for the default.
	OnConnection() string
}

// ---- Value types ----

type MailMessage struct {
	// Subject is the email subject line. Defaults to the notification type name.
	Subject string
	// To overrides the recipient address(es). Leave empty to use
	// RouteNotificationFor("mail") / MailRoutable.
	To []string
	// From overrides the sender address. Leave empty to use the global mail.from config.
	From string
	// ReplyTo sets the Reply-To header.
	ReplyTo string
	// Content holds the plain-text and/or Html bodies — the same type
	// facades.Mail() itself uses, so nothing is lost or re-mapped between
	// a notification's mail representation and a plain Mailable's.
	Content contractsmail.Content
	// Attachments is a list of absolute file paths to attach.
	Attachments []string
	// Headers are arbitrary additional email headers.
	Headers map[string]string
}
