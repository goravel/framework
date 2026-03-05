package mail

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/mail"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/foundation/json"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksmail "github.com/goravel/framework/mocks/mail"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/queue"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

var testBcc, testCc, testTo, testFromAddress, testFromName string

type ApplicationTestSuite struct {
	suite.Suite
	mockConfig *mocksconfig.Config
}

func TestApplicationTestSuite(t *testing.T) {
	if !file.Exists(support.EnvFilePath) && os.Getenv("MAIL_HOST") == "" {
		color.Errorln("No mail tests run, need create .env based on .env.example, then initialize it")
		return
	}

	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {
}

func (s *ApplicationTestSuite) TestSendMail() {
	s.mockConfig = mockConfig(465)

	app, err := NewApplication(s.mockConfig, nil)
	s.Nil(err)
	s.Nil(app.To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Subject("Goravel Test 465").
		Content(Html("<h1>Hello Goravel</h1>")).
		Headers(map[string]string{"Test-Mailer-Port": "465"}).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailViaTemplate() {
	s.mockConfig = mockConfig(465)

	app, err := NewApplication(s.mockConfig, nil)
	s.Nil(err)
	s.Nil(app.To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Subject("Goravel Test template").
		Content(mail.Content{
			View: "test.tmpl",
			With: map[string]any{
				"name": "Goravel",
			},
		}).
		Headers(map[string]string{"Test-Mailer-Port": "465"}).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailWithFromBy587Port() {
	s.mockConfig = mockConfig(587)

	app, err := NewApplication(s.mockConfig, nil)
	s.Nil(err)
	s.Nil(app.From(Address(testFromAddress, testFromName)).
		To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Subject("Goravel Test 587 With From").
		Content(Html("<h1>Hello Goravel</h1>")).
		Headers(map[string]string{"Test-Mailer-Port": "587", "Test-Mailer-From": testFromAddress}).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailWithMailable() {
	s.mockConfig = mockConfig(465)

	app, err := NewApplication(s.mockConfig, nil)
	s.Nil(err)
	s.Nil(app.Send(NewTestMailable()))
}

func (s *ApplicationTestSuite) TestQueueMail() {
	s.mockConfig = mockConfig(465)
	s.mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	s.mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return("default").Once()
	s.mockConfig.EXPECT().GetInt("queue.connections.redis.concurrent", 1).Return(2).Once()
	s.mockConfig.EXPECT().GetString("app.name", "goravel").Return("goravel").Once()
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().GetString("queue.failed.database").Return("mysql").Once()
	s.mockConfig.EXPECT().GetString("queue.failed.table").Return("failed_jobs").Once()

	queueFacade := queue.NewApplication(queue.NewConfig(s.mockConfig), nil, nil, queue.NewJobStorer(), json.New(), nil)
	queueFacade.Register([]contractsqueue.Job{
		NewSendMailJob(s.mockConfig),
	})

	app, err := NewApplication(s.mockConfig, queueFacade)
	s.Nil(err)

	s.Nil(app.To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Subject("Goravel Test Queue").
		Content(Html("<h1>Hello Goravel</h1>")).
		Headers(map[string]string{"Test-Mailer": "QueueMail"}).
		Queue())
	time.Sleep(3 * time.Second)
}

func (s *ApplicationTestSuite) TestQueueMailWithMailable() {
	s.mockConfig = mockConfig(465)
	s.mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	s.mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return("default").Once()
	s.mockConfig.EXPECT().GetInt("queue.connections.redis.concurrent", 1).Return(2).Once()
	s.mockConfig.EXPECT().GetString("app.name", "goravel").Return("goravel").Once()
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().GetString("queue.failed.database").Return("mysql").Once()
	s.mockConfig.EXPECT().GetString("queue.failed.table").Return("failed_jobs").Once()

	queueFacade := queue.NewApplication(queue.NewConfig(s.mockConfig), nil, nil, queue.NewJobStorer(), json.New(), nil)
	queueFacade.Register([]contractsqueue.Job{
		NewSendMailJob(s.mockConfig),
	})

	app, err := NewApplication(s.mockConfig, queueFacade)
	s.Nil(err)
	s.Nil(app.Queue(NewTestMailable()))
}

func mockConfig(mailPort int) *mocksconfig.Config {
	config := &mocksconfig.Config{}
	config.EXPECT().GetString("app.name").Return("goravel")
	config.EXPECT().GetString("queue.default").Return("sync")
	config.EXPECT().GetString("queue.connections.sync.queue", "default").Return("default")
	config.EXPECT().GetString("queue.connections.sync.driver").Return("sync")
	config.EXPECT().GetInt("queue.connections.sync.concurrent", 1).Return(1)
	config.EXPECT().GetString("queue.failed.database").Return("database")
	config.EXPECT().GetString("queue.failed.table").Return("failed_jobs")
	if file.Exists(support.EnvFilePath) {
		vip := viper.New()
		vip.SetConfigName(support.EnvFilePath)
		vip.SetConfigType("env")
		vip.AddConfigPath(".")
		_ = vip.ReadInConfig()
		vip.SetEnvPrefix("goravel")
		vip.AutomaticEnv()

		config.EXPECT().GetString("mail.host").Return(vip.GetString("MAIL_HOST"))
		config.EXPECT().GetInt("mail.port").Return(mailPort)
		config.EXPECT().GetString("mail.from.address").Return(vip.GetString("MAIL_FROM_ADDRESS"))
		config.EXPECT().GetString("mail.from.name").Return(vip.GetString("MAIL_FROM_NAME"))
		config.EXPECT().GetString("mail.username").Return(vip.GetString("MAIL_USERNAME"))
		config.EXPECT().GetString("mail.password").Return(vip.GetString("MAIL_PASSWORD"))
		config.EXPECT().GetString("mail.to").Return(vip.GetString("MAIL_TO"))
		config.EXPECT().GetString("mail.cc").Return(vip.GetString("MAIL_CC"))
		config.EXPECT().GetString("mail.bcc").Return(vip.GetString("MAIL_BCC"))
		config.EXPECT().GetString("mail.template.default", "html").Return("html").Once()
		config.EXPECT().GetString("mail.template.engines.html.driver", "html").Return("html").Once()
		config.EXPECT().GetString("mail.template.engines.html.path", "resources/views/mail").
			Return("resources/views/mail").Once()

		testFromAddress = vip.Get("MAIL_FROM_ADDRESS").(string)
		testFromName = vip.Get("MAIL_FROM_NAME").(string)
		testTo = vip.Get("MAIL_TO").(string)
	}
	if os.Getenv("MAIL_HOST") != "" {
		config.EXPECT().GetString("mail.host").Return(os.Getenv("MAIL_HOST"))
		config.EXPECT().GetInt("mail.port").Return(mailPort)
		config.EXPECT().GetString("mail.from.address").Return(os.Getenv("MAIL_FROM_ADDRESS"))
		config.EXPECT().GetString("mail.from.name").Return(os.Getenv("MAIL_FROM_NAME"))
		config.EXPECT().GetString("mail.username").Return(os.Getenv("MAIL_USERNAME"))
		config.EXPECT().GetString("mail.password").Return(os.Getenv("MAIL_PASSWORD"))
		config.EXPECT().GetString("mail.to").Return(os.Getenv("MAIL_TO"))
		config.EXPECT().GetString("mail.cc").Return(os.Getenv("MAIL_CC"))
		config.EXPECT().GetString("mail.bcc").Return(os.Getenv("MAIL_BCC"))
		config.EXPECT().GetString("mail.template.default", "html").Return("html").Once()
		config.EXPECT().GetString("mail.template.engines.html.driver", "html").Return("html").Once()
		config.EXPECT().GetString("mail.template.engines.html.path", "resources/views/mail").
			Return(".").Once()

		testFromAddress = os.Getenv("MAIL_FROM_ADDRESS")
		testFromName = os.Getenv("MAIL_FROM_NAME")
		testBcc = os.Getenv("MAIL_BCC")
		testCc = os.Getenv("MAIL_CC")
		testTo = os.Getenv("MAIL_TO")
	}

	return config
}

type TestMailable struct {
}

func NewTestMailable() *TestMailable {
	return &TestMailable{}
}

func (m *TestMailable) Attachments() []string {
	return []string{"../logo.png"}
}

func (m *TestMailable) Content() *mail.Content {
	html := Html("<h1>Hello Goravel</h1>")

	return &html
}

func (m *TestMailable) Envelope() *mail.Envelope {
	return &mail.Envelope{
		Bcc:     []string{testBcc},
		Cc:      []string{testCc},
		From:    Address(testFromAddress, testFromName),
		Subject: "Goravel Test Mailable",
		To:      []string{testTo},
	}
}

func (m *TestMailable) Headers() map[string]string {
	return map[string]string{
		"Test-Mailer": "TestMailable",
	}
}

func (m *TestMailable) Queue() *mail.Queue {
	return &mail.Queue{}
}

type stubMailable struct {
	attachments []string
	content     *mail.Content
	envelope    *mail.Envelope
	headers     map[string]string
	queue       *mail.Queue
}

func (s *stubMailable) Attachments() []string      { return s.attachments }
func (s *stubMailable) Content() *mail.Content     { return s.content }
func (s *stubMailable) Envelope() *mail.Envelope   { return s.envelope }
func (s *stubMailable) Headers() map[string]string { return s.headers }
func (s *stubMailable) Queue() *mail.Queue         { return s.queue }

func matchWithID(data any) bool {
	with, ok := data.(map[string]any)
	return ok && with["id"] == 1
}

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
		Content(mail.Content{Html: "<h1>Hello</h1>", View: "mail.tmpl", Text: "mail.txt", With: contentWith}).(*Application)

	// Builder methods should keep returning the same cloned instance for fluent chaining.
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
		content: &mail.Content{
			Html: "new-html",
			View: "email.tmpl",
			Text: "email.txt",
			With: map[string]any{"name": "goravel"},
		},
		envelope: &mail.Envelope{
			From:    mail.Address{Address: "from@example.com", Name: "Mailer"},
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
		template.EXPECT().Render("mail.tmpl", mock.MatchedBy(matchWithID)).Return("", errors.New("render failed")).Once()

		app := &Application{template: template, view: "mail.tmpl", with: map[string]any{"id": 1}}
		err := app.renderViewTemplate()
		assert.ErrorContains(t, err, "render failed")
	})

	t.Run("text render failed", func(t *testing.T) {
		template := mocksmail.NewTemplate(t)
		template.EXPECT().Render("mail.txt", mock.MatchedBy(matchWithID)).Return("", errors.New("text render failed")).Once()

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
		content:     &mail.Content{Html: "<h1>Queue</h1>"},
		envelope: &mail.Envelope{
			From:    mail.Address{Address: "from@example.com", Name: "From"},
			To:      []string{"to@example.com"},
			Cc:      []string{"cc@example.com"},
			Bcc:     []string{"bcc@example.com"},
			Subject: "queue-subject",
		},
		headers: map[string]string{"X-Test": "queue"},
		queue:   &mail.Queue{Connection: "redis", Queue: "emails"},
	}

	mockQueue.EXPECT().Job(
		mock.MatchedBy(func(job contractsqueue.Job) bool { return job != nil && job.Signature() == "goravel_send_mail_job" }),
		mock.MatchedBy(func(args []contractsqueue.Arg) bool { return len(args) == 10 }),
	).
		Run(func(job contractsqueue.Job, args ...[]contractsqueue.Arg) {
			assert.Equal(t, "goravel_send_mail_job", job.Signature())
			assert.Len(t, args, 1)
			assert.Len(t, args[0], 10)
			assert.Equal(t, "queue-subject", args[0][0].Value)
			assert.Equal(t, "<h1>Queue</h1>", args[0][1].Value)
			assert.Equal(t, "", args[0][2].Value)
			assert.Equal(t, "from@example.com", args[0][3].Value)
			assert.Equal(t, "From", args[0][4].Value)
			assert.Equal(t, []string{"to@example.com"}, args[0][5].Value)
			assert.Equal(t, []string{"cc@example.com"}, args[0][6].Value)
			assert.Equal(t, []string{"bcc@example.com"}, args[0][7].Value)
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
	template.EXPECT().Render("mail.tmpl", mock.MatchedBy(matchWithID)).Return("", errors.New("render failed")).Once()

	app := &Application{template: template, view: "mail.tmpl", with: map[string]any{"id": 1}}

	err := app.Queue()
	assert.ErrorContains(t, err, "render failed")
}

func TestApplicationSendRenderError(t *testing.T) {
	template := mocksmail.NewTemplate(t)
	template.EXPECT().Render("mail.tmpl", mock.MatchedBy(matchWithID)).Return("", errors.New("render failed")).Once()

	app := &Application{template: template}
	err := app.Send(&stubMailable{content: &mail.Content{View: "mail.tmpl", With: map[string]any{"id": 1}}})

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
		assert.ErrorContains(t, err, "not supported")
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
	assert.ErrorIs(t, err, os.ErrNotExist)
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

	app.EXPECT().Bind(
		binding.Mail,
		mock.AnythingOfType("func(foundation.Application) (interface {}, error)"),
	).
		Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			t.Run("without config", func(t *testing.T) {
				withoutConfig := mocksfoundation.NewApplication(t)
				withoutConfig.EXPECT().MakeConfig().Return(nil).Once()

				instance, err := callback(withoutConfig)
				assert.Nil(t, instance)
				assert.Error(t, err)
			})

			t.Run("without queue", func(t *testing.T) {
				withoutQueue := mocksfoundation.NewApplication(t)
				configOnly := mocksconfig.NewConfig(t)
				withoutQueue.EXPECT().MakeConfig().Return(configOnly).Once()
				withoutQueue.EXPECT().MakeQueue().Return(nil).Once()

				instance, err := callback(withoutQueue)
				assert.Nil(t, instance)
				assert.Error(t, err)
			})

			t.Run("with config and queue", func(t *testing.T) {
				withAll := mocksfoundation.NewApplication(t)
				configAndQueue := mocksconfig.NewConfig(t)
				queue := mocksqueue.NewQueue(t)
				withAll.EXPECT().MakeConfig().Return(configAndQueue).Once()
				withAll.EXPECT().MakeQueue().Return(queue).Once()
				configAndQueue.EXPECT().GetString("mail.template.default", "html").Return("mail_service_provider").Once()
				configAndQueue.EXPECT().GetString("mail.template.engines.mail_service_provider.driver", "html").Return("html").Once()
				configAndQueue.EXPECT().GetString("mail.template.engines.mail_service_provider.path", "resources/views/mail").Return(".").Once()

				instance, err := callback(withAll)
				assert.NoError(t, err)
				assert.NotNil(t, instance)
			})
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

		app.EXPECT().Commands(mock.AnythingOfType("[]console.Command")).
			Run(func(commands []contractsconsole.Command) {
				assert.Len(t, commands, 1)
			}).
			Once()
		app.EXPECT().MakeQueue().Return(queue).Once()
		app.EXPECT().MakeConfig().Return(config).Once()
		queue.EXPECT().Register(mock.AnythingOfType("[]queue.Job")).
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
