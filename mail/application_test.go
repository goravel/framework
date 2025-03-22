package mail

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/mail"
	queuecontract "github.com/goravel/framework/contracts/queue"
	configmock "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/queue"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

var testBcc, testCc, testTo, testFromAddress, testFromName string

type ApplicationTestSuite struct {
	suite.Suite
	mockConfig *configmock.Config
}

func TestApplicationTestSuite(t *testing.T) {
	if !file.Exists(support.EnvFilePath) && os.Getenv("MAIL_HOST") == "" {
		color.Errorln("No mail tests run, need create .env based on .env.example, then initialize it")
		return
	}

	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {
	s.mockConfig = mockConfig(465)
}

func (s *ApplicationTestSuite) TestSendMailBy465Port() {
	app := NewApplication(s.mockConfig, nil)
	s.Nil(app.To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Subject("Goravel Test 465").
		Content(Html("<h1>Hello Goravel</h1>")).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailBy587Port() {
	app := NewApplication(s.mockConfig, nil)
	s.Nil(app.To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Subject("Goravel Test 587").
		Content(Html("<h1>Hello Goravel</h1>")).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailWithFrom() {
	app := NewApplication(s.mockConfig, nil)
	s.Nil(app.From(Address(testFromAddress, testFromName)).
		To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Subject("Goravel Test 587 With From").
		Content(Html("<h1>Hello Goravel</h1>")).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailWithMailable() {
	app := NewApplication(s.mockConfig, nil)
	s.Nil(app.Send(NewTestMailable()))
}

func (s *ApplicationTestSuite) TestQueueMail() {
	queueFacade := queue.NewApplication(s.mockConfig)
	queueFacade.Register([]queuecontract.Job{
		NewSendMailJob(s.mockConfig),
	})

	app := NewApplication(s.mockConfig, queueFacade)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(queueFacade.Worker().Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(3 * time.Second)
	s.Nil(app.To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Subject("Goravel Test Queue").
		Content(Html("<h1>Hello Goravel</h1>")).
		Queue())
	time.Sleep(3 * time.Second)
}

func (s *ApplicationTestSuite) TestQueueMailWithConnection() {
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Twice()
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("async").Twice()
	s.mockConfig.On("GetInt", "queue.connections.redis.size", 100).Return(100).Twice()

	queueFacade := queue.NewApplication(s.mockConfig)
	queueFacade.Register([]queuecontract.Job{
		NewSendMailJob(s.mockConfig),
	})

	app := NewApplication(s.mockConfig, queueFacade)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(queueFacade.Worker(queuecontract.Args{
			Connection: "redis",
			Queue:      "test",
		}).Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(3 * time.Second)
	s.Nil(app.To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Subject("Goravel Test Queue with connection").
		Content(Html("<h1>Hello Goravel</h1>")).
		Queue(Queue().OnConnection("redis").OnQueue("test")))
	time.Sleep(3 * time.Second)
}

func (s *ApplicationTestSuite) TestQueueMailWithMailable() {
	queueFacade := queue.NewApplication(s.mockConfig)
	queueFacade.Register([]queuecontract.Job{
		NewSendMailJob(s.mockConfig),
	})

	app := NewApplication(s.mockConfig, queueFacade)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(queueFacade.Worker().Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(3 * time.Second)
	s.Nil(app.Queue(NewTestMailable()))
	time.Sleep(3 * time.Second)
}

func mockConfig(mailPort int) *configmock.Config {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "app.name").Return("goravel")
	mockConfig.On("GetString", "queue.default").Return("async")
	mockConfig.On("GetString", "queue.connections.async.queue", "default").Return("default")
	mockConfig.On("GetString", "queue.connections.async.driver").Return("async")
	mockConfig.On("GetInt", "queue.connections.async.size", 100).Return(100)
	mockConfig.On("GetString", "queue.failed.database").Return("database")
	mockConfig.On("GetString", "queue.failed.table").Return("failed_jobs")

	if file.Exists(support.EnvFilePath) {
		vip := viper.New()
		vip.SetConfigName(support.EnvFilePath)
		vip.SetConfigType("env")
		vip.AddConfigPath(".")
		_ = vip.ReadInConfig()
		vip.SetEnvPrefix("goravel")
		vip.AutomaticEnv()

		mockConfig.On("GetString", "mail.host").Return(vip.Get("MAIL_HOST"))
		mockConfig.On("GetInt", "mail.port").Return(mailPort)
		mockConfig.On("GetString", "mail.from.address").Return(vip.Get("MAIL_FROM_ADDRESS"))
		mockConfig.On("GetString", "mail.from.name").Return(vip.Get("MAIL_FROM_NAME"))
		mockConfig.On("GetString", "mail.username").Return(vip.Get("MAIL_USERNAME"))
		mockConfig.On("GetString", "mail.password").Return(vip.Get("MAIL_PASSWORD"))
		mockConfig.On("GetString", "mail.to").Return(vip.Get("MAIL_TO"))
		mockConfig.On("GetString", "mail.cc").Return(vip.Get("MAIL_CC"))
		mockConfig.On("GetString", "mail.bcc").Return(vip.Get("MAIL_BCC"))

		testFromAddress = vip.Get("MAIL_FROM_ADDRESS").(string)
		testFromName = vip.Get("MAIL_FROM_NAME").(string)
		testTo = vip.Get("MAIL_TO").(string)
	}
	if os.Getenv("MAIL_HOST") != "" {
		mockConfig.On("GetString", "mail.host").Return(os.Getenv("MAIL_HOST"))
		mockConfig.On("GetInt", "mail.port").Return(mailPort)
		mockConfig.On("GetString", "mail.from.address").Return(os.Getenv("MAIL_FROM_ADDRESS"))
		mockConfig.On("GetString", "mail.from.name").Return(os.Getenv("MAIL_FROM_NAME"))
		mockConfig.On("GetString", "mail.username").Return(os.Getenv("MAIL_USERNAME"))
		mockConfig.On("GetString", "mail.password").Return(os.Getenv("MAIL_PASSWORD"))
		mockConfig.On("GetString", "mail.to").Return(os.Getenv("MAIL_TO"))
		mockConfig.On("GetString", "mail.cc").Return(os.Getenv("MAIL_CC"))
		mockConfig.On("GetString", "mail.bcc").Return(os.Getenv("MAIL_BCC"))

		testFromAddress = os.Getenv("MAIL_FROM_ADDRESS")
		testFromName = os.Getenv("MAIL_FROM_NAME")
		testBcc = os.Getenv("MAIL_BCC")
		testCc = os.Getenv("MAIL_CC")
		testTo = os.Getenv("MAIL_TO")
	}

	return mockConfig
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
		Subject: "Goravel Test 587 With Mailable",
		To:      []string{testTo},
	}
}

func (m *TestMailable) Queue() *mail.Queue {
	return &mail.Queue{}
}
