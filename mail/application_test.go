package mail

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/gookit/color"
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
}

func TestApplicationTestSuite(t *testing.T) {
	if !file.Exists("../.env") {
		color.Redln("No mail tests run, need create .env based on .env.example, then initialize it")
		return
	}

	redisPool, redisResource, err := testingdocker.Redis()
	if err != nil {
		log.Fatalf("Get redis error: %s", err)
	}

	initConfig(redisResource.GetPort("6379/tcp"))
	facades.Mail = NewApplication()
	suite.Run(t, new(ApplicationTestSuite))

	if err := redisPool.Purge(redisResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestSendMail() {
	s.Nil(facades.Mail.To([]string{facades.Config.Env("MAIL_TO").(string)}).
		Cc([]string{facades.Config.Env("MAIL_CC").(string)}).
		Bcc([]string{facades.Config.Env("MAIL_BCC").(string)}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test", Html: "<h1>Hello Goravel</h1>"}).
		Send())
}

func (s *ApplicationTestSuite) TestSendMailWithFrom() {
	s.Nil(facades.Mail.From(mail.From{Address: facades.Config.GetString("mail.from.address"), Name: facades.Config.GetString("mail.from.name")}).
		To([]string{facades.Config.Env("MAIL_TO").(string)}).
		Cc([]string{facades.Config.Env("MAIL_CC").(string)}).
		Bcc([]string{facades.Config.Env("MAIL_BCC").(string)}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test With From", Html: "<h1>Hello Goravel</h1>"}).
		Send())
}

func (s *ApplicationTestSuite) TestQueueMail() {
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

		for {
			select {
			case <-ctx.Done():
				return
			}
		}
	}(ctx)
	time.Sleep(3 * time.Second)
	s.Nil(facades.Mail.To([]string{facades.Config.Env("MAIL_TO").(string)}).
		Cc([]string{facades.Config.Env("MAIL_CC").(string)}).
		Bcc([]string{facades.Config.Env("MAIL_BCC").(string)}).
		Attach([]string{"../logo.png"}).
		Content(mail.Content{Subject: "Goravel Test Queue", Html: "<h1>Hello Goravel</h1>"}).
		Queue(nil))
	time.Sleep(1 * time.Second)

	mockEvent.AssertExpectations(s.T())
}

func initConfig(redisPort string) {
	application := config.NewApplication("../.env")
	application.Add("app", map[string]interface{}{
		"name": "goravel",
	})
	application.Add("mail", map[string]any{
		"host": application.Env("MAIL_HOST", ""),
		"port": application.Env("MAIL_PORT", 587),
		"from": map[string]interface{}{
			"address": application.Env("MAIL_FROM_ADDRESS", "hello@example.com"),
			"name":    application.Env("MAIL_FROM_NAME", "Example"),
		},
		"username": application.Env("MAIL_USERNAME"),
		"password": application.Env("MAIL_PASSWORD"),
	})
	application.Add("queue", map[string]interface{}{
		"default": "redis",
		"connections": map[string]interface{}{
			"sync": map[string]interface{}{
				"driver": "sync",
			},
			"redis": map[string]interface{}{
				"driver":     "redis",
				"connection": "default",
				"queue":      "default",
			},
		},
	})
	application.Add("database", map[string]interface{}{
		"redis": map[string]interface{}{
			"default": map[string]interface{}{
				"host":     "localhost",
				"password": "",
				"port":     redisPort,
				"database": 0,
			},
		},
	})

	facades.Config = application
}
