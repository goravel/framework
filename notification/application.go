package notification

import (
	"fmt"
	"github.com/goravel/framework/contracts/config"
	contractsqueuedb "github.com/goravel/framework/contracts/database/db"
	contractsmail "github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/contracts/notification"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/notification/channels"
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
func (r *Application) Send(notifiable notification.Notifiable, notif notification.Notif) error {
	vias := notif.Via(notifiable)
	if len(vias) == 0 {
		return errors.New("no channels defined for notification")
	}

	for _, chName := range vias {
		ch, ok := GetChannel(chName)
		if !ok {
			return fmt.Errorf("channel not registered: %s", chName)
		}
		if chName == "database" {
			if databaseChannel, ok := ch.(*channels.DatabaseChannel); ok {
				databaseChannel.SetDB(r.db)
			}
		} else if chName == "mail" {
			if mailChannel, ok := ch.(*channels.MailChannel); ok {
				mailChannel.SetMail(r.mail)
			}
		}
		if err := ch.Send(notifiable, notif); err != nil {
			return fmt.Errorf("channel %s send error: %w", chName, err)
		}
	}
	return nil
}
