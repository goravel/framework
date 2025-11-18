package notification

import (
	"fmt"
	contractsqueuedb "github.com/goravel/framework/contracts/database/db"
	contractsmail "github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/contracts/notification"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/notification/channels"
	"github.com/goravel/framework/support/json"
)

type NotificationSender struct {
	db    contractsqueuedb.DB
	mail  contractsmail.Mail
	queue contractsqueue.Queue
}

func NewNotificationSender(db contractsqueuedb.DB, mail contractsmail.Mail, queue contractsqueue.Queue) *NotificationSender {
	return &NotificationSender{
		db:    db,
		mail:  mail,
		queue: queue,
	}
}

// Send(notifiables []Notifiable, notification Notif) error
func (s *NotificationSender) Send(notifiables []notification.Notifiable, notification notification.Notif) error {
	if err := s.queueNotification(notifiables, notification); err != nil {
		return err
	}
	return nil
}

func (s *NotificationSender) SendNow(notifiables []notification.Notifiable, notif notification.Notif) error {
	for _, notifiable := range notifiables {
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
					databaseChannel.SetDB(s.db)
				}
			} else if chName == "mail" {
				if mailChannel, ok := ch.(*channels.MailChannel); ok {
					mailChannel.SetMail(s.mail)
				}
			}
			if err := ch.Send(notifiable, notif); err != nil {
				return fmt.Errorf("channel %s send error: %w", chName, err)
			}
		}
	}
	return nil
}

// queueNotification
func (s *NotificationSender) queueNotification(notifiables []notification.Notifiable, notif notification.Notif) error {
	var err error
	var notifiablesJson string
	if notifiablesJson, err = json.MarshalString(notifiables); err != nil {
		return err
	}
	var notifJson string
	if notifJson, err = json.MarshalString(notif); err != nil {
		return err
	}
	pendingJob := s.queue.Job(&SendNotificationJob{}, []contractsqueue.Arg{
		{
			Type:  "string",
			Value: notifiablesJson,
		},
		{
			Type:  "string",
			Value: notifJson,
		},
	})
	return pendingJob.Dispatch()
}
