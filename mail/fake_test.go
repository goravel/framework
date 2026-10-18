package mail

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/binding"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/mail"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksmail "github.com/goravel/framework/mocks/mail"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

type FakeTestSuite struct {
	suite.Suite
	mockConfig   *mocksconfig.Config
	mockTemplate *mocksmail.Template
	app          *Application
}

func TestFakeTestSuite(t *testing.T) {
	suite.Run(t, new(FakeTestSuite))
}

func (s *FakeTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockTemplate = mocksmail.NewTemplate(s.T())
	// No queue: a dispatch while faking would panic.
	s.app = &Application{config: s.mockConfig, template: s.mockTemplate, fake: NewFake()}
}

func (s *FakeTestSuite) TestSendWithTheBuilder() {
	s.mockConfig.EXPECT().GetString("mail.from.address").Return("noreply@goravel.dev").Once()
	s.mockConfig.EXPECT().GetString("mail.from.name").Return("Goravel").Once()

	fake := s.app.Fake()
	s.True(fake.AssertNothingOutgoing())

	s.NoError(s.app.To([]string{"Ana@Example.com"}).
		Cc([]string{"cc@example.com"}).
		Bcc([]string{"bcc@example.com"}).
		Attach([]string{"invoice.pdf"}).
		Subject("Welcome").
		Content(mail.Content{Html: "<h1>Hi</h1>", Text: "Hi"}).
		Headers(map[string]string{"X-Tag": "welcome"}).
		Send())

	s.True(fake.AssertSentCount(1))
	s.True(fake.AssertNothingQueued())
	s.False(fake.AssertNothingSent())
	s.False(fake.AssertNothingOutgoing())
	s.True(fake.AssertSent(func(message mail.Message) bool {
		return message.HasTo("ana@example.com") && message.HasCc("cc@example.com") &&
			message.HasBcc("bcc@example.com") && message.HasAttachment("invoice.pdf") &&
			message.Subject == "Welcome"
	}))
	s.True(fake.AssertNotSent(func(message mail.Message) bool {
		return message.HasTo("someone-else@example.com")
	}))

	s.Equal([]mail.Message{{
		Subject:     "Welcome",
		Html:        "<h1>Hi</h1>",
		Text:        "Hi",
		From:        mail.Address{Address: "noreply@goravel.dev", Name: "Goravel"},
		To:          []string{"Ana@Example.com"},
		Cc:          []string{"cc@example.com"},
		Bcc:         []string{"bcc@example.com"},
		Attachments: []string{"invoice.pdf"},
		Headers:     map[string]string{"X-Tag": "welcome"},
	}}, fake.Sent())
}

func (s *FakeTestSuite) TestSendMailableRendersTheViews() {
	mailable := &stubMailable{
		content: &mail.Content{HtmlView: "welcome.tmpl", With: map[string]any{"name": "Ana"}},
		envelope: &mail.Envelope{
			From:    mail.Address{Address: "team@goravel.dev", Name: "Team"},
			To:      []string{"ana@example.com"},
			Subject: "Hello Ana",
		},
	}
	s.mockTemplate.EXPECT().Render("welcome.tmpl", map[string]any{"name": "Ana"}).Return("<p>Hello Ana</p>", nil).Once()

	fake := s.app.Fake()
	s.NoError(s.app.Send(mailable))

	sent := fake.Sent()
	s.Require().Len(sent, 1)
	s.Same(mailable, sent[0].Mailable)
	s.Equal("<p>Hello Ana</p>", sent[0].Html)
	s.Equal(mail.Address{Address: "team@goravel.dev", Name: "Team"}, sent[0].From)
	s.True(fake.AssertSent(func(message mail.Message) bool {
		_, ok := message.Mailable.(*stubMailable)
		return ok && message.HasTo("ana@example.com")
	}))
}

func (s *FakeTestSuite) TestTemplateErrorsStillFail() {
	mailable := &stubMailable{content: &mail.Content{HtmlView: "missing.tmpl"}}
	s.mockTemplate.EXPECT().Render("missing.tmpl", map[string]any(nil)).Return("", assert.AnError).Once()

	fake := s.app.Fake()

	s.ErrorIs(s.app.Send(mailable), assert.AnError)
	s.True(fake.AssertNothingSent())
}

