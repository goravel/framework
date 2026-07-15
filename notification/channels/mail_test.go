package channels_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsmail "github.com/goravel/framework/contracts/mail"
	contractsnotification "github.com/goravel/framework/contracts/notification"
	mocksmail "github.com/goravel/framework/mocks/mail"
	"github.com/goravel/framework/notification/channels"
	"github.com/goravel/framework/notification/mailmessage"
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
// Notifiable, to test multi-address mail routing. routes is
// address→name, matching RouteNotificationForMail's map[string]string.
type mailRoutableNotifiable struct {
	addr   string // used only as the RouteNotificationFor("mail") fallback
	routes map[string]string
}

func (m *mailRoutableNotifiable) RouteNotificationFor(channel string) string {
	if channel == "mail" {
		return m.addr
	}
	return ""
}

func (m *mailRoutableNotifiable) RouteNotificationForMail(_ contractsnotification.Notification) map[string]string {
	return m.routes
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
	ch := channels.NewMailChannel(nil)
	assert.Equal(t, "mail", ch.Name())
}

func TestMailChannel_Send_UsesDefaultMessage_WhenNotMailableNotification(t *testing.T) {
	mailer := mocksmail.NewMail(t)

	mailer.EXPECT().Send(mock.AnythingOfType("*channels.NotificationMailable")).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
}

func TestMailChannel_Send_UsesToMail_WhenMailableNotification(t *testing.T) {
	mailer := mocksmail.NewMail(t)

	mailer.EXPECT().Send(mock.AnythingOfType("*channels.NotificationMailable")).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &richNotification{
		msg: mailmessage.NewMailMessage().
			Subject("Invoice Paid").
			Text("Your invoice was paid.").
			Build(),
	}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
}

// Regression test: contracts/mail.Content has no single View field — it
// splits into HtmlView and TextView. An earlier version of this channel
// wrote msg.Content.View into a field that doesn't exist, which slipped
// past every other test here because none of them set a view-template
// field at all. This asserts the mapping explicitly so that mistake
// can't reappear silently.
func TestMailChannel_Send_MapsHtmlViewAndTextView(t *testing.T) {
	mailer := mocksmail.NewMail(t)

	mailer.EXPECT().Send(mock.MatchedBy(func(m contractsmail.Mailable) bool {
		content := m.Content()
		if content == nil {
			return false
		}
		with, _ := content.With["amount"].(string)
		return content.HtmlView == "emails.invoice" &&
			content.TextView == "emails.invoice_text" &&
			with == "99.99"
	})).Return(nil).Once()

	ch := channels.NewMailChannel(mailer)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &richNotification{
		msg: mailmessage.NewMailMessage().
			Subject("Invoice Paid").
			HtmlView("emails.invoice", map[string]any{"amount": "99.99"}).
			TextView("emails.invoice_text", map[string]any{"amount": "99.99"}).
			Build(),
	}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
}

func TestMailChannel_Send_UsesMailRoutable_ForMultipleAddresses(t *testing.T) {
	mailer := mocksmail.NewMail(t)

	mailer.EXPECT().Send(mock.MatchedBy(func(m contractsmail.Mailable) bool {
		env := m.Envelope()
		if env == nil || len(env.To) != 2 {
			return false
		}
		// formatAddresses sorts by address for determinism.
		return env.To[0] == "Jane Doe <primary@example.com>" &&
			env.To[1] == "secondary@example.com"
	})).Return(nil).Once()

	ch := channels.NewMailChannel(mailer)
	notifiable := &mailRoutableNotifiable{
		addr: "fallback@example.com", // should be ignored — MailRoutable takes priority
		routes: map[string]string{
			"primary@example.com":   "Jane Doe",
			"secondary@example.com": "", // no display name
		},
	}
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
}

func TestMailChannel_Send_FallsBackToRouteNotificationFor_WhenMailRoutableReturnsEmpty(t *testing.T) {
	mailer := mocksmail.NewMail(t)

	mailer.EXPECT().Send(mock.MatchedBy(func(m contractsmail.Mailable) bool {
		env := m.Envelope()
		return env != nil && len(env.To) == 1 && env.To[0] == "fallback@example.com"
	})).Return(nil).Once()

	ch := channels.NewMailChannel(mailer)
	notifiable := &mailRoutableNotifiable{
		addr:   "fallback@example.com",
		routes: nil, // implements MailRoutable but has nothing to say
	}
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
}

func TestMailChannel_Send_ToMailOverridesResolvedAddresses_WhenSet(t *testing.T) {
	mailer := mocksmail.NewMail(t)

	mailer.EXPECT().Send(mock.MatchedBy(func(m contractsmail.Mailable) bool {
		env := m.Envelope()
		return env != nil && len(env.To) == 1 && env.To[0] == "explicit@example.com"
	})).Return(nil).Once()

	ch := channels.NewMailChannel(mailer)
	notifiable := &mailRoutableNotifiable{
		addr: "fallback@example.com",
		routes: map[string]string{
			"primary@example.com":   "",
			"secondary@example.com": "",
		},
	}
	n := &richNotification{
		msg: mailmessage.NewMailMessage().
			To("explicit@example.com"). // ToMail() explicitly sets To
			Subject("Invoice Paid").
			Text("Your invoice was paid.").
			Build(),
	}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
}

