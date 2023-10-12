package schedule

//go:generate mockery --name=Schedule
type Schedule interface {
	// Call add a new callback event to the schedule.
	Call(callback func()) Event
	// Command adds a new Artisan command event to the schedule.
	Command(command string) Event
	// Register schedules.
	Register(events []Event)
	// Run schedules.
	Run()
}
