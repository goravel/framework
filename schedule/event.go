package schedule

import (
	"strings"

	"github.com/goravel/framework/contracts/schedule"
)

type Event struct {
	callback            func()
	command             string
	cron                string
	delayIfStillRunning bool
	name                string
	onOneServer         bool
	skipIfStillRunning  bool
}

func NewCallbackEvent(callback func()) *Event {
	return &Event{callback: callback}
}

func NewCommandEvent(command string) *Event {
	return &Event{command: command, name: command}
}

//At Schedule the command at a given time.
func (receiver *Event) At(time string) schedule.Event {
	return receiver.DailyAt(time)
}

//Cron The Cron expression representing the event's frequency.
func (receiver *Event) Cron(expression string) schedule.Event {
	receiver.cron = expression

	return receiver
}

//Daily Schedule the event to run daily.
func (receiver *Event) Daily() schedule.Event {
	event := receiver.Cron(receiver.spliceIntoPosition(1, "0"))

	return event.Cron(receiver.spliceIntoPosition(2, "0"))
}

//DailyAt Schedule the event to run daily at a given time (10:00, 19:30, etc).
func (receiver *Event) DailyAt(time string) schedule.Event {
	segments := strings.Split(time, ":")
	event := receiver.Cron(receiver.spliceIntoPosition(2, segments[0]))

	if len(segments) == 2 {
		return event.Cron(receiver.spliceIntoPosition(1, segments[1]))
	} else {
		return event.Cron(receiver.spliceIntoPosition(1, "0"))
	}
}

//DelayIfStillRunning Do not allow the event to overlap each other.
func (receiver *Event) DelayIfStillRunning() schedule.Event {
	receiver.delayIfStillRunning = true

	return receiver
}

//EveryMinute Schedule the event to run every minute.
func (receiver *Event) EveryMinute() schedule.Event {
	return receiver.Cron(receiver.spliceIntoPosition(1, "*"))
}

//EveryTwoMinutes Schedule the event to run every two minutes.
func (receiver *Event) EveryTwoMinutes() schedule.Event {
	return receiver.Cron(receiver.spliceIntoPosition(1, "*/2"))
}

//EveryThreeMinutes Schedule the event to run every three minutes.
func (receiver *Event) EveryThreeMinutes() schedule.Event {
	return receiver.Cron(receiver.spliceIntoPosition(1, "*/3"))
}

//EveryFourMinutes Schedule the event to run every four minutes.
func (receiver *Event) EveryFourMinutes() schedule.Event {
	return receiver.Cron(receiver.spliceIntoPosition(1, "*/4"))
}

//EveryFiveMinutes Schedule the event to run every five minutes.
func (receiver *Event) EveryFiveMinutes() schedule.Event {
	return receiver.Cron(receiver.spliceIntoPosition(1, "*/5"))
}

//EveryTenMinutes Schedule the event to run every ten minutes.
func (receiver *Event) EveryTenMinutes() schedule.Event {
	return receiver.Cron(receiver.spliceIntoPosition(1, "*/10"))
}

//EveryFifteenMinutes Schedule the event to run every fifteen minutes.
func (receiver *Event) EveryFifteenMinutes() schedule.Event {
	return receiver.Cron(receiver.spliceIntoPosition(1, "*/15"))
}

//EveryThirtyMinutes Schedule the event to run every thirty minutes.
func (receiver *Event) EveryThirtyMinutes() schedule.Event {
	return receiver.Cron(receiver.spliceIntoPosition(1, "0,30"))
}

//EveryTwoHours Schedule the event to run every two hours.
func (receiver *Event) EveryTwoHours() schedule.Event {
	event := receiver.Cron(receiver.spliceIntoPosition(1, "0"))

	return event.Cron(receiver.spliceIntoPosition(2, "*/2"))
}

//EveryThreeHours Schedule the event to run every three hours.
func (receiver *Event) EveryThreeHours() schedule.Event {
	event := receiver.Cron(receiver.spliceIntoPosition(1, "0"))

	return event.Cron(receiver.spliceIntoPosition(2, "*/3"))
}

//EveryFourHours Schedule the event to run every four hours.
func (receiver *Event) EveryFourHours() schedule.Event {
	event := receiver.Cron(receiver.spliceIntoPosition(1, "0"))

	return event.Cron(receiver.spliceIntoPosition(2, "*/4"))
}

//EverySixHours Schedule the event to run every six hours.
func (receiver *Event) EverySixHours() schedule.Event {
	event := receiver.Cron(receiver.spliceIntoPosition(1, "0"))

	return event.Cron(receiver.spliceIntoPosition(2, "*/6"))
}

func (receiver *Event) GetCron() string {
	if receiver.cron == "" {
		receiver.cron = "* * * * *"
	}

	return receiver.cron
}

func (receiver *Event) GetCommand() string {
	return receiver.command
}

func (receiver *Event) GetCallback() func() {
	return receiver.callback
}

func (receiver *Event) GetName() string {
	return receiver.name
}

func (receiver *Event) GetSkipIfStillRunning() bool {
	return receiver.skipIfStillRunning
}

func (receiver *Event) GetDelayIfStillRunning() bool {
	return receiver.delayIfStillRunning
}

//Hourly Schedule the event to run hourly.
func (receiver *Event) Hourly() schedule.Event {
	return receiver.Cron(receiver.spliceIntoPosition(1, "0"))
}

//HourlyAt Schedule the event to run hourly at a given offset in the hour.
func (receiver *Event) HourlyAt(offset []string) schedule.Event {
	return receiver.Cron(receiver.spliceIntoPosition(1, strings.Join(offset, ",")))
}

func (receiver *Event) IsOnOneServer() bool {
	return receiver.onOneServer
}

func (receiver *Event) Name(name string) schedule.Event {
	receiver.name = name

	return receiver
}

func (receiver *Event) OnOneServer() schedule.Event {
	receiver.onOneServer = true

	return receiver
}

//SkipIfStillRunning Do not allow the event to overlap each other.
func (receiver *Event) SkipIfStillRunning() schedule.Event {
	receiver.skipIfStillRunning = true

	return receiver
}

//spliceIntoPosition Splice the given value into the given position of the expression.
func (receiver *Event) spliceIntoPosition(position int, value string) string {
	segments := strings.Split(receiver.GetCron(), " ")

	segments[position-1] = value

	return strings.Join(segments, " ")
}