func (s *FakeTestSuite) TestQueue() {
	mailable := &stubMailable{
		envelope: &mail.Envelope{From: mail.Address{Address: "team@goravel.dev"}, To: []string{"ana@example.com"}, Subject: "Queued"},
		queue:    &mail.Queue{Connection: "redis", Queue: "mails"},
	}

	fake := s.app.Fake()
	s.NoError(s.app.Queue(mailable))

	s.True(fake.AssertNothingSent())
	s.True(fake.AssertQueuedCount(1))
	s.True(fake.AssertQueued(func(message mail.Message) bool {
		return message.Subject == "Queued" && message.Queue != nil && message.Queue.Queue == "mails"
	}))
	s.True(fake.AssertNotQueued(func(message mail.Message) bool {
		return message.Subject == "Other"
	}))
	s.Equal(&mail.Queue{Connection: "redis", Queue: "mails"}, fake.Queued()[0].Queue)
}

func (s *FakeTestSuite) TestFilters() {
	s.mockConfig.EXPECT().GetString("mail.from.address").Return("noreply@goravel.dev")
	s.mockConfig.EXPECT().GetString("mail.from.name").Return("Goravel")

	fake := s.app.Fake()
	for _, subject := range []string{"a", "b", "a"} {
		s.NoError(s.app.Subject(subject).Send())
	}

	s.Len(fake.Sent(), 3)
	s.Len(fake.Sent(nil), 3)
	s.Len(fake.Sent(func(message mail.Message) bool { return message.Subject == "a" }), 2)
	s.Empty(fake.Queued(func(message mail.Message) bool { return true }))
}

func (s *FakeTestSuite) TestResetAndFakeAgain() {
	s.mockConfig.EXPECT().GetString("mail.from.address").Return("")
	s.mockConfig.EXPECT().GetString("mail.from.name").Return("")

	fake := s.app.Fake()
	s.NoError(s.app.Subject("one").Send())
	s.True(fake.AssertSentCount(1))

	fake.Reset()
	s.False(s.app.fake.isEnabled())
	s.True(fake.AssertNothingSent())

	// Faking again starts with empty records.
	fake = s.app.Fake()
	s.NoError(s.app.Subject("two").Send())
	s.Equal("two", fake.Sent()[0].Subject)
	s.True(fake.AssertSentCount(1))
}

func (s *FakeTestSuite) TestConcurrentSends() {
	s.mockConfig.EXPECT().GetString("mail.from.address").Return("")
	s.mockConfig.EXPECT().GetString("mail.from.name").Return("")

	fake := s.app.Fake()

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.NoError(s.app.Subject(fmt.Sprint(i)).Send())
		}()
	}
	wg.Wait()

	s.True(fake.AssertSentCount(50))
}

// The facade resolves a new Application on every call: the fake enabled on
// one of them must apply to the next ones, like the notification mailer.
func TestServiceProviderSharesTheFake(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	var resolve func(contractsfoundation.Application) (any, error)
	mockApp.EXPECT().Bind(binding.Mail, mock.Anything).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
		resolve = callback
	}).Once()

	(&ServiceProvider{}).Register(mockApp)

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.default", "html").Return("html")
	mockConfig.EXPECT().GetString("mail.template.engines.html.driver", "html").Return("html").Maybe()
	mockConfig.EXPECT().GetString("mail.template.engines.html.path", "resources/views/mail").Return("resources/views/mail").Maybe()
	mockConfig.EXPECT().GetString("mail.from.address").Return("noreply@goravel.dev")
	mockConfig.EXPECT().GetString("mail.from.name").Return("Goravel")
	mockApp.EXPECT().MakeConfig().Return(mockConfig).Twice()
	mockApp.EXPECT().MakeQueue().Return(mocksqueue.NewQueue(t)).Twice()

	first, err := resolve(mockApp)
	assert.NoError(t, err)
	second, err := resolve(mockApp)
	assert.NoError(t, err)
	assert.NotSame(t, first, second)

	fake := first.(mail.Mail).Fake()
	assert.NoError(t, second.(mail.Mail).To([]string{"ana@example.com"}).Subject("Shared").Send())

	assert.True(t, fake.AssertSent(func(message mail.Message) bool {
		return message.Subject == "Shared" && message.HasTo("ana@example.com")
	}))
	fake.Reset()
}

func TestMessageHelpers(t *testing.T) {
	message := mail.Message{
		To:          []string{" Ana@Example.com "},
		Cc:          []string{"cc@example.com"},
		Bcc:         []string{"bcc@example.com"},
		Attachments: []string{"storage/invoice.pdf"},
	}

	assert.True(t, message.HasTo("ana@example.com"))
	assert.False(t, message.HasTo("cc@example.com"))
	assert.True(t, message.HasCc("CC@example.com"))
	assert.True(t, message.HasBcc("bcc@example.com"))
	assert.True(t, message.HasAttachment("storage/invoice.pdf"))
	assert.False(t, message.HasAttachment("invoice.pdf"))
}
