package orm

import "time"

type Model struct {
	ID        uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}
