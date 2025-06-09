package models

import "github.com/goravel/framework/support/carbon"

type Job struct {
	ID          uint             `db:"id"`
	Queue       string           `db:"queue"`
	Payload     string           `db:"payload"`
	Attempts    int              `db:"attempts"`
	ReservedAt  *carbon.DateTime `db:"reserved_at"`
	AvailableAt *carbon.DateTime `db:"available_at"`
	CreatedAt   *carbon.DateTime `db:"created_at"`
}

func (r *Job) Increment() int {
	r.Attempts++

	return r.Attempts
}

func (r *Job) Touch() *carbon.DateTime {
	r.ReservedAt = carbon.NewDateTime(carbon.Now())

	return r.ReservedAt
}

type FailedJob struct {
	ID         uint             `db:"id"`
	UUID       string           `db:"uuid"`
	Connection string           `db:"connection"`
	Queue      string           `db:"queue"`
	Payload    string           `db:"payload"`
	Exception  string           `db:"exception"`
	FailedAt   *carbon.DateTime `db:"failed_at"`
}
