package mail

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/mail"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/foundation/json"
	mocksconfig "github.com/goravel/framework/mocks/config"
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

	queueFacade := queue.NewApplication(queue.NewConfig(s.mockConfig), nil, queue.NewJobStorer(), json.New(), nil)
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

	queueFacade := queue.NewApplication(queue.NewConfig(s.mockConfig), nil, queue.NewJobStorer(), json.New(), nil)
	queueFacade.Register([]contractsqueue.Job{
		NewSendMailJob(s.mockConfig),
	})

	app, err := NewApplication(s.mockConfig, queueFacade)
	s.Nil(err)
	s.Nil(app.Queue(NewTestMailable()))
}

func mockConfig(mailPort int) *mocksconfig.Config {
	config := &mocksconfig.Config{}
	config.On("GetString", "app.name").Return("goravel")
	config.On("GetString", "queue.default").Return("sync")
	config.On("GetString", "queue.connections.sync.queue", "default").Return("default")
	config.On("GetString", "queue.connections.sync.driver").Return("sync")
	config.On("GetInt", "queue.connections.sync.concurrent", 1).Return(1)
	config.On("GetString", "queue.failed.database").Return("database")
	config.On("GetString", "queue.failed.table").Return("failed_jobs")

	if file.Exists(support.EnvFilePath) {
		vip := viper.New()
		vip.SetConfigName(support.EnvFilePath)
		vip.SetConfigType("env")
		vip.AddConfigPath(".")
		_ = vip.ReadInConfig()
		vip.SetEnvPrefix("goravel")
		vip.AutomaticEnv()

		config.On("GetString", "mail.host").Return(vip.Get("MAIL_HOST"))
		config.On("GetInt", "mail.port").Return(mailPort)
		config.On("GetString", "mail.from.address").Return(vip.Get("MAIL_FROM_ADDRESS"))
		config.On("GetString", "mail.from.name").Return(vip.Get("MAIL_FROM_NAME"))
		config.On("GetString", "mail.username").Return(vip.Get("MAIL_USERNAME"))
		config.On("GetString", "mail.password").Return(vip.Get("MAIL_PASSWORD"))
		config.On("GetString", "mail.to").Return(vip.Get("MAIL_TO"))
		config.On("GetString", "mail.cc").Return(vip.Get("MAIL_CC"))
		config.On("GetString", "mail.bcc").Return(vip.Get("MAIL_BCC"))
		config.EXPECT().GetString("mail.template.default", "html").Return("html").Once()
		config.EXPECT().GetString("mail.template.engines.html.driver", "html").Return("html").Once()
		config.EXPECT().GetString("mail.template.engines.html.path", "resources/views/mail").
			Return("resources/views/mail").Once()

		testFromAddress = vip.Get("MAIL_FROM_ADDRESS").(string)
		testFromName = vip.Get("MAIL_FROM_NAME").(string)
		testTo = vip.Get("MAIL_TO").(string)
	}
	if os.Getenv("MAIL_HOST") != "" {
		config.On("GetString", "mail.host").Return(os.Getenv("MAIL_HOST"))
		config.On("GetInt", "mail.port").Return(mailPort)
		config.On("GetString", "mail.from.address").Return(os.Getenv("MAIL_FROM_ADDRESS"))
		config.On("GetString", "mail.from.name").Return(os.Getenv("MAIL_FROM_NAME"))
		config.On("GetString", "mail.username").Return(os.Getenv("MAIL_USERNAME"))
		config.On("GetString", "mail.password").Return(os.Getenv("MAIL_PASSWORD"))
		config.On("GetString", "mail.to").Return(os.Getenv("MAIL_TO"))
		config.On("GetString", "mail.cc").Return(os.Getenv("MAIL_CC"))
		config.On("GetString", "mail.bcc").Return(os.Getenv("MAIL_BCC"))
		config.EXPECT().GetString("mail.template.default", "html").Return("html").Once()
		config.EXPECT().GetString("mail.template.engines.html.driver", "html").Return("html").Once()
		config.EXPECT().GetString("mail.template.engines.html.path", "resources/views/mail").
			Return("resources/views/mail").Once()

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
