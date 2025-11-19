package notification

import (
	"fmt"
	"github.com/goravel/framework/contracts/config"
	contractsqueuedb "github.com/goravel/framework/contracts/database/db"
	contractsmail "github.com/goravel/framework/contracts/mail"
	contractsnotification "github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/notification/channels"
	"github.com/goravel/framework/support/json"
)

type SendNotificationJob struct {
	config config.Config
	db     contractsqueuedb.DB
	mail   contractsmail.Mail
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

	channelsArg, ok := args[0].([]string)
	if !ok {
		return fmt.Errorf("channels should be of type []string")
	}
	routesJSON, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("routes should be of type string")
	}
	payloadsJSON, ok := args[2].(string)
	if !ok {
		return fmt.Errorf("payloads should be of type string")
	}

	var routes map[string]any
	if routesJSON != "" {
		if err := json.UnmarshalString(routesJSON, &routes); err != nil {
			return err
		}
	}
	var payloads map[string]map[string]interface{}
	if payloadsJSON != "" {
		if err := json.UnmarshalString(payloadsJSON, &payloads); err != nil {
			return err
		}
	}

	var notifiable contractsnotification.Notifiable = MapNotifiable{Routes: routes}
	if nt, ok := routes["_notifiable_type"].(string); ok && nt != "" && NotifiableHasWithRoutes(nt) {
		if inst, ok := GetNotifiableInstance(nt, routes); ok {
			notifiable = inst
		}
	}

	payloadNotif := PayloadNotification{Channels: channelsArg, Payloads: payloads}
	notifObj := any(payloadNotif)
	if t, ok := routes["_notif_type"].(string); ok && t != "" {
		if inst, ok := GetNotificationInstance(t); ok {
			notifObj = inst
		}
	}

	for _, chName := range channelsArg {
		ch, ok := GetChannel(chName)
		if !ok {
			return fmt.Errorf("channel not registered: %s", chName)
		}
		if chName == "database" && r.db != nil {
			if databaseChannel, ok := ch.(*channels.DatabaseChannel); ok {
				databaseChannel.SetDB(r.db)
			}
		} else if chName == "mail" && r.mail != nil {
			if mailChannel, ok := ch.(*channels.MailChannel); ok {
				mailChannel.SetMail(r.mail)
			}
		}
		if err := ch.Send(notifiable, notifObj); err != nil {
			return fmt.Errorf("channel %s send error: %w", chName, err)
		}
	}

	return nil
}