func TestMailChannel_Send_ReturnsError_WhenEmptyAddress(t *testing.T) {
	mailer := mocksmail.NewMail(t) // no calls expected

	ch := channels.NewMailChannel(mailer)
	notifiable := &mailNotifiable{addr: ""} // no address
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty address")
}

func TestMailChannel_Send_WrapsMailerError(t *testing.T) {
	mailer := mocksmail.NewMail(t)

	mailer.EXPECT().Send(mock.Anything).
		Return(errors.New("SMTP connection refused")).Once()

	ch := channels.NewMailChannel(mailer)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &plainNotification{}

	err := ch.Send(notifiable, n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP connection refused")
}

// Verify the Mailable adapter satisfies the contractsmail.Mailable interface at compile time.
var _ contractsmail.Mailable = (*channels.NotificationMailable)(nil)

// TestMailChannel_Send_MailableExposesAllContractFields exercises every
// accessor on the internal NotificationMailable adapter (Envelope,
// Content, Attachments, Headers, Queue) directly — mocksmail.NewMail's
// Send expectation intercepts before the real mail package would ever
// call these itself, so nothing else in this test file reaches them.
func TestMailChannel_Send_MailableExposesAllContractFields(t *testing.T) {
	mailer := mocksmail.NewMail(t)

	var captured contractsmail.Mailable
	mailer.EXPECT().Send(mock.AnythingOfType("*channels.NotificationMailable")).
		Run(func(mailable ...contractsmail.Mailable) {
			captured = mailable[0]
		}).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer)
	notifiable := &mailNotifiable{addr: "user@example.com"}
	n := &richNotification{
		msg: mailmessage.NewMailMessage().
			Subject("Invoice Paid").
			Text("paid").
			Attach("invoice.pdf").
			Header("X-Test", "1").
			Build(),
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
	mailer := mocksmail.NewMail(t) // no calls expected

	ch := channels.NewMailChannel(mailer)
	err := ch.Deliver("", []byte(`{}`))
	assert.NoError(t, err)
}

func TestMailChannel_Deliver_ReturnsError_WhenMalformedPayload(t *testing.T) {
	mailer := mocksmail.NewMail(t) // no calls expected

	ch := channels.NewMailChannel(mailer)
	err := ch.Deliver("user@example.com", []byte(`{not valid json`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}

func TestMailChannel_Deliver_UsesRouteAsRecipients_WhenToEmpty(t *testing.T) {
	mailer := mocksmail.NewMail(t)

	var captured contractsmail.Mailable
	mailer.EXPECT().Send(mock.AnythingOfType("*channels.NotificationMailable")).
		Run(func(mailable ...contractsmail.Mailable) {
			captured = mailable[0]
		}).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer)
	payload, err := json.Marshal(mailmessage.NewMailMessage().Text("hi").Build())
	assert.NoError(t, err)

	err = ch.Deliver("a@example.com,b@example.com", payload)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a@example.com", "b@example.com"}, captured.Envelope().To)
}

func TestMailChannel_Deliver_DefaultsSubject_WhenEmpty(t *testing.T) {
	mailer := mocksmail.NewMail(t)

	var captured contractsmail.Mailable
	mailer.EXPECT().Send(mock.AnythingOfType("*channels.NotificationMailable")).
		Run(func(mailable ...contractsmail.Mailable) {
			captured = mailable[0]
		}).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer)
	payload, err := json.Marshal(mailmessage.NewMailMessage().Text("hi").Build()) // Subject deliberately left empty
	assert.NoError(t, err)

	err = ch.Deliver("user@example.com", payload)
	assert.NoError(t, err)
	assert.Equal(t, "Notification", captured.Envelope().Subject)
}

func TestMailChannel_Deliver_SetsFrom_WhenSpecified(t *testing.T) {
	mailer := mocksmail.NewMail(t)

	var captured contractsmail.Mailable
	mailer.EXPECT().Send(mock.AnythingOfType("*channels.NotificationMailable")).
		Run(func(mailable ...contractsmail.Mailable) {
			captured = mailable[0]
		}).
		Return(nil).Once()

	ch := channels.NewMailChannel(mailer)
	payload, err := json.Marshal(mailmessage.NewMailMessage().
		Subject("Hi").
		Text("hi").
		From("billing@example.com").
		Build())
	assert.NoError(t, err)

	err = ch.Deliver("user@example.com", payload)
	assert.NoError(t, err)
	assert.Equal(t, "billing@example.com", captured.Envelope().From.Address)
}
