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

// NotificationWithBackoff is an optional extension for queued
// notifications that want to control the delay before a retry after a
// failed delivery attempt.
//
// NOT CURRENTLY WIRED — implementing this on a notification has no
// effect today. DispatchJob is registered once as a single shared
// instance (see service_provider.go) and the queue worker's retry hook,
// contracts/queue.JobWithShouldRetry.ShouldRetry(err, attempt), is
// called on that shared instance without the failing task's args. There
// is currently no race-free way for DispatchJob to know which
// notification's Backoff() applies to a given retry decision, since
// concurrent deliveries for different notifications share the same
// DispatchJob instance. Wiring this safely needs one of:
//   - a contracts/queue change so ShouldRetry also receives the task's
//     args (the cleanest fix, but changes the queue package's public
//     contract — belongs in its own proposal, not silently in this PR), or
//   - DispatchJob running its own internal retry loop using values
//     encoded into the dispatch item at resolve time, bypassing the
//     queue framework's own retry/failed-job tracking entirely.
//
// The contract is defined now so notification structs that declare
// Backoff() today don't need a breaking change once one of the above
// lands — but until then, this method is read by nothing.
type NotificationWithBackoff interface {
	Notification
	// Backoff returns the number of seconds to wait before retrying
	// after channel is the channel that failed.
	Backoff(channel string) int
}

// NotificationWithRetryUntil is an optional extension for queued
// notifications that want to stop retrying after a deadline rather than
// a fixed attempt count.
//
// NOT CURRENTLY WIRED — same root cause as NotificationWithBackoff:
// DispatchJob has no race-free way to associate this value with a
// specific in-flight retry given the shared-instance, args-less
// ShouldRetry hook. See NotificationWithBackoff's doc comment for the
// two real paths to wiring this in.
type NotificationWithRetryUntil interface {
	Notification
	// RetryUntil returns the time after which delivery attempts should
	// stop being retried.
	RetryUntil() time.Time
}

type Notifiable interface {
	// RouteNotificationFor returns the delivery address for channel: an
	// email address for "mail", the model's string primary key for
	// "database".
	RouteNotificationFor(channel string) string
}

type MailRoutable interface {
	// RouteNotificationForMail returns every address this notifiable
	// should receive the given notification at, keyed by address with
	// an optional display name as the value (empty string for no name).
	RouteNotificationForMail(notification Notification) map[string]string
}

// Channel is the interface every delivery driver must satisfy. Register
// custom channels via Manager.Extend.
type Channel interface {
	// Name returns the unique identifier for this channel, e.g. "mail", "database".
	Name() string

	// Send delivers the notification to the notifiable target. Channel
	// drivers only need this one method — the facade (Manager) is what
	// decides whether a send goes through the queue. Manager.SendNow and
	// OnDemandNotifiable.NotifyNow are the actual bypass points for
	// callers who want to skip queue routing; they call this same Send.
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
