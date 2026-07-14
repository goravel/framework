// Package notification defines the public contracts for Goravel's
// notification module: what a notification is, what can receive one, the
// channel driver interface, and the Manager service bound into the
// container as facades.Notification().
package notification

// Notification is implemented by every notification struct.
//
//	type InvoicePaid struct{ Invoice *models.Invoice }
//	func (n *InvoicePaid) Via(_ notification.Notifiable) []string { return []string{"mail", "database"} }
type Notification interface {
	// Via returns the channel names ("mail", "database", "slack") this
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

// Notifiable is implemented by any model that can receive notifications.
//
//	func (u *User) RouteNotificationFor(channel string) string {
//	    switch channel {
//	    case "mail":     return u.Email
//	    case "slack":    return u.SlackWebhookURL
//	    case "database": return fmt.Sprintf("%d", u.ID)
//	    }
//	    return ""
//	}
type Notifiable interface {
	// RouteNotificationFor returns the delivery address for channel: an
	// email address for "mail", a webhook URL for "slack", the model's
	// string primary key for "database".
	RouteNotificationFor(channel string) string
}

// MailRoutable is an optional extension of Notifiable for notifiables
// that should receive mail at more than one address. The mail channel
// prefers this over RouteNotificationFor("mail") when both are
// implemented.
//
//	func (u *User) RouteNotificationForMail(_ notification.Notification) []string {
//	    return []string{u.Email, u.SecondaryEmail}
//	}
type MailRoutable interface {
	// RouteNotificationForMail returns every address this notifiable
	// should receive the given notification at.
	RouteNotificationForMail(notification Notification) []string
}

// Channel is the interface every delivery driver must satisfy. Register
// custom channels via Manager.Extend.
type Channel interface {
	// Name returns the unique identifier for this channel, e.g. "mail", "database", "slack".
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
// All three built-in channels implement this. A channel that only
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

// SlackNotification lets a notification control the outgoing Slack
// incoming-webhook payload.
type SlackNotification interface {
	Notification
	// ToSlack returns the SlackMessage to POST to the webhook URL.
	ToSlack(notifiable Notifiable) SlackMessage
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

// SlackMessage is a full incoming-webhook payload.
// See https://api.slack.com/messaging/webhooks for field semantics.
type SlackMessage struct {
	// Text is the fallback / primary message text.
	Text string
	// Username overrides the bot display name.
	Username string
	// IconEmoji overrides the bot icon, e.g. ":robot_face:".
	IconEmoji string
	// Channel overrides the target channel, e.g. "#alerts".
	Channel string
	// Attachments are legacy Slack attachment blocks.
	Attachments []SlackAttachment
}

// SlackAttachment is a single Slack message attachment.
type SlackAttachment struct {
	// Title is the bold attachment title.
	Title string
	// Text is the attachment body text.
	Text string
	// Color is "good", "warning", "danger", or a hex string like "#36a64f".
	Color string
	// Fields are key-value pairs displayed in a table inside the attachment.
	Fields []SlackField
	// Footer is small text shown at the bottom of the attachment.
	Footer string
	// Timestamp is a Unix timestamp shown in the attachment footer.
	Timestamp int64
}

// SlackField is a single key-value pair inside a SlackAttachment.
type SlackField struct {
	// Title is the field label.
	Title string
	// Value is the field content (supports Slack mrkdwn).
	Value string
	// Short controls whether the field appears side-by-side with other short fields.
	Short bool
}
