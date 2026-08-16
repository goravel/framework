package notification

import (
	"time"

	contractsmail "github.com/goravel/framework/contracts/mail"
)

// Channel name constants. Use these instead of raw string literals in
// Notification.Via, Notifiable.RouteNotificationFor, and Manager.Route so
// a typo is caught at compile time rather than silently dropping a route.
const (
	// ChannelMail is the name of the built-in mail delivery channel.
	ChannelMail = "mail"

	// ChannelDatabase is the name of the built-in database delivery channel.
	ChannelDatabase = "database"
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
	// RouteNotificationFor returns the delivery address for specified channel.
	// Concrete address types vary by channel implementation:
	// mail:
	//   string        single recipient address
	//   []string      multiple recipient addresses
	//   map[string]string address-name recipient mapping
	// database:
	//   string        model primary key (numeric IDs auto converted via cast.ToString)
	// Custom channels support arbitrary custom types. Unrecognized types for a channel are treated as no valid route.
	//
	// Prefer the typed alternatives for the built-in channels instead of
	// matching on the channel name constants (ChannelMail / ChannelDatabase)
	// here: implement MailRoutable (mail) or DatabaseRoutable (database)
	// and the channel uses them directly, so a typo can't silently drop a
	// route.
	RouteNotificationFor(channel string) any
}

// MailRoutable is implemented by a Notifiable to provide multiple mail
// recipients (address→name mapping). It is preferred over
// RouteNotificationFor, but an empty result is not an error by itself: the
// mail channel falls back to RouteNotificationFor(ChannelMail). An empty
// result from both is an error.
type MailRoutable interface {
	RouteNotificationForMail(notification Notification) map[string]string
}

// DatabaseRoutable is implemented by a Notifiable to provide the database
// channel's delivery route (the primary key persisted as NotifiableID) in
// a type-safe way, without matching the channel name in RouteNotificationFor.
//
// RouteNotificationForDatabase is preferred over RouteNotificationFor, but
// returning "" is not an error by itself: like MailRoutable's empty-result
// fallback, the channel then falls back to
// RouteNotificationFor(ChannelDatabase). An empty result from both is an
// error.
type DatabaseRoutable interface {
	RouteNotificationForDatabase() string
}

// Channel is the interface every delivery driver must satisfy. Register
// custom channels via Manager.Extend.
type Channel interface {
	// Name returns the unique identifier for this channel, e.g. ChannelMail, ChannelDatabase.
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

// NotificationWithDatabaseConnection is implemented by a Notification to
// select a non-default database connection for the database channel.
//
// Breaking change: this interface was previously named DatabaseRoutable,
// which is now the notifiable-side route interface. Any code implementing
// the old name must be updated to NotificationWithDatabaseConnection.
type NotificationWithDatabaseConnection interface {
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
	// RouteNotificationFor(ChannelMail) / MailRoutable.
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
