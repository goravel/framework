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
// NOTE: wiring this into actual retry behavior depends on the queue
// driver's retry mechanism, which DispatchJob does not currently
// implement — see the PR discussion. The contract is defined now so
// notifications can declare intent ahead of that wiring landing.
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
// NOTE: same caveat as NotificationWithBackoff — contract defined ahead
// of the queue-driver-level wiring.
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

	// Send delivers the notification to the notifiable target.
	Send(notifiable Notifiable, notification Notification) error

	// SendNow delivers the notification to the notifiable target,
	// bypassing Manager.Send()'s queue routing entirely — even if the
	// notification implements ShouldQueue. Used for channel-scoped
	// direct dispatch: facades.Notification().Channel("mail").SendNow(u, n).
	//
	// For every built-in channel this is identical to Send — none of
	// them have a queued mode of their own, only Manager does. The
	// distinction exists for custom channels that might internally
	// defer work (e.g. batching); implementations should treat SendNow
	// as a hard synchronous guarantee regardless.
	SendNow(notifiable Notifiable, notification Notification) error
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
// Build one inside ToMail() using NewMailMessage()'s fluent builder, or
// by setting fields directly.
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

// NewMailMessage returns a fluent builder for MailMessage. Every method
// returns the same builder so calls can be chained; call Build() (or use
// the builder as a MailMessage directly — it IS one, embedded) to finish.
//
//	func (n *InvoicePaid) ToMail(_ notification.Notifiable) notification.MailMessage {
//	    return notification.NewMailMessage().
//	        Subject("Invoice Paid").
//	        Html("<p>Thanks!</p>").
//	        Build()
//	}
func NewMailMessage() *MailMessageBuilder {
	return &MailMessageBuilder{}
}

// MailMessageBuilder is the fluent builder returned by NewMailMessage.
type MailMessageBuilder struct {
	msg MailMessage
}

func (b *MailMessageBuilder) Subject(subject string) *MailMessageBuilder {
	b.msg.Subject = subject
	return b
}

func (b *MailMessageBuilder) To(addresses ...string) *MailMessageBuilder {
	b.msg.To = addresses
	return b
}

func (b *MailMessageBuilder) From(address string) *MailMessageBuilder {
	b.msg.From = address
	return b
}

func (b *MailMessageBuilder) ReplyTo(address string) *MailMessageBuilder {
	b.msg.ReplyTo = address
	return b
}

func (b *MailMessageBuilder) Html(html string) *MailMessageBuilder {
	b.msg.Content.Html = html
	return b
}

func (b *MailMessageBuilder) Text(text string) *MailMessageBuilder {
	b.msg.Content.Text = text
	return b
}

func (b *MailMessageBuilder) HtmlView(view string, with map[string]any) *MailMessageBuilder {
	b.msg.Content.HtmlView = view
	b.msg.Content.With = with
	return b
}

func (b *MailMessageBuilder) TextView(view string, with map[string]any) *MailMessageBuilder {
	b.msg.Content.TextView = view
	b.msg.Content.With = with
	return b
}

func (b *MailMessageBuilder) Attach(paths ...string) *MailMessageBuilder {
	b.msg.Attachments = append(b.msg.Attachments, paths...)
	return b
}

func (b *MailMessageBuilder) Header(key, value string) *MailMessageBuilder {
	if b.msg.Headers == nil {
		b.msg.Headers = map[string]string{}
	}
	b.msg.Headers[key] = value
	return b
}

// Build returns the finished MailMessage.
func (b *MailMessageBuilder) Build() MailMessage {
	return b.msg
}
