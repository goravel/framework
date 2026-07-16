package channels

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	contractsmail "github.com/goravel/framework/contracts/mail"
	contractsnotification "github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/notification/mail"
)

// MailChannel delivers notifications via Goravel's mail facade.
type MailChannel struct {
	mailer contractsmail.Mail
}

func NewMailChannel(mailer contractsmail.Mail) *MailChannel {
	return &MailChannel{mailer: mailer}
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
// JSON-encodes it alongside the resolved recipient address(es).
func (c *MailChannel) Resolve(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) (string, []byte, error) {
	addresses, err := c.resolveAddresses(notifiable, n)
	if err != nil {
		return "", nil, err
	}

	var msg contractsnotification.MailMessage
	if mn, ok := n.(contractsnotification.MailableNotification); ok {
		msg = mn.ToMail(notifiable)
	} else {
		msg = c.defaultMessage(n)
	}
	if len(msg.To) == 0 {
		msg.To = addresses
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return "", nil, errors.NotificationMailMarshalPayloadFailed.Args(n, err)
	}

	return strings.Join(msg.To, ","), payload, nil
}

// resolveAddresses prefers MailRoutable (multiple addresses, each with an
// optional display name) when the notifiable implements it, falling back
// to the single RouteNotificationFor("mail") address otherwise. Named
// addresses are formatted "Name <address>" per RFC 5322, matching how
// Laravel's mail routing presents name+address pairs.
func (c *MailChannel) resolveAddresses(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) ([]string, error) {
	if mr, ok := notifiable.(contractsnotification.MailRoutable); ok {
		if routes := mr.RouteNotificationForMail(n); len(routes) > 0 {
			return formatAddresses(routes), nil
		}
	}

	to, _ := notifiable.RouteNotificationFor("mail").(string)
	if to == "" {
		return nil, errors.NotificationMailEmptyRoute.Args(notifiable)
	}
	return []string{to}, nil
}

// formatAddresses turns an address→name map into a deterministically
// ordered list of "Name <address>" (or bare "address" when name is
// empty) strings.
func formatAddresses(routes map[string]string) []string {
	addresses := make([]string, 0, len(routes))
	for address := range routes {
		addresses = append(addresses, address)
	}
	sort.Strings(addresses) // deterministic order — map iteration isn't

	formatted := make([]string, 0, len(addresses))
	for _, address := range addresses {
		if name := routes[address]; name != "" {
			formatted = append(formatted, fmt.Sprintf("%s <%s>", name, address))
		} else {
			formatted = append(formatted, address)
		}
	}
	return formatted
}

// Deliver unmarshals payload and sends via facades.Mail(), same
// mailable-construction logic the original Send() had inline.
func (c *MailChannel) Deliver(route string, payload []byte) error {
	if route == "" {
		return nil
	}

	var msg contractsnotification.MailMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return errors.NotificationMailUnmarshalPayloadFailed.Args(err)
	}

	recipients := msg.To
	if len(recipients) == 0 {
		recipients = strings.Split(route, ",")
	}

	subject := msg.Subject
	if subject == "" {
		subject = "Notification"
	}

	content := msg.Content
	mailable := &NotificationMailable{
		envelope: &contractsmail.Envelope{
			To:      recipients,
			Subject: subject,
		},
		content:     &content,
		attachments: msg.Attachments,
		headers:     msg.Headers,
	}

	if msg.From != "" {
		mailable.envelope.From = contractsmail.Address{Address: msg.From}
	}

	if err := c.mailer.Send(mailable); err != nil {
		return errors.NotificationMailSendFailed.Args(err)
	}
	return nil
}

func (c *MailChannel) defaultMessage(n contractsnotification.Notification) contractsnotification.MailMessage {
	return mail.NewMessage().
		Subject(fmt.Sprintf("Notification: %T", n)).
		Text(fmt.Sprintf("You have a new %T notification.", n)).
		Build()
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
