package notification

type Notification interface {
	// Via returns the channel names this notification should be delivered on
	// for the given notifiable recipient.
	// Channel names: "mail", "database", "slack", etc.
	Via(notifiable Notifiable) []string
}

type NotificationWithID interface {
	Notification
	// ID returns a caller-assigned identifier. An empty string is treated
	// the same as not implementing this interface.
	ID() string
}

type Notifiable interface {
	// RouteNotificationFor returns the delivery address for channel: an
	// email address for "mail", the model's string primary key for
	// "database".
	RouteNotificationFor(channel string) string
}

type MailRoutable interface {
	// RouteNotificationForMail returns every address this notifiable
	// should receive the given notification at.
	RouteNotificationForMail(notification Notification) []string
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
	// without a Notifiable model.
	//
	//	facades.Notification().Route("mail", "taylor@example.com").Notify(notification)
	Route(channel, route string) OnDemandNotifiable
}

// OnDemandNotifiable is a Notifiable built on the fly by Manager.Route,
// with no backing model. Chain Route to target more than one channel.
type OnDemandNotifiable interface {
	Notifiable

	// Route adds another channel/address pair to this on-demand target.
	Route(channel, route string) OnDemandNotifiable

	// Notify sends to this target, respecting ShouldQueue.
	Notify(notification Notification) error

	// NotifyNow always delivers synchronously.
	NotifyNow(notification Notification) error
}

// ---- Optional per-channel representation interfaces ----

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
// Build one inside ToMail() using the fluent helpers or by setting fields directly.
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
	// Content holds the plain-text and/or Html bodies.
	Content MailContent
	// Attachments is a list of absolute file paths to attach.
	Attachments []string
	// Headers are arbitrary additional email headers.
	Headers map[string]string
}

// MailContent mirrors contracts/mail.Content so callers don't need to
// import the mail package directly.
type MailContent struct {
	// Html is the HTML body.
	Html string `json:"html"` //nolint:revive,stylecheck
	// Text is the plain-text body.
	Text string `json:"text"`
	// HtmlView is a Goravel view template rendered as the HTML body (alternative to Html).
	HtmlView string `json:"html_view"`
	// TextView is a Goravel view template rendered as the plain-text body (alternative to Text).
	TextView string `json:"text_view"`
	// With is the data passed to HtmlView/TextView.
	With map[string]any `json:"with"`
}
