package channels

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	contractsnotification "github.com/goravel/framework/contracts/notification"

	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/log"
)

// DatabaseNotificationModel is the ORM model written to the notifications
// table. Column shape matches the migration in
// notification/console/table_command.go — keep both in sync.
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

// DatabaseChannel persists notifications to the database.
type DatabaseChannel struct {
	orm    orm.Orm
	log    log.Log
	config contractsconfig.Config
}

func NewDatabaseChannel(o orm.Orm, logger log.Log, config contractsconfig.Config) *DatabaseChannel {
	return &DatabaseChannel{orm: o, log: logger, config: config}
}

func (c *DatabaseChannel) Name() string { return "database" }

func (c *DatabaseChannel) Send(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	return c.SendNow(notifiable, n)
}

// SendNow is identical to Send — DatabaseChannel has no queued mode of
// its own, only Manager does. Exists so callers can bypass Manager's
// queue routing entirely: facades.Notification().Channel("database").SendNow(u, n).
func (c *DatabaseChannel) SendNow(
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
	Data           json.RawMessage `json:"data"`
}

func (c *DatabaseChannel) Resolve(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) (string, []byte, error) {
	notifiableID := notifiable.RouteNotificationFor("database")
	if notifiableID == "" {
		return "", nil, fmt.Errorf("database channel: %T.RouteNotificationFor(\"database\") returned empty ID", notifiable)
	}

	var data map[string]any
	if dn, ok := n.(contractsnotification.DatabaseNotification); ok {
		data = dn.ToDatabase(notifiable)
	} else {
		data = map[string]any{"type": fmt.Sprintf("%T", n)}
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", nil, fmt.Errorf("database channel: failed to marshal payload for %T: %w", n, err)
	}

	var id string
	if withID, ok := n.(contractsnotification.NotificationWithID); ok {
		id = withID.ID()
	}
	if id == "" {
		id = uuid.NewString()
	}

	record := resolvedRecord{
		ID:             id,
		Type:           fmt.Sprintf("%T", n),
		NotifiableType: fmt.Sprintf("%T", notifiable),
		Data:           dataJSON,
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return "", nil, fmt.Errorf("database channel: failed to marshal record: %w", err)
	}

	return notifiableID, payload, nil
}

func (c *DatabaseChannel) Deliver(route string, payload []byte) error {
	if route == "" {
		return nil
	}

	var record resolvedRecord
	if err := json.Unmarshal(payload, &record); err != nil {
		return fmt.Errorf("database channel: failed to unmarshal record: %w", err)
	}

	model := &DatabaseNotificationModel{
		ID:             record.ID,
		Type:           record.Type,
		NotifiableType: record.NotifiableType,
		NotifiableID:   route,
		Data:           string(record.Data),
	}

	// notification.channels.database.connection lets an app store notifications
	// on a connection other than the default (e.g. a separate read-heavy DB).
	o := c.orm
	if c.config != nil {
		if connection := c.config.GetString("notification.channels.database.connection"); connection != "" {
			o = o.Connection(connection)
		}
	}

	if err := o.Query().Create(model); err != nil {
		return fmt.Errorf("database channel: failed to insert notification record: %w", err)
	}

	c.log.Debugf("notifications: persisted %s to database (id=%s)", record.Type, record.ID)
	return nil
}

var _ contractsnotification.ResolvableChannel = (*DatabaseChannel)(nil)
