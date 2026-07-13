// Package notification defines the public contracts for Goravel's
// notification module. Ported from codedsultan/goravel-notification with
// one addition: ResolvableChannel (see below), which fixes a queue-safety
// issue in the original standalone package's dispatch job — see
// notification/job.go for the full explanation.
package notification

// Notification must be implemented by every notification struct in the application.
// A notification describes what should be sent and via which channels.
//
// Example:
//
//	type InvoicePaid struct{ Invoice *models.Invoice }
//
//	func (n *InvoicePaid) Via(_ notification.Notifiable) []string { return []string{"mail", "database"} }
//	func (n *InvoicePaid) ID() string                            { return "" } // auto UUID
type Notification interface {
	// Via returns the channel names this notification should be delivered on
	// for the given notifiable recipient.
	// Channel names: "mail", "database", "slack".
	Via(notifiable Notifiable) []string

	// ID returns a unique identifier used for deduplication and database storage.
	// Return an empty string to have the package generate a UUID automatically.
	ID() string
}

// Notifiable must be implemented by any model that can receive notifications.
// Embed this in your User (or any other) model.
//
// Example:
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
	// RouteNotificationFor returns the delivery address for the given channel.
	// For "mail" this is an email address; for "slack" a webhook URL;
	// for "database" the model's string primary key.
	RouteNotificationFor(channel string) string
}

// Channel is the interface every delivery driver must satisfy.
// Register custom channels via Manager.Extend.
type Channel interface {
	// Name returns the unique identifier for this channel, e.g. "mail", "database", "slack".
	Name() string

	// Send delivers the notification to the notifiable target.
	// Implementations should return a descriptive error on failure.
	Send(notifiable Notifiable, notification Notification) error
}

// ResolvableChannel is an optional extension of Channel. A channel that
// implements it can be dispatched via a queue safely — Resolve is called
// on the calling goroutine (with live Notifiable/Notification values
// still in scope) to produce plain, serializable data, and Deliver is
// called later, potentially on a different goroutine or after a
// queue round-trip, using only that plain data.
//
// This mirrors how contracts/mail's Mailable is handled: Envelope/
// Content/Attachments are resolved eagerly before anything is queued,
// so the queue never needs to carry a live interface value across a
// serialization boundary. All three built-in channels (mail, database,
// slack) implement this. A custom channel that only implements Channel
// still works for synchronous Send(); Manager.Send() falls back to a
// clear error if a ShouldQueue notification targets a channel that
// doesn't implement ResolvableChannel, rather than silently losing data.
type ResolvableChannel interface {
	Channel

	// Resolve computes what Deliver will need, using the live notifiable
	// and notification. route is whatever RouteNotificationFor returned
	// for this channel; payload is this channel's message type
	// (MailMessage/a map/SlackMessage), JSON-marshaled.
	Resolve(notifiable Notifiable, notification Notification) (route string, payload []byte, err error)

	// Deliver sends using only the plain data Resolve produced — no
	// access to the original Notifiable/Notification.
	Deliver(route string, payload []byte) error
}

// Manager is the top-level service bound in the container and exposed via
// facades.Notification(). It dispatches notifications to the appropriate channels.
type Manager interface {
	// Send dispatches the notification to all channels returned by notification.Via().
	// If the notification also implements ShouldQueue, it is dispatched via
	// Goravel's queue; otherwise it is delivered synchronously.
	Send(notifiable Notifiable, notification Notification) error

	// SendNow always delivers synchronously, even if the notification
	// implements ShouldQueue. Useful in tests or time-critical paths.
	SendNow(notifiable Notifiable, notification Notification) error

	// Extend registers a custom channel driver. Call this in your
	// ServiceProvider.Boot() to add community or application-specific channels.
	Extend(channel Channel)

	// Channel returns the registered driver for name, or an error if not found.
	Channel(name string) (Channel, error)
}

// ---- Optional per-channel representation interfaces ----
// A notification may implement any of these to control its per-channel payload.
// If a notification does NOT implement the typed interface for a given channel,
// the channel driver falls back to a sensible default representation.

// MailableNotification is optionally implemented by notifications that want to
// control their mail representation explicitly.
type MailableNotification interface {
	Notification
	// ToMail returns the MailMessage used to build the outgoing email.
	ToMail(notifiable Notifiable) MailMessage
}

// DatabaseNotification is optionally implemented by notifications that want to
// control what data is persisted in the notifications table.
type DatabaseNotification interface {
	Notification
	// ToDatabase returns the map that will be JSON-encoded into the data column.
	ToDatabase(notifiable Notifiable) map[string]any
}

// SlackNotification is optionally implemented by notifications that want full
// control over the Slack incoming-webhook payload.
type SlackNotification interface {
	Notification
	// ToSlack returns the SlackMessage to POST to the webhook URL.
	ToSlack(notifiable Notifiable) SlackMessage
}

// ShouldQueue is an optional marker interface. Notifications that implement it
// are dispatched via Goravel's queue system instead of being sent inline.
// The Manager respects this interface in Send() but ignores it in SendNow().
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
	// To overrides the recipient address. Leave empty to use RouteNotificationFor("mail").
	To string
	// From overrides the sender address. Leave empty to use the global mail.from config.
	From string
	// ReplyTo sets the Reply-To header.
	ReplyTo string
	// Content holds the plain-text and/or Html bodies.
	// Mirrors the goravel/framework contracts/mail.Content struct.
	Content MailContent
	// Attachments is a list of absolute file paths to attach.
	Attachments []string
	// Headers are arbitrary additional email headers.
	Headers map[string]string
}

// MailContent mirrors contracts/mail.Content exactly so callers do not need to
// import the framework mail package directly.
//
// CONFIRMED against real contracts/mail.Content this pass: it does NOT
// have a single View field — it splits into HtmlView and TextView. The
// original standalone package's MailContent (a single View string) was
// built against an incorrect assumption about that shape; fixed here.
type MailContent struct {
	// Html is the HTML body.
	Html string `json:"html"` //nolint:revive,stylecheck
	// Text is the plain-text body.
	Text string `json:"text"`
	// HtmlView is a Goravel view template name rendered as the HTML body
	// (alternative to Html).
	HtmlView string `json:"html_view"`
	// TextView is a Goravel view template name rendered as the
	// plain-text body (alternative to Text).
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
