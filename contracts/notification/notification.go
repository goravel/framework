// Package notification defines the public contracts for Goravel's
// notification module: what a notification is, what can receive one, the
// channel driver interface, and the Manager service bound into the
// container as facades.Notification().
package notification

import (
	"time"

	contractsmail "github.com/goravel/framework/contracts/mail"
)

// Notification is implemented by every notification struct.
//
//	type InvoicePaid struct{ Invoice *models.Invoice }
//	func (n *InvoicePaid) Via(_ notification.Notifiable) []string { return []string{"mail", "database"} }
type Notification interface {
	// Via returns the channel names ("mail", "database") this
	// notification should be delivered on for the given notifiable.
	Via(notifiable Notifiable) []string
}

// NotificationWithID is an optional extension of Notification for
// notifications that need a stable, caller-assigned ID — e.g. to
// deduplicate retries or correlate a persisted database record with an
// external one. When not implemented, the database channel generates a
// UUID.
type NotificationWithID interface {
	Notification
	// ID returns a caller-assigned identifier. An empty string is treated
	// the same as not implementing this interface.
	ID() string
}

// NotificationWithShouldSend is an optional extension letting a
// notification veto delivery to a specific notifiable/channel pair at
// send time — e.g. to respect a user's per-channel notification
// preferences without duplicating that check in every Via() method.
type NotificationWithShouldSend interface {
	Notification
	// ShouldSend returns false to skip delivery on this channel for this
	// notifiable. Checked after Via() has already included the channel.
	ShouldSend(notifiable Notifiable, channel string) bool
}

// NotificationWithAfterSending is an optional extension for running a
// side effect after a notification has been successfully delivered on a
// channel — e.g. incrementing a sent-count, or triggering an event.
type NotificationWithAfterSending interface {
	Notification
	// AfterSending is called once per channel after a successful Send/Deliver.
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

// Notifiable is implemented by any model that can receive notifications.
//
//	func (u *User) RouteNotificationFor(channel string) string {
//	    switch channel {
//	    case "mail":     return u.Email
//	    case "database": return fmt.Sprintf("%d", u.ID)
//	    }
//	    return ""
//	}
type Notifiable interface {
	// RouteNotificationFor returns the delivery address for channel: an
	// email address for "mail", the model's string primary key for
	// "database".
	RouteNotificationFor(channel string) string
}

// MailRoutable is an optional extension of Notifiable for notifiables
// that should receive mail at more than one address, optionally with a
// display name per address. The mail channel prefers this over
// RouteNotificationFor("mail") when both are implemented.
//
//	func (u *User) RouteNotificationForMail(_ notification.Notification) map[string]string {
//	    return map[string]string{u.Email: u.Name, u.SecondaryEmail: ""}
//	}
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

// ResolvableChannel is an optional extension of Channel that makes a
// channel safe to dispatch via the queue. Resolve runs synchronously,
// while the live Notifiable/Notification are still in scope, and
// produces plain, serializable data; Deliver later sends using only that
// data, possibly on a different goroutine or after a queue round-trip.
// Both built-in channels implement this. A channel that only
// implements Channel still works for synchronous Send(), but
// Manager.Send() returns a clear error if a ShouldQueue notification
// targets it.
type ResolvableChannel interface {
	Channel

	// Resolve computes what Deliver will need. route is whatever
	// RouteNotificationFor returned for this channel; payload is this
	// channel's message type, JSON-marshaled.
	Resolve(notifiable Notifiable, notification Notification) (route string, payload []byte, err error)

	// Deliver sends using only the plain data Resolve produced.
	Deliver(route string, payload []byte) error
}

// Manager is the top-level service bound in the container and exposed via
// facades.Notification(). It dispatches notifications to the appropriate channels.
type Manager interface {
	// Send dispatches to all channels returned by notification.Via(). If
	// the notification also implements ShouldQueue, it's dispatched via
	// Goravel's queue; otherwise delivered synchronously.
	Send(notifiable Notifiable, notification Notification) error

	// SendNow always delivers synchronously, even if the notification
	// implements ShouldQueue.
	SendNow(notifiable Notifiable, notification Notification) error

	// Extend registers a custom channel driver, typically from
	// ServiceProvider.Boot().
	Extend(channel Channel)

	// Channel returns the registered driver for name, or nil (logging the
	// lookup error) if no driver is registered under that name.
	Channel(name string) Channel

	// Route begins an on-demand notification: send to a raw address
	// without a Notifiable model. route is typically a string (an email
	// address, a phone number) but is typed any so channels that need a
	// richer route (e.g. a struct pairing an ID with metadata) aren't
	// forced through a string encoding — built-in channels only ever use
	// string routes today.
	//
	//	facades.Notification().Route("mail", "taylor@example.com").Notify(notification)
	Route(channel string, route any) OnDemandNotifiable
}

// OnDemandNotifiable is a Notifiable built on the fly by Manager.Route,
// with no backing model. Chain Route to target more than one channel.
type OnDemandNotifiable interface {
	Notifiable

	// Route adds another channel/route pair to this on-demand target.
	Route(channel string, route any) OnDemandNotifiable

	// Notify sends to this target, respecting ShouldQueue.
	Notify(notification Notification) error

	// NotifyNow always delivers synchronously.
	NotifyNow(notification Notification) error
}

// ---- Optional per-channel representation interfaces ----
// A notification may implement any of these to control its per-channel
// payload. If not implemented, the channel driver falls back to a
// sensible default.

// MailableNotification lets a notification control its mail representation.
type MailableNotification interface {
	Notification
	// ToMail returns the MailMessage used to build the outgoing email.
	ToMail(notifiable Notifiable) MailMessage
}

// DatabaseNotification lets a notification control what's persisted in
// the notifications table.
type DatabaseNotification interface {
	Notification
	// ToDatabase returns the map that will be JSON-encoded into the data column.
	ToDatabase(notifiable Notifiable) map[string]any
}

// DatabaseRoutable is an optional extension letting a notification pick
// which database connection its DatabaseNotification is persisted on —
// mirrors ShouldQueue's OnConnection() pattern rather than a global
// config default, since the right connection is usually a property of
// the notification (or its data volume/retention needs), not a
// blanket app-wide setting.
type DatabaseRoutable interface {
	Notification
	// DatabaseConnection returns the connection name to use. Return ""
	// for the default connection.
	DatabaseConnection() string
}

// ShouldQueue is an optional marker interface. Notifications that
// implement it are dispatched via Goravel's queue instead of inline.
// Manager respects this in Send() but ignores it in SendNow().
type ShouldQueue interface {
	// OnQueue returns the queue name to use. Return "" for the default queue.
	OnQueue() string
	// OnConnection returns the connection name to use. Return "" for the default.
	OnConnection() string
}

// ---- Value types ----

// MailMessage describes the email that should be sent for a notification.
// Build one inside ToMail() using notification/mailmessage's fluent
// builder (see mailmessage.NewMailMessage), or by setting fields
// directly. The builder lives in its own leaf package rather than here
// because it's concrete implementation, not a contract — contracts/
// holds public interfaces and plain value types only (see contracts/mail
// for the same shape), and a builder that both notification/channels and
// the root notification package need to reach can't live in the root
// notification package itself without an import cycle (root notification
// already imports notification/channels).
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
