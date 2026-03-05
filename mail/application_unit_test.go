package mail

import (
	"errors"
	"testing"

	"github.com/goravel/framework/contracts/binding"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsmail "github.com/goravel/framework/contracts/mail"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksmail "github.com/goravel/framework/mocks/mail"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type stubMailable struct {
	attachments []string
	content     *contractsmail.Content
	envelope    *contractsmail.Envelope
	headers     map[string]string
	queue       *contractsmail.Queue
}

func (s *stubMailable) Attachments() []string             { return s.attachments }
func (s *stubMailable) Content() *contractsmail.Content   { return s.content }
func (s *stubMailable) Envelope() *contractsmail.Envelope { return s.envelope }
func (s *stubMailable) Headers() map[string]string        { return s.headers }
func (s *stubMailable) Queue() *contractsmail.Queue       { return s.queue }

func TestApplicationBuilderMethodsCloneAndMutate(t *testing.T) {
	base := &Application{}

	instance := base.To([]string{"to@example.com"}).(*Application)
	assert.NotSame(t, base, instance)
	assert.Empty(t, base.params.To)
	assert.Equal(t, []string{"to@example.com"}, instance.params.To)

	contentWith := map[string]any{"name": "goravel"}
	chained := instance.
		Subject("subject").
		Cc([]string{"cc@example.com"}).
		Bcc([]string{"bcc@example.com"}).
		Attach([]string{"/tmp/a.txt"}).
		From(Address("from@example.com", "From Name")).
		Headers(map[string]string{"X-Test": "yes"}).
		Content(contractsmail.Content{Html: "<h1>Hello</h1>", View: "mail.tmpl", Text: "mail.txt", With: contentWith}).(*Application)

	assert.Same(t, instance, chained)
	assert.Equal(t, "subject", instance.params.Subject)
	assert.Equal(t, []string{"cc@example.com"}, instance.params.CC)
	assert.Equal(t, []string{"bcc@example.com"}, instance.params.BCC)
	assert.Equal(t, []string{"/tmp/a.txt"}, instance.params.Attachments)
	assert.Equal(t, "from@example.com", instance.params.FromAddress)
	assert.Equal(t, "From Name", instance.params.FromName)
	assert.Equal(t, map[string]string{"X-Test": "yes"}, instance.params.Headers)
	assert.Equal(t, "<h1>Hello</h1>", instance.params.HTML)
	assert.Equal(t, "mail.tmpl", instance.view)
	assert.Equal(t, "mail.txt", instance.text)
	assert.Equal(t, contentWith, instance.with)
}

func TestApplicationSetUsingMailable(t *testing.T) {
	app := &Application{params: Params{HTML: "old-html", Subject: "old-subject"}}

	mailable := &stubMailable{
		attachments: []string{"/tmp/logo.png"},
		content: &contractsmail.Content{
			Html: "new-html",
			View: "email.tmpl",
			Text: "email.txt",
			With: map[string]any{"name": "goravel"},
		},
		envelope: &contractsmail.Envelope{
			From:    contractsmail.Address{Address: "from@example.com", Name: "Mailer"},
			To:      []string{"to@example.com"},
			Cc:      []string{"cc@example.com"},
			Bcc:     []string{"bcc@example.com"},
			Subject: "new-subject",
		},
		headers: map[string]string{"X-Test": "true"},
	}

	app.setUsingMailable(mailable)

	assert.Equal(t, "new-html", app.params.HTML)
	assert.Equal(t, []string{"/tmp/logo.png"}, app.params.Attachments)
	assert.Equal(t, map[string]string{"X-Test": "true"}, app.params.Headers)
	assert.Equal(t, "from@example.com", app.params.FromAddress)
	assert.Equal(t, "Mailer", app.params.FromName)
	assert.Equal(t, []string{"to@example.com"}, app.params.To)
	assert.Equal(t, []string{"cc@example.com"}, app.params.CC)
	assert.Equal(t, []string{"bcc@example.com"}, app.params.BCC)
	assert.Equal(t, "new-subject", app.params.Subject)
	assert.Equal(t, "email.tmpl", app.view)
	assert.Equal(t, "email.txt", app.text)
	assert.Equal(t, map[string]any{"name": "goravel"}, app.with)
}

func TestApplicationRenderViewTemplate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		template := mocksmail.NewTemplate(t)
		template.EXPECT().Render("mail.tmpl", map[string]any{"id": 1}).Return("<h1>Hello</h1>", nil).Once()
		template.EXPECT().Render("mail.txt", map[string]any{"id": 1}).Return("Hello", nil).Once()

		app := &Application{template: template, view: "mail.tmpl", text: "mail.txt", with: map[string]any{"id": 1}}

		err := app.renderViewTemplate()
		assert.NoError(t, err)
		assert.Equal(t, "<h1>Hello</h1>", app.params.HTML)
		assert.Equal(t, "Hello", app.params.Text)
	})

	t.Run("html render failed", func(t *testing.T) {
		template := mocksmail.NewTemplate(t)
		template.EXPECT().Render("mail.tmpl", mock.Anything).Return("", errors.New("render failed")).Once()

		app := &Application{template: template, view: "mail.tmpl", with: map[string]any{"id": 1}}
		err := app.renderViewTemplate()
		assert.ErrorContains(t, err, "render failed")
	})

	t.Run("text render failed", func(t *testing.T) {
		template := mocksmail.NewTemplate(t)
		template.EXPECT().Render("mail.txt", mock.Anything).Return("", errors.New("text render failed")).Once()

		app := &Application{template: template, text: "mail.txt", with: map[string]any{"id": 1}}
		err := app.renderViewTemplate()
		assert.ErrorContains(t, err, "text render failed")
	})
}

