package mail

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/mail"
	queuecontract "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/queue"
	testingdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/file"
)

var testBcc, testCc, testTo, testFromAddress, testFromName string

type ApplicationTestSuite struct {
	suite.Suite
	redisPort int
}

func TestApplicationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}

	if !file.Exists("../.env") && os.Getenv("MAIL_HOST") == "" {
		color.Redln("No mail tests run, need create .env based on .env.example, then initialize it")
		return
	}

	redisPool, redisResource, err := testingdocker.Redis()
	assert.Nil(t, err)

	suite.Run(t, &ApplicationTestSuite{
		redisPort: cast.ToInt(redisResource.GetPort("6379/tcp")),
	})

	assert.Nil(t, redisPool.Purge(redisResource))
}

func (s *ApplicationTestSuite) SetupTest() {}

func (s *ApplicationTestSuite) TestSendMailBy465Port() {
	mockConfig := mockConfig(465, s.redisPort)
	app := NewApplication(mockConfig, nil)
	s.Nil(app.To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test 465", Html: "<h1>Hello Goravel</h1>"}).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailBy587Port() {
	mockConfig := mockConfig(587, s.redisPort)
	app := NewApplication(mockConfig, nil)
	s.Nil(app.To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test 587", Html: "<h1>Hello Goravel</h1>"}).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailWithFrom() {
	mockConfig := mockConfig(587, s.redisPort)
	app := NewApplication(mockConfig, nil)
	s.Nil(app.From(mail.From{Address: testFromAddress, Name: testFromName}).
		To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test 587 With From", Html: "<h1>Hello Goravel</h1>"}).
		Send())
}

func (s *ApplicationTestSuite) TestQueueMail() {
	mockConfig := mockConfig(587, s.redisPort)

	queueFacade := queue.NewApplication(mockConfig)
	queueFacade.Register([]queuecontract.Job{
		NewSendMailJob(mockConfig),
	})

	app := NewApplication(mockConfig, queueFacade)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(queueFacade.Worker(nil).Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(3 * time.Second)
	s.Nil(app.To([]string{testTo}).
		Cc([]string{testCc}).
		Bcc([]string{testBcc}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test Queue", Html: "<h1>Hello Goravel</h1>"}).
		Queue(nil))
	time.Sleep(3 * time.Second)
}

func mockConfig(mailPort, redisPort int) *configmock.Config {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "app.name").Return("goravel")
	mockConfig.On("GetString", "queue.default").Return("redis")
	mockConfig.On("GetString", "queue.connections.sync.driver").Return("sync")
	mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis")
	mockConfig.On("GetString", "queue.connections.redis.connection").Return("default")
	mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default")
	mockConfig.On("GetString", "database.redis.default.host").Return("localhost")
	mockConfig.On("GetString", "database.redis.default.password").Return("")
	mockConfig.On("GetInt", "database.redis.default.port").Return(redisPort)
	mockConfig.On("GetInt", "database.redis.default.database").Return(0)

	if file.Exists("../.env") {
		vip := viper.New()
		vip.SetConfigName("../.env")
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
