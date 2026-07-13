package channels

// Package channels contains the built-in delivery drivers for Goravel's
// notification module. Ported from codedsultan/goravel-notification;
// each channel is split into Resolve (live values → plain data) and
// Deliver (plain data → actual send), with Send kept as a thin wrapper
// of the two — so Send's external behavior, and every existing test
// against it, is unchanged. See notification/job.go for why the split
// exists.

import (
	"encoding/json"
	"fmt"

	contractsnotification "github.com/goravel/framework/contracts/notification"

	"github.com/goravel/framework/contracts/log"
	contractsmail "github.com/goravel/framework/contracts/mail"
)

// MailChannel delivers notifications via Goravel's mail facade.
type MailChannel struct {
	mail contractsmail.Mail
	log  log.Log
}

func NewMailChannel(mail contractsmail.Mail, logger log.Log) *MailChannel {
	return &MailChannel{mail: mail, log: logger}
}

func (c *MailChannel) Name() string { return "mail" }

func (c *MailChannel) Send(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	route, payload, err := c.Resolve(notifiable, n)
	if err != nil {
		return err
	}
	return c.Deliver(route, payload)
}

// Resolve builds the MailMessage — via ToMail() if the notification
// implements MailableNotification, else a sensible fallback — and
// JSON-encodes it alongside the resolved recipient address. Identical
// logic to the original Send(), just split out and made reusable by the
// queue path.
func (c *MailChannel) Resolve(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) (string, []byte, error) {
	to := notifiable.RouteNotificationFor("mail")
	if to == "" {
		return "", nil, fmt.Errorf("mail channel: %T.RouteNotificationFor(\"mail\") returned empty address", notifiable)
	}

	var msg contractsnotification.MailMessage
	if mn, ok := n.(contractsnotification.MailableNotification); ok {
		msg = mn.ToMail(notifiable)
	} else {
		msg = c.defaultMessage(n)
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return "", nil, fmt.Errorf("mail channel: failed to marshal payload for %T: %w", n, err)
	}

	return to, payload, nil
}

// Deliver unmarshals payload and sends via facades.Mail(), same
// mailable-construction logic the original Send() had inline.
func (c *MailChannel) Deliver(route string, payload []byte) error {
	if route == "" {
		return nil
	}

	var msg contractsnotification.MailMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return fmt.Errorf("mail channel: failed to unmarshal payload: %w", err)
	}

	recipients := []string{route}
	if msg.To != "" {
		recipients = []string{msg.To}
	}

	subject := msg.Subject
	if subject == "" {
		subject = "Notification"
	}

	mailable := &NotificationMailable{
		envelope: &contractsmail.Envelope{
			To:      recipients,
			Subject: subject,
		},
		content: &contractsmail.Content{
			Html:     msg.Content.Html,
			Text:     msg.Content.Text,
			HtmlView: msg.Content.HtmlView,
			TextView: msg.Content.TextView,
			With:     msg.Content.With,
		},
		attachments: msg.Attachments,
		headers:     msg.Headers,
	}

	if msg.From != "" {
		mailable.envelope.From = contractsmail.Address{Address: msg.From}
	}

	if err := c.mail.Send(mailable); err != nil {
		return fmt.Errorf("mail channel: failed to send: %w", err)
	}
	return nil
}

func (c *MailChannel) defaultMessage(n contractsnotification.Notification) contractsnotification.MailMessage {
	return contractsnotification.MailMessage{
		Subject: fmt.Sprintf("Notification: %T", n),
		Content: contractsnotification.MailContent{
			Text: fmt.Sprintf("You have a new %T notification.", n),
		},
	}
}

// NotificationMailable adapts a MailMessage into contractsmail.Mailable.
type NotificationMailable struct {
	envelope    *contractsmail.Envelope
	content     *contractsmail.Content
	attachments []string
	headers     map[string]string
}

func (m *NotificationMailable) Envelope() *contractsmail.Envelope { return m.envelope }
func (m *NotificationMailable) Content() *contractsmail.Content   { return m.content }
func (m *NotificationMailable) Attachments() []string             { return m.attachments }
func (m *NotificationMailable) Headers() map[string]string        { return m.headers }
func (m *NotificationMailable) Queue() *contractsmail.Queue       { return nil }

var _ contractsnotification.ResolvableChannel = (*MailChannel)(nil)
