package queue

import "time"

type Task struct {
	Uuid string   `json:"uuid"`
	Data TaskData `json:"data"`
}

type TaskData struct {
	Job     Job        `json:"job"`
	Args    []Arg      `json:"args"`
	Delay   *time.Time `json:"delay"`
	Chained []TaskData `json:"chained"`
}