func TestApplicationQueue(t *testing.T) {
	mockQueue := mocksqueue.NewQueue(t)
	pendingJob := mocksqueue.NewPendingJob(t)
	mockConfig := mocksconfig.NewConfig(t)

	mailable := &stubMailable{
		attachments: []string{"/tmp/logo.png"},
		content:     &contractsmail.Content{Html: "<h1>Queue</h1>"},
		envelope: &contractsmail.Envelope{
			From:    contractsmail.Address{Address: "from@example.com", Name: "From"},
			To:      []string{"to@example.com"},
			Cc:      []string{"cc@example.com"},
			Bcc:     []string{"bcc@example.com"},
			Subject: "queue-subject",
		},
		headers: map[string]string{"X-Test": "queue"},
		queue:   &contractsmail.Queue{Connection: "redis", Queue: "emails"},
	}

	mockQueue.EXPECT().Job(mock.Anything, mock.Anything).
		Run(func(_ contractsqueue.Job, args ...[]contractsqueue.Arg) {
			assert.Len(t, args, 1)
			assert.Len(t, args[0], 10)
			assert.Equal(t, "queue-subject", args[0][0].Value)
			assert.Equal(t, "<h1>Queue</h1>", args[0][1].Value)
			assert.Equal(t, "from@example.com", args[0][3].Value)
			assert.Equal(t, []string{"to@example.com"}, args[0][5].Value)
			assert.Equal(t, []string{"/tmp/logo.png"}, args[0][8].Value)
			assert.Equal(t, []string{"X-Test: queue"}, args[0][9].Value)
		}).
		Return(pendingJob).Once()
	pendingJob.EXPECT().OnConnection("redis").Return(pendingJob).Once()
	pendingJob.EXPECT().OnQueue("emails").Return(pendingJob).Once()
	pendingJob.EXPECT().Dispatch().Return(nil).Once()

	app := &Application{config: mockConfig, queue: mockQueue}

	err := app.Queue(mailable)
	assert.NoError(t, err)
}

func TestApplicationQueueRenderError(t *testing.T) {
	template := mocksmail.NewTemplate(t)
	template.EXPECT().Render("mail.tmpl", mock.Anything).Return("", errors.New("render failed")).Once()

	app := &Application{template: template, view: "mail.tmpl", with: map[string]any{"id": 1}}

	err := app.Queue()
	assert.ErrorContains(t, err, "render failed")
}

func TestApplicationSendRenderError(t *testing.T) {
	template := mocksmail.NewTemplate(t)
	template.EXPECT().Render("mail.tmpl", mock.Anything).Return("", errors.New("render failed")).Once()

	app := &Application{template: template}
	err := app.Send(&stubMailable{content: &contractsmail.Content{View: "mail.tmpl", With: map[string]any{"id": 1}}})

	assert.ErrorContains(t, err, "render failed")
}

func TestLoginAuth(t *testing.T) {
	auth := LoginAuth("user", "pass")

	method, payload, err := auth.Start(nil)
	assert.NoError(t, err)
	assert.Equal(t, "LOGIN", method)
	assert.Equal(t, []byte("user"), payload)

	next, err := auth.Next([]byte("Username:"), true)
	assert.NoError(t, err)
	assert.Equal(t, []byte("user"), next)

	next, err = auth.Next([]byte("Password:"), true)
	assert.NoError(t, err)
	assert.Equal(t, []byte("pass"), next)

	next, err = auth.Next([]byte("Unknown:"), true)
	assert.NoError(t, err)
	assert.Nil(t, next)

	next, err = auth.Next([]byte("Username:"), false)
	assert.NoError(t, err)
	assert.Nil(t, next)
}

