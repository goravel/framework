package notification

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/notification"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/mail"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	"github.com/goravel/framework/notification/channels"
	"github.com/goravel/framework/notification/models"
	"github.com/goravel/framework/queue"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func (u User) NotificationParams() map[string]interface{} {
	return map[string]interface{}{
		"id":    u.ID,
		"email": u.Email,
	}
}

type RegisterSuccessNotification struct {
	Title   string
	Content string
}

// New
func NewRegisterSuccessNotification(title, content string) *RegisterSuccessNotification {
	return &RegisterSuccessNotification{
		Title:   title,
		Content: content,
	}
}

func (n RegisterSuccessNotification) Via(notifiable notification.Notifiable) []string {
	return []string{
		"mail",
	}
}
func (n RegisterSuccessNotification) ToMail(notifiable notification.Notifiable) map[string]string {
	return map[string]string{
		"subject": n.Title,
		"content": n.Content,
	}
}

type LoginSuccessNotification struct {
	Title   string
	Content string
}

func NewLoginSuccessNotification(title, content string) *LoginSuccessNotification {
	return &LoginSuccessNotification{
		Title:   title,
		Content: content,
	}
}

func (n LoginSuccessNotification) Via(notifiable notification.Notifiable) []string {
	return []string{
		"database",
	}
}
func (n LoginSuccessNotification) ToDatabase(notifiable notification.Notifiable) map[string]string {
	return map[string]string{
		"title":   n.Title,
		"content": n.Content,
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
		ID:    "1",
		Email: "657873584@qq.com",
		Name:  "test",
	}

	var registerSuccessNotification = NewRegisterSuccessNotification("Registration successful!", "Congratulations, your registration is successful!")

	RegisterChannel("mail", &channels.MailChannel{})

	users := []notification.Notifiable{user}
	err = app.SendNow(users, registerSuccessNotification)
	s.Nil(err)
}

func (s *ApplicationTestSuite) TestMailNotificationOnQueue() {
	s.mockConfig = mockConfig(465)

	queueFacade := mockQueueFacade(s.mockConfig)

	mailFacade, err := mail.NewApplication(s.mockConfig, nil)
	s.Nil(err)

	app, err := NewApplication(s.mockConfig, queueFacade, nil, mailFacade)
	s.Nil(err)

	var user = User{
		ID:    "1",
		Email: "657873584@qq.com",
		Name:  "test",
	}

	var registerSuccessNotification = NewRegisterSuccessNotification("Registration successful!", "Congratulations, your registration is successful!")

	RegisterChannel("mail", &channels.MailChannel{})

	users := []notification.Notifiable{user}
	err = app.Send(users, registerSuccessNotification)
	s.Nil(err)
}

func (s *ApplicationTestSuite) TestDatabaseNotification() {
	var user = User{
		ID:    "1",
		Email: "657873584@qq.com",
		Name:  "test",
	}
	var loginSuccessNotification = NewLoginSuccessNotification("Login success", "Congratulations, your login is successful!")

	s.mockConfig = mockConfig(465)

	mockDB := mocksdb.NewDB(s.T())
	s.mockConfig.EXPECT().GetString("DB_CONNECTION").Return("mysql").Once()
	mockQuery := mocksdb.NewQuery(s.T())
	mockDB.EXPECT().Table("notifications").Return(mockQuery).Once()

	var notificationModel models.Notification
	notificationModel.ID = uuid.New().String()
	notificationModel.Data = "{\"content\":\"Congratulations, your login is successful!\",\"title\":\"Login success\"}"
	notificationModel.NotifiableId = user.ID
	notificationModel.NotifiableType = str.Of(fmt.Sprintf("%T", user)).Replace("*", "").String()
	notificationModel.Type = fmt.Sprintf("%T", loginSuccessNotification)

	mockQuery.EXPECT().Insert(mock.MatchedBy(func(model *models.Notification) bool {
		return model.Data == "{\"content\":\"Congratulations, your login is successful!\",\"title\":\"Login success\"}" &&
			model.NotifiableId == user.ID &&
			model.NotifiableType == str.Of(fmt.Sprintf("%T", user)).Replace("*", "").String() &&
			model.Type == str.Of(fmt.Sprintf("%T", loginSuccessNotification)).Replace("*", "").String()
	})).Return(nil, nil).Once()

	app, err := NewApplication(s.mockConfig, nil, mockDB, nil)
	s.Nil(err)

	RegisterChannel("database", &channels.DatabaseChannel{})

	users := []notification.Notifiable{user}
	err = app.SendNow(users, loginSuccessNotification)
	s.Nil(err)
}

func (s *ApplicationTestSuite) TestDatabaseNotificationOnQueue() {
	var user = User{
		ID:    "1",
		Email: "657873584@qq.com",
		Name:  "test",
	}

	var loginSuccessNotification = NewLoginSuccessNotification("Login success", "Congratulations, your login is successful!")

	s.mockConfig = mockConfig(465)
	queueFacade := mockQueueFacade(s.mockConfig)

	mockDB := mocksdb.NewDB(s.T())
	s.mockConfig.EXPECT().GetString("DB_CONNECTION").Return("mysql").Once()
	mockQuery := mocksdb.NewQuery(s.T())
	mockDB.EXPECT().Table("notifications").Return(mockQuery).Once()

	var notificationModel models.Notification
	notificationModel.ID = uuid.New().String()
	notificationModel.Data = "{\"content\":\"Congratulations, your login is successful!\",\"title\":\"Login success\"}"
	notificationModel.NotifiableId = user.ID
	notificationModel.NotifiableType = str.Of(fmt.Sprintf("%T", user)).Replace("*", "").String()
	notificationModel.Type = fmt.Sprintf("%T", loginSuccessNotification)

	mockQuery.EXPECT().Insert(mock.MatchedBy(func(model *models.Notification) bool {
		return model.Data == "{\"content\":\"Congratulations, your login is successful!\",\"title\":\"Login success\"}" &&
			model.NotifiableId == user.ID &&
			model.NotifiableType == str.Of(fmt.Sprintf("%T", user)).Replace("*", "").String() &&
			model.Type == str.Of(fmt.Sprintf("%T", loginSuccessNotification)).Replace("*", "").String()
	})).Return(nil, nil).Once()

	app, err := NewApplication(s.mockConfig, queueFacade, mockDB, nil)
	s.Nil(err)

	RegisterChannel("database", &channels.DatabaseChannel{})

	users := []notification.Notifiable{user}
	err = app.Send(users, loginSuccessNotification)
	s.Nil(err)
}

func (s *ApplicationTestSuite) TestNotifiableSerialize() {
	var loginSuccessNotification = NewLoginSuccessNotification("Login success", "Congratulations, your login is successful!")
	// 创建数据缓冲区
	var buf bytes.Buffer
	// 创建编码器
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(loginSuccessNotification)
	s.Nil(err)
	var loginSuccessNotification2 LoginSuccessNotification
	decoder := gob.NewDecoder(&buf)
	err = decoder.Decode(&loginSuccessNotification2)
	s.Nil(err)
	s.Equal(loginSuccessNotification.Title, loginSuccessNotification2.Title)
	s.Equal(loginSuccessNotification.Content, loginSuccessNotification2.Content)
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
		NewSendNotificationJob(mockConfig, nil, nil),
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
