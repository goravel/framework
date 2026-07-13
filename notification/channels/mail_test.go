package channels_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsnotification "github.com/goravel/framework/contracts/notification"
	mocklog "github.com/goravel/framework/mocks/log"
	mockmail "github.com/goravel/framework/mocks/mail"
	"github.com/goravel/framework/notification/channels"

	contractsmail "github.com/goravel/framework/contracts/mail"
)

// ---- Fakes ----

type mailNotifiable struct{ addr string }

func (m *mailNotifiable) RouteNotificationFor(channel string) string {
	if channel == "mail" {
		return m.addr
	}
	return ""
}

// plainNotification does NOT implement MailableNotification — tests the fallback path.
type plainNotification struct{}

func (p *plainNotification) Via(_ contractsnotification.Notifiable) []string { return []string{"mail"} }
func (p *plainNotification) ID() string                                      { return "" }

// richNotification implements MailableNotification — tests the ToMail() path.
type richNotification struct {
	msg contractsnotification.MailMessage
}

func (r *richNotification) Via(_ contractsnotification.Notifiable) []string { return []string{"mail"} }
func (r *richNotification) ID() string                                      { return "" }
func (r *richNotification) ToMail(_ contractsnotification.Notifiable) contractsnotification.MailMessage {
	return r.msg
}

// ---- Tests ----

func TestMailChannel_Name(t *testing.T) {
	ch := channels.NewMailChannel(nil, nil)
	assert.Equal(t, "mail", ch.Name())
}

func TestMailChannel_Send_UsesDefaultMessage_WhenNotMailableNotification(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	// The channel should call mailer.Send with a Mailable — capture it.
	mailer.On("Send", mock.AnythingOfType("*channels.NotificationMailable")).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	mailer.AssertExpectations(t)
}

func TestMailChannel_Send_UsesToMail_WhenMailableNotification(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	mailer.On("Send", mock.AnythingOfType("*channels.NotificationMailable")).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &richNotification{
		msg: contractsnotification.MailMessage{
			Subject: "Invoice Paid",
			Content: contractsnotification.MailContent{Text: "Your invoice was paid."},
		},
	}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	mailer.AssertExpectations(t)
}

// Regression test: contracts/mail.Content has no single View field — it
// splits into HtmlView and TextView. An earlier version of this channel
// wrote msg.Content.View into a field that doesn't exist, which slipped
// past every other test here because none of them set a view-template
// field at all. This asserts the mapping explicitly so that mistake
// can't reappear silently.
func TestMailChannel_Send_MapsHtmlViewAndTextView(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	mailer.On("Send", mock.MatchedBy(func(m contractsmail.Mailable) bool {
		content := m.Content()
		if content == nil {
			return false
		}
		with, _ := content.With["amount"].(string)
		return content.HtmlView == "emails.invoice" &&
			content.TextView == "emails.invoice_text" &&
			with == "99.99"
	})).Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &richNotification{
		msg: contractsnotification.MailMessage{
			Subject: "Invoice Paid",
			Content: contractsnotification.MailContent{
				HtmlView: "emails.invoice",
				TextView: "emails.invoice_text",
				With:     map[string]any{"amount": "99.99"},
			},
		},
	}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	mailer.AssertExpectations(t)
}

func TestMailChannel_Send_ReturnsError_WhenEmptyAddress(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailNotifiable{addr: ""} // no address
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty address")
	mailer.AssertNotCalled(t, "Send", mock.Anything)
}

func TestMailChannel_Send_WrapsMailerError(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	mailer.On("Send", mock.Anything).
		Return(errors.New("SMTP connection refused")).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP connection refused")
}

// Verify the Mailable adapter satisfies the contractsmail.Mailable interface at compile time.
var _ contractsmail.Mailable = (*channels.NotificationMailable)(nil)
