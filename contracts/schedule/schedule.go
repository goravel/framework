package schedule

import (
	"github.com/goravel/framework/schedule/support"
)

type Schedule interface {
	//Call Add a new callback event to the schedule.
	Call(callback func()) *support.Event

	//Command Add a new Artisan command event to the schedule.
	Command(command string) *support.Event

	//Register schedules.
	Register(events []*support.Event)

	//Run schedules.
	Run()
}
