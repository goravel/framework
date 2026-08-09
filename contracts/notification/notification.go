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

// NotificationWithTries is an optional extension for queued
// notifications that want to cap retry attempts on delivery failure.
// Mirrors contracts/broadcasting.ShouldBroadcastWithTries exactly, for
// consistency across the two queued-retry mechanisms in this codebase.
type NotificationWithTries interface {
	Notification
	// Tries returns the maximum number of attempts for the given
	// channel. 0 / not implementing this interface means the
	// notification is single-shot on that channel — no retry at all,
	// not even once, matching ShouldBroadcastWithTries's semantics.
	Tries(channel string) int
}

// NotificationWithBackoff is an optional extension for queued
// notifications that want to control the delay before each retry
// attempt. Mirrors contracts/broadcasting.ShouldBroadcastWithBackoff
// exactly: Backoff(channel) is called once, while the notification is
// still live, and returns the FULL per-attempt schedule up front — not
// re-invoked per retry. ShouldRetry indexes into the captured slice by
// attempt number, and the last value repeats for any attempt beyond the
// slice's length (min(attempt-1, len(backoff)-1)), matching Laravel's
// Worker::calculateBackoff and BroadcastJob.ShouldRetry precisely.
//
// Only takes effect together with NotificationWithTries; without Tries
// the notification is single-shot regardless of Backoff.
type NotificationWithBackoff interface {
	Notification
	// Backoff returns the delay before each retry attempt on channel,
	// in order; the last value repeats for subsequent attempts.
	Backoff(channel string) []time.Duration
}

type Notifiable interface {
	// RouteNotificationFor returns the delivery address for specified channel.
	// Concrete address types vary by channel implementation:
	// mail:
	//   string        single recipient address
	//   []string      multiple recipient addresses
	//   map[string]string address-name recipient mapping
	// database:
	//   string        model primary key (numeric IDs auto converted via cast.ToString)
	// Custom channels support arbitrary custom types. Unrecognized types for a channel are treated as no valid route.
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
