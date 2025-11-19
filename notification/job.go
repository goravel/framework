package notification

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/goravel/framework/contracts/config"
	contractsqueuedb "github.com/goravel/framework/contracts/database/db"
	contractsmail "github.com/goravel/framework/contracts/mail"
	contractsnotification "github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/notification/channels"
)

type SendNotificationJob struct {
	config config.Config
	db     contractsqueuedb.DB
	mail   contractsmail.Mail
}

type GobEnvelope struct {
	Notifiable any
	Notif      any
}

func NewSendNotificationJob(config config.Config, db contractsqueuedb.DB, mail contractsmail.Mail) *SendNotificationJob {
	return &SendNotificationJob{
		config: config,
		db:     db,
		mail:   mail,
	}
}

// Signature The name and signature of the job.
func (r *SendNotificationJob) Signature() string {
	return "goravel_send_notification_job"
}

// Handle Execute the job.
func (r *SendNotificationJob) Handle(args ...any) error {
	if len(args) != 3 {
		return fmt.Errorf("expected 3 arguments, got %d", len(args))
	}

	notifiableBytes, _ := args[0].([]uint8)
	notifBytes, _ := args[1].([]uint8)
	vias, ok := args[2].([]string)
	if !ok {
		return fmt.Errorf("invalid channels payload type: %T", args[2])
	}

	var notifiable any
	var notif any
	// Try envelope decode first
	if len(notifiableBytes) > 0 && len(notifBytes) == 0 {
		var env GobEnvelope
		if err := gob.NewDecoder(bytes.NewReader(notifiableBytes)).Decode(&env); err != nil {
			return err
		}
		notifiable = env.Notifiable
		notif = env.Notif
	} else {
		dec1 := gob.NewDecoder(bytes.NewReader(notifiableBytes))
		if err := dec1.Decode(&notifiable); err != nil {
			return err
		}

		dec2 := gob.NewDecoder(bytes.NewReader(notifBytes))
		if err := dec2.Decode(&notif); err != nil {
			return err
		}
	}

	nbl, ok := notifiable.(contractsnotification.Notifiable)
	if !ok {
		return fmt.Errorf("decoded notifiable does not implement Notifiable: %T", notifiable)
	}
	nf, ok := notif.(contractsnotification.Notif)
	if !ok {
		return fmt.Errorf("decoded notif does not implement Notif: %T", notif)
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
		if err := ch.Send(nbl, nf); err != nil {
			return fmt.Errorf("channel %s send error: %w", chName, err)
		}
	}

	return nil
}