func TestNewApplication(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("mail.template.default", "html").Return("mail_unit_success").Once()
		mockConfig.EXPECT().GetString("mail.template.engines.mail_unit_success.driver", "html").Return("html").Once()
		mockConfig.EXPECT().GetString("mail.template.engines.mail_unit_success.path", "resources/views/mail").Return(".").Once()

		app, err := NewApplication(mockConfig, nil)
		assert.NoError(t, err)
		assert.NotNil(t, app)
		assert.Equal(t, mockConfig, app.config)
	})

	t.Run("unsupported template driver", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("mail.template.default", "html").Return("mail_unit_fail").Once()
		mockConfig.EXPECT().GetString("mail.template.engines.mail_unit_fail.driver", "html").Return("unsupported").Once()

		app, err := NewApplication(mockConfig, nil)
		assert.Nil(t, app)
		assert.Error(t, err)
	})
}

func TestSendMailAttachmentError(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.from.address").Return("from@example.com").Once()
	mockConfig.EXPECT().GetString("mail.from.name").Return("From").Once()

	err := SendMail(mockConfig, Params{
		To:          []string{"to@example.com"},
		Attachments: []string{"/tmp/does-not-exist.txt"},
	})
	assert.Error(t, err)
}

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}
	relation := provider.Relationship()

	assert.Equal(t, 1, len(relation.Bindings))
	assert.Equal(t, binding.Mail, relation.Bindings[0])
	assert.Equal(t, binding.Bindings[binding.Mail].Dependencies, relation.Dependencies)
	assert.Empty(t, relation.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)

	app.EXPECT().Bind(mock.Anything, mock.Anything).
		Run(func(_ interface{}, callback func(contractsfoundation.Application) (interface{}, error)) {
			withoutConfig := mocksfoundation.NewApplication(t)
			withoutConfig.EXPECT().MakeConfig().Return(nil).Once()

			instance, err := callback(withoutConfig)
			assert.Nil(t, instance)
			assert.Error(t, err)

			withoutQueue := mocksfoundation.NewApplication(t)
			configOnly := mocksconfig.NewConfig(t)
			withoutQueue.EXPECT().MakeConfig().Return(configOnly).Once()
			withoutQueue.EXPECT().MakeQueue().Return(nil).Once()

			instance, err = callback(withoutQueue)
			assert.Nil(t, instance)
			assert.Error(t, err)

			withAll := mocksfoundation.NewApplication(t)
			configAndQueue := mocksconfig.NewConfig(t)
			queue := mocksqueue.NewQueue(t)
			withAll.EXPECT().MakeConfig().Return(configAndQueue).Once()
			withAll.EXPECT().MakeQueue().Return(queue).Once()
			configAndQueue.EXPECT().GetString("mail.template.default", "html").Return("mail_service_provider").Once()
			configAndQueue.EXPECT().GetString("mail.template.engines.mail_service_provider.driver", "html").Return("html").Once()
			configAndQueue.EXPECT().GetString("mail.template.engines.mail_service_provider.path", "resources/views/mail").Return(".").Once()

			instance, err = callback(withAll)
			assert.NoError(t, err)
			assert.NotNil(t, instance)
		}).
		Once()

	provider.Register(app)
}

func TestServiceProviderBootAndRegisterJobs(t *testing.T) {
	t.Run("boot registers command and job", func(t *testing.T) {
		provider := &ServiceProvider{}
		app := mocksfoundation.NewApplication(t)
		queue := mocksqueue.NewQueue(t)
		config := mocksconfig.NewConfig(t)

		app.EXPECT().Commands(mock.Anything).Once()
		app.EXPECT().MakeQueue().Return(queue).Once()
		app.EXPECT().MakeConfig().Return(config).Once()
		queue.EXPECT().Register(mock.Anything).
			Run(func(jobs []contractsqueue.Job) {
				assert.Len(t, jobs, 1)
				assert.Equal(t, "goravel_send_mail_job", jobs[0].Signature())
			}).
			Once()

		provider.Boot(app)
	})

	t.Run("registerJobs skips when queue is nil", func(t *testing.T) {
		provider := &ServiceProvider{}
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().MakeQueue().Return(nil).Once()

		provider.registerJobs(app)
	})

	t.Run("registerJobs skips when config is nil", func(t *testing.T) {
		provider := &ServiceProvider{}
		app := mocksfoundation.NewApplication(t)
		queue := mocksqueue.NewQueue(t)
		app.EXPECT().MakeQueue().Return(queue).Once()
		app.EXPECT().MakeConfig().Return(nil).Once()

		provider.registerJobs(app)
	})
}
