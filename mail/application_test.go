package mail

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/mail"
	queuecontract "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/queue"
	"github.com/goravel/framework/support/file"
	testingdocker "github.com/goravel/framework/testing/docker"
	"github.com/goravel/framework/testing/mock"
)

type ApplicationTestSuite struct {
	suite.Suite
	redisPort int
}

func TestApplicationTestSuite(t *testing.T) {
	if !file.Exists("../.env") && os.Getenv("MAIL_HOST") == "" {
		color.Redln("No mail tests run, need create .env based on .env.example, then initialize it")
		return
	}

	redisPool, redisResource, err := testingdocker.Redis()
	assert.Nil(t, err)

	facades.Mail = NewApplication()
	suite.Run(t, &ApplicationTestSuite{
		redisPort: cast.ToInt(redisResource.GetPort("6379/tcp")),
	})

	assert.Nil(t, redisPool.Purge(redisResource))
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestSendMailBy25Port() {
	// sendinblue doesn't support 25 port
	if !file.Exists("../.env") {
		return
	}
	initConfig(25, s.redisPort)
	s.Nil(facades.Mail.To([]string{facades.Config.GetString("mail.to")}).
		Cc([]string{facades.Config.GetString("mail.cc")}).
		Bcc([]string{facades.Config.GetString("mail.bcc")}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test 25", Html: "<h1>Hello Goravel</h1>"}).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailBy465Port() {
	initConfig(465, s.redisPort)
	s.Nil(facades.Mail.To([]string{facades.Config.GetString("mail.to")}).
		Cc([]string{facades.Config.GetString("mail.cc")}).
		Bcc([]string{facades.Config.GetString("mail.bcc")}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test 465", Html: "<h1>Hello Goravel</h1>"}).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailBy587Port() {
	initConfig(587, s.redisPort)
	s.Nil(facades.Mail.To([]string{facades.Config.GetString("mail.to")}).
		Cc([]string{facades.Config.GetString("mail.cc")}).
		Bcc([]string{facades.Config.GetString("mail.bcc")}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test 587", Html: "<h1>Hello Goravel</h1>"}).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailWithFrom() {
	initConfig(587, s.redisPort)
	s.Nil(facades.Mail.From(mail.From{Address: facades.Config.GetString("mail.from.address"), Name: facades.Config.GetString("mail.from.name")}).
		To([]string{facades.Config.GetString("mail.to")}).
		Cc([]string{facades.Config.GetString("mail.cc")}).
		Bcc([]string{facades.Config.GetString("mail.bcc")}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test 587 With From", Html: "<h1>Hello Goravel</h1>"}).
		Send())
}

func (s *ApplicationTestSuite) TestQueueMail() {
	initConfig(587, s.redisPort)
	facades.Queue = queue.NewApplication()
	facades.Queue.Register([]queuecontract.Job{
		&SendMailJob{},
	})

	mockEvent, _ := mock.Event()
	mockEvent.On("GetEvents").Return(map[event.Event][]event.Listener{}).Once()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(facades.Queue.Worker(nil).Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(3 * time.Second)
	s.Nil(facades.Mail.To([]string{facades.Config.GetString("mail.to")}).
		Cc([]string{facades.Config.GetString("mail.cc")}).
		Bcc([]string{facades.Config.GetString("mail.bcc")}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test Queue", Html: "<h1>Hello Goravel</h1>"}).
		Queue(nil))
	time.Sleep(1 * time.Second)

	mockEvent.AssertExpectations(s.T())
}

func initConfig(mailPort, redisPort int) {
	mockConfig := mock.Config()
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
		application := config.NewApplication("../.env")

		mockConfig.On("GetString", "mail.host").Return(application.Env("MAIL_HOST", ""))
		mockConfig.On("GetInt", "mail.port").Return(mailPort)
		mockConfig.On("GetString", "mail.from.address").Return(application.Env("MAIL_FROM_ADDRESS", "hello@example.com"))
		mockConfig.On("GetString", "mail.from.name").Return(application.Env("MAIL_FROM_NAME", "Example"))
		mockConfig.On("GetString", "mail.username").Return(application.Env("MAIL_USERNAME"))
		mockConfig.On("GetString", "mail.password").Return(application.Env("MAIL_PASSWORD"))
		mockConfig.On("GetString", "mail.to").Return(application.Env("MAIL_TO"))
		mockConfig.On("GetString", "mail.cc").Return(application.Env("MAIL_CC"))
		mockConfig.On("GetString", "mail.bcc").Return(application.Env("MAIL_BCC"))
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
	}
}
