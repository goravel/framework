package channels

import (
    "fmt"
    "github.com/google/uuid"
    contractsqueuedb "github.com/goravel/framework/contracts/database/db"
    "github.com/goravel/framework/contracts/notification"
    "github.com/goravel/framework/notification/models"
    "github.com/goravel/framework/notification/utils"
    "github.com/goravel/framework/support/json"
    "github.com/goravel/framework/support/str"
)

// DatabaseChannel is the default database persistence channel.
type DatabaseChannel struct {
    db contractsqueuedb.DB
}

// Send persists the notification payload to the notifications table.
// It expects the notification to implement a ToDatabase(notifiable) method or PayloadProvider.
func (c *DatabaseChannel) Send(notifiable notification.Notifiable, notif interface{}) error {
    data, err := utils.CallToMethod(notif, "ToDatabase", notifiable)
    if err != nil {
        return fmt.Errorf("[DatabaseChannel] %s", err.Error())
    }

	var notificationModel models.Notification
	notificationModel.ID = uuid.New().String()
	notificationModel.NotifiableId = notifiable.NotificationParams()["id"].(string)

	notificationModel.NotifiableType = str.Of(fmt.Sprintf("%T", notifiable)).Replace("*", "").String()
	notificationModel.Type = str.Of(fmt.Sprintf("%T", notif)).Replace("*", "").String()

	jsonData, _ := json.MarshalString(data)
	notificationModel.Data = jsonData

    if _, err = c.db.Table("notifications").Insert(&notificationModel); err != nil {
        return err
    }
    return nil
}

// SetDB injects the database facade into the channel.
func (c *DatabaseChannel) SetDB(db contractsqueuedb.DB) {
    c.db = db
}
