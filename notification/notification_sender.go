package notification

import (
	"bytes"
	"encoding/gob"
	"fmt"
	contractsqueuedb "github.com/goravel/framework/contracts/database/db"
	contractsmail "github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/contracts/notification"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/notification/channels"
	"github.com/goravel/framework/notification/utils"
	"github.com/goravel/framework/support/json"
	"github.com/goravel/framework/support/str"
	"strings"
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
	// 创建数据缓冲区
	var buf bytes.Buffer

	// 创建编码器
	encoder := gob.NewEncoder(&buf)
	for _, notifiable := range notifiables {

		notifiableSerialize := utils.Serialize(notifiable)

		pendingJob := s.queue.Job(NewSendNotificationJob(nil, s.db, s.mail), []contractsqueue.Arg{
			{Type: "[]string", Value: vias},
			{Type: "string", Value: routesJSON},
			{Type: "string", Value: payloadsJSON},
		})
		if err := pendingJob.Dispatch(); err != nil {
			return err
		}
	}
	return nil
}
