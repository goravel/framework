package notification

import (
	"github.com/goravel/framework/contracts/config"
	contractsqueuedb "github.com/goravel/framework/contracts/database/db"
	contractsmail "github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/contracts/notification"
	contractsqueue "github.com/goravel/framework/contracts/queue"
)

type Application struct {
	config config.Config
	queue  contractsqueue.Queue
	db     contractsqueuedb.DB
	mail   contractsmail.Mail
}

func NewApplication(config config.Config, queue contractsqueue.Queue, db contractsqueuedb.DB, mail contractsmail.Mail) (*Application, error) {
	return &Application{
		config: config,
		queue:  queue,
		db:     db,
		mail:   mail,
	}, nil
}

// Send a notification.
func (r *Application) Send(notifiables []notification.Notifiable, notif notification.Notif) error {
	if err := (NewNotificationSender(r.db, r.mail, r.queue)).Send(notifiables, notif); err != nil {
		return err
	}
	return nil
}

func (r *Application) SendNow(notifiables []notification.Notifiable, notif notification.Notif) error {
	if err := (NewNotificationSender(r.db, r.mail, nil)).SendNow(notifiables, notif); err != nil {
		return err
	}
	return nil
}
