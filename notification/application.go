package notification

import (
	"github.com/goravel/framework/contracts/config"
	contractsqueuedb "github.com/goravel/framework/contracts/database/db"
	contractsmail "github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/contracts/notification"
	contractsqueue "github.com/goravel/framework/contracts/queue"
)

// Application provides a facade-backed entry point for sending notifications.
// It wires configuration, queue, database, and mail facades needed by channels.
type Application struct {
	config config.Config
	queue  contractsqueue.Queue
	db     contractsqueuedb.DB
	mail   contractsmail.Mail
}

// NewApplication constructs an Application for the notification module.
func NewApplication(config config.Config, queue contractsqueue.Queue, db contractsqueuedb.DB, mail contractsmail.Mail) (*Application, error) {
	return &Application{
		config: config,
		queue:  queue,
		db:     db,
		mail:   mail,
	}, nil
}

// Send enqueues a notification to be processed asynchronously.
func (r *Application) Send(notifiables []notification.Notifiable, notif notification.Notif) error {
	if err := (NewNotificationSender(r.db, r.mail, r.queue)).Send(notifiables, notif); err != nil {
		return err
	}
	return nil
}

// SendNow sends a notification immediately without queueing.
func (r *Application) SendNow(notifiables []notification.Notifiable, notif notification.Notif) error {
	if err := (NewNotificationSender(r.db, r.mail, nil)).SendNow(notifiables, notif); err != nil {
		return err
	}
	return nil
}
