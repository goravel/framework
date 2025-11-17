package notification

import (
	"fmt"
	contractsqueuedb "github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/notification"
	contractsqueue "github.com/goravel/framework/contracts/queue"

	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/mail"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	"github.com/goravel/framework/notification/channels"
	"github.com/goravel/framework/queue"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func (u User) RouteNotificationFor(channel string) any {
	switch channel {
	case "mail":
		return u.Email
	case "database":
		return u.ID
	default:
		return ""
	}
}

type RegisterSuccessNotification struct {
}

func (n RegisterSuccessNotification) Via(notifiable notification.Notifiable) []string {
	return []string{
		"mail",
	}
}
func (n RegisterSuccessNotification) ToMail(notifiable notification.Notifiable) map[string]string {
	return map[string]string{
		"subject": "【sign】Register success",
		"content": "Congratulations, your registration is successful!",
	}
}

type LoginSuccessNotification struct {
}

func (n LoginSuccessNotification) Via(notifiable notification.Notifiable) []string {
	return []string{
		"database",
	}
}
func (n LoginSuccessNotification) ToDatabase(notifiable notification.Notifiable) map[string]string {
	return map[string]string{
		"title":   "Login success",
		"content": "Congratulations, your login is successful!",
	}
}

type ApplicationTestSuite struct {
	suite.Suite
	mockConfig *mocksconfig.Config
}

func TestApplicationTestSuite(t *testing.T) {
	if !file.Exists(support.EnvFilePath) && os.Getenv("MAIL_HOST") == "" {
		color.Errorln("No notification tests run, need create .env based on .env.example, then initialize it")
		return
	}

	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {
}

func (s *ApplicationTestSuite) TestMailNotification() {
	s.mockConfig = mockConfig(465)

	queueFacade := mockQueueFacade(s.mockConfig)

	mailFacade, err := mail.NewApplication(s.mockConfig, nil)
	s.Nil(err)

	app, err := NewApplication(s.mockConfig, queueFacade, nil, mailFacade)
	s.Nil(err)

	var user = User{
		ID:    1,
		Email: "657873584@qq.com",
		Name:  "test",
	}

	var registerSuccessNotification = RegisterSuccessNotification{}

	RegisterChannel("mail", &channels.MailChannel{})

	err = app.Send(user, registerSuccessNotification)
	s.Nil(err)
}

func (s *ApplicationTestSuite) TestDatabaseNotification() {
	s.mockConfig = mockConfig(465)

	var mockDB contractsqueuedb.DB
	dbFacade := mocksdb.NewDB(s.T())
	dbFacade.EXPECT().Connection("mysql").Return(mockDB).Once()
	dbFacade.EXPECT().Table("notifications").Return(nil).Once()

	fmt.Println(mockDB)

	app, err := NewApplication(s.mockConfig, nil, mockDB, nil)
	s.Nil(err)

	var user = User{
		ID:    1,
		Email: "657873584@qq.com",
		Name:  "test",
	}

	var loginSuccessNotification = LoginSuccessNotification{}

	RegisterChannel("database", &channels.DatabaseChannel{})

	err = app.Send(user, loginSuccessNotification)
	s.Nil(err)
}

func mockQueueFacade(mockConfig *mocksconfig.Config) contractsqueue.Queue {
	mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return("default").Once()
	mockConfig.EXPECT().GetInt("queue.connections.redis.concurrent", 1).Return(2).Once()
	mockConfig.EXPECT().GetString("app.name", "goravel").Return("goravel").Once()
	mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	mockConfig.EXPECT().GetString("queue.failed.database").Return("mysql").Once()
	mockConfig.EXPECT().GetString("queue.failed.table").Return("failed_jobs").Once()

	queueFacade := queue.NewApplication(queue.NewConfig(mockConfig), nil, queue.NewJobStorer(), json.New(), nil)
	queueFacade.Register([]contractsqueue.Job{
		NewSendNotificationJob(mockConfig),
	})
	return queueFacade
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
	}

	return config
}
