package channels

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	contractsnotification "github.com/goravel/framework/contracts/notification"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

type DatabaseNotificationModel struct {
	ID             string     `gorm:"primaryKey;type:varchar(36);column:id"`
	Type           string     `gorm:"not null;column:type"`
	NotifiableType string     `gorm:"not null;column:notifiable_type"`
	NotifiableID   string     `gorm:"not null;column:notifiable_id"`
	Data           string     `gorm:"type:text;not null;column:data"`
	ReadAt         *time.Time `gorm:"column:read_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (DatabaseNotificationModel) TableName() string { return "notifications" }

type DatabaseChannel struct {
	orm orm.Orm
}

func NewDatabaseChannel(o orm.Orm) *DatabaseChannel {
	return &DatabaseChannel{orm: o}
}

func (c *DatabaseChannel) Name() string { return "database" }

func (c *DatabaseChannel) Send(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	route, payload, err := c.Resolve(notifiable, n)
	if err != nil {
		return err
	}
	return c.Deliver(route, payload)
}

// resolvedRecord is what actually gets persisted — built at Resolve time
// so notifiable/notification's type names (%T) are captured while those
// values are still live, then carried as plain data to Deliver.
type resolvedRecord struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	NotifiableType string          `json:"notifiable_type"`
	Connection     string          `json:"connection"`
	Data           json.RawMessage `json:"data"`
}

func (c *DatabaseChannel) Resolve(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) (string, []byte, error) {
	notifiableID := notifiable.RouteNotificationFor("database")
	if notifiableID == "" {
		return "", nil, errors.NotificationDatabaseEmptyRoute.Args(notifiable)
	}

	var data map[string]any
	if dn, ok := n.(contractsnotification.DatabaseNotification); ok {
		data = dn.ToDatabase(notifiable)
	} else {
		data = map[string]any{"type": fmt.Sprintf("%T", n)}
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", nil, errors.NotificationDatabaseMarshalDataFailed.Args(n, err)
	}

	var id string
	if withID, ok := n.(contractsnotification.NotificationWithID); ok {
		id = withID.ID()
	}
	if id == "" {
		id = uuid.NewString()
	}

	var connection string
	if dr, ok := n.(contractsnotification.DatabaseRoutable); ok {
		connection = dr.DatabaseConnection()
	}

	record := resolvedRecord{
		ID:             id,
		Type:           fmt.Sprintf("%T", n),
		NotifiableType: fmt.Sprintf("%T", notifiable),
		Connection:     connection,
		Data:           dataJSON,
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return "", nil, errors.NotificationDatabaseMarshalRecordFailed.Args(err)
	}

	return notifiableID, payload, nil
}

func (c *DatabaseChannel) Deliver(route string, payload []byte) error {
	if route == "" {
		return nil
	}

	var record resolvedRecord
	if err := json.Unmarshal(payload, &record); err != nil {
		return errors.NotificationDatabaseUnmarshalRecordFailed.Args(err)
	}

	model := &DatabaseNotificationModel{
		ID:             record.ID,
		Type:           record.Type,
		NotifiableType: record.NotifiableType,
		NotifiableID:   route,
		Data:           string(record.Data),
	}

	o := c.orm
	if record.Connection != "" {
		o = o.Connection(record.Connection)
	}

	if err := o.Query().Create(model); err != nil {
		return errors.NotificationDatabaseInsertFailed.Args(err)
	}

	return nil
}

var _ contractsnotification.ResolvableChannel = (*DatabaseChannel)(nil)
