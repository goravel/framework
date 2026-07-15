package channels_test

import (
	"encoding/json"
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

// mailRoutableNotifiable implements MailRoutable in addition to
// Notifiable, to test multi-address mail routing.
type mailRoutableNotifiable struct {
	addr      string // used only as the RouteNotificationFor("mail") fallback
	addresses []string
}

func (m *mailRoutableNotifiable) RouteNotificationFor(channel string) string {
	if channel == "mail" {
		return m.addr
	}
	return ""
}

func (m *mailRoutableNotifiable) RouteNotificationForMail(_ contractsnotification.Notification) []string {
	return m.addresses
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

func TestMailChannel_Send_UsesMailRoutable_ForMultipleAddresses(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	mailer.On("Send", mock.MatchedBy(func(m contractsmail.Mailable) bool {
		env := m.Envelope()
		return env != nil &&
			len(env.To) == 2 &&
			env.To[0] == "primary@example.com" &&
			env.To[1] == "secondary@example.com"
	})).Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailRoutableNotifiable{
		addr:      "fallback@example.com", // should be ignored — MailRoutable takes priority
		addresses: []string{"primary@example.com", "secondary@example.com"},
	}
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	mailer.AssertExpectations(t)
}

func TestMailChannel_Send_FallsBackToRouteNotificationFor_WhenMailRoutableReturnsEmpty(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	mailer.On("Send", mock.MatchedBy(func(m contractsmail.Mailable) bool {
		env := m.Envelope()
		return env != nil && len(env.To) == 1 && env.To[0] == "fallback@example.com"
	})).Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailRoutableNotifiable{
		addr:      "fallback@example.com",
		addresses: nil, // implements MailRoutable but has nothing to say
	}
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	mailer.AssertExpectations(t)
}

func TestMailChannel_Send_ToMailOverridesResolvedAddresses_WhenSet(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	mailer.On("Send", mock.MatchedBy(func(m contractsmail.Mailable) bool {
		env := m.Envelope()
		return env != nil && len(env.To) == 1 && env.To[0] == "explicit@example.com"
	})).Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailRoutableNotifiable{
		addr:      "fallback@example.com",
		addresses: []string{"primary@example.com", "secondary@example.com"},
	}
	n := &richNotification{
		msg: contractsnotification.MailMessage{
			To:      []string{"explicit@example.com"}, // ToMail() explicitly sets To
			Subject: "Invoice Paid",
			Content: contractsnotification.MailContent{Text: "Your invoice was paid."},
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

// TestMailChannel_SendNow_BehavesIdenticallyToSend confirms the
// Send-delegates-to-SendNow relationship: calling SendNow directly
// (bypassing Manager entirely) produces the same delivery as Send.
func TestMailChannel_SendNow_BehavesIdenticallyToSend(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	mailer.On("Send", mock.AnythingOfType("*channels.NotificationMailable")).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &plainNotification{}

	err := ch.SendNow(notifiable, n)
	assert.NoError(t, err)
	mailer.AssertExpectations(t)
}

// TestMailChannel_Send_MailableExposesAllContractFields exercises every
// accessor on the internal NotificationMailable adapter (Envelope,
// Content, Attachments, Headers, Queue) directly — mockmail.NewMail's
// Send expectation intercepts before the real mail package would ever
// call these itself, so nothing else in this test file reaches them.
func TestMailChannel_Send_MailableExposesAllContractFields(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	var captured contractsmail.Mailable
	mailer.On("Send", mock.AnythingOfType("*channels.NotificationMailable")).
		Run(func(args mock.Arguments) {
			captured = args.Get(0).(contractsmail.Mailable)
		}).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &richNotification{
		msg: contractsnotification.MailMessage{
			Subject:     "Invoice Paid",
			Content:     contractsnotification.MailContent{Text: "paid"},
			Attachments: []string{"invoice.pdf"},
			Headers:     map[string]string{"X-Test": "1"},
		},
	}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)

	assert.NotNil(t, captured.Envelope())
	assert.Equal(t, "Invoice Paid", captured.Envelope().Subject)
	assert.NotNil(t, captured.Content())
	assert.Equal(t, "paid", captured.Content().Text)
	assert.Equal(t, []string{"invoice.pdf"}, captured.Attachments())
	assert.Equal(t, map[string]string{"X-Test": "1"}, captured.Headers())
	assert.Nil(t, captured.Queue())
}

// ---- Deliver-specific branch coverage ----
//
// Deliver is exported and callable directly (DispatchJob does exactly
// this on the queued path), so these exercise branches Send()/Resolve()
// never reach: an empty route, malformed payload JSON, the
// route-as-fallback-recipient path when To is empty AND route contains
// multiple comma-separated addresses, the default-subject fallback, and
// an explicit From override.

func TestMailChannel_Deliver_NoOp_WhenEmptyRoute(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t) // no calls expected

	ch := channels.NewMailChannel(mailer, logger)
	err := ch.Deliver("", []byte(`{}`))
	assert.NoError(t, err)
}

func TestMailChannel_Deliver_ReturnsError_WhenMalformedPayload(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t) // no calls expected

	ch := channels.NewMailChannel(mailer, logger)
	err := ch.Deliver("user@example.com", []byte(`{not valid json`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}

func TestMailChannel_Deliver_UsesRouteAsRecipients_WhenToEmpty(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	var captured contractsmail.Mailable
	mailer.On("Send", mock.AnythingOfType("*channels.NotificationMailable")).
		Run(func(args mock.Arguments) {
			captured = args.Get(0).(contractsmail.Mailable)
		}).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	payload, err := json.Marshal(contractsnotification.MailMessage{
		Content: contractsnotification.MailContent{Text: "hi"},
	})
	assert.NoError(t, err)

	err = ch.Deliver("a@example.com,b@example.com", payload)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a@example.com", "b@example.com"}, captured.Envelope().To)
}

func TestMailChannel_Deliver_DefaultsSubject_WhenEmpty(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	var captured contractsmail.Mailable
	mailer.On("Send", mock.AnythingOfType("*channels.NotificationMailable")).
		Run(func(args mock.Arguments) {
			captured = args.Get(0).(contractsmail.Mailable)
		}).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	payload, err := json.Marshal(contractsnotification.MailMessage{
		Content: contractsnotification.MailContent{Text: "hi"},
		// Subject deliberately left empty.
	})
	assert.NoError(t, err)

	err = ch.Deliver("user@example.com", payload)
	assert.NoError(t, err)
	assert.Equal(t, "Notification", captured.Envelope().Subject)
}

func TestMailChannel_Deliver_SetsFrom_WhenSpecified(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)

	var captured contractsmail.Mailable
	mailer.On("Send", mock.AnythingOfType("*channels.NotificationMailable")).
		Run(func(args mock.Arguments) {
			captured = args.Get(0).(contractsmail.Mailable)
		}).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer, logger)
	payload, err := json.Marshal(contractsnotification.MailMessage{
		Subject: "Hi",
		Content: contractsnotification.MailContent{Text: "hi"},
		From:    "billing@example.com",
	})
	assert.NoError(t, err)

	err = ch.Deliver("user@example.com", payload)
	assert.NoError(t, err)
	assert.Equal(t, "billing@example.com", captured.Envelope().From.Address)
}
