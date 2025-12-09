package models

import (
    "github.com/goravel/framework/database/orm"
    "github.com/goravel/framework/support/carbon"
)

// Notification represents a stored notification record.
// Fields align with the `notifications` table schema.
type Notification struct {
    ID             string           `json:"id"`
    Type           string           `json:"type"`
    NotifiableType string           `json:"notifiable_type"`
    NotifiableId   string           `json:"notifiable_id"`
    Data           string           `json:"data"`
    ReadAt         *carbon.DateTime `json:"read_at"`
    orm.Timestamps
}
