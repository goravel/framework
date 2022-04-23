package support

import (
	"strings"
)

type Event struct {
	Command  string
	Callback func()
	cron                string
	withoutOverlapping  bool
	skipIfStillRunning  bool
	delayIfStillRunning bool
}

func (receiver *Event) GetCron() string {
	if receiver.cron == "" {
		receiver.cron = "* * * * *"
	}

	return receiver.cron
}

func (receiver *Event) GetSkipIfStillRunning() bool {
	return receiver.skipIfStillRunning
}

func (receiver *Event) GetDelayIfStillRunning() bool {
	return receiver.delayIfStillRunning
}

//Cron The Cron expression representing the event's frequency.
func (receiver *Event) Cron(expression string) *Event {
	receiver.cron = expression

	return receiver
}

//EveryMinute Schedule the event to run every minute.
func (receiver *Event) EveryMinute() *Event {
	return receiver.spliceIntoPosition(1, "*")
}

//EveryTwoMinutes Schedule the event to run every two minutes.
func (receiver *Event) EveryTwoMinutes() *Event {
	return receiver.spliceIntoPosition(1, "*/2")
}

//EveryThreeMinutes Schedule the event to run every three minutes.
func (receiver *Event) EveryThreeMinutes() *Event {
	return receiver.spliceIntoPosition(1, "*/3")
}

//EveryFourMinutes Schedule the event to run every four minutes.
func (receiver *Event) EveryFourMinutes() *Event {
	return receiver.spliceIntoPosition(1, "*/4")
}

//EveryFiveMinutes Schedule the event to run every five minutes.
func (receiver *Event) EveryFiveMinutes() *Event {
	return receiver.spliceIntoPosition(1, "*/5")
}

//EveryTenMinutes Schedule the event to run every ten minutes.
func (receiver *Event) EveryTenMinutes() *Event {
	return receiver.spliceIntoPosition(1, "*/10")
}

//EveryFifteenMinutes Schedule the event to run every fifteen minutes.
func (receiver *Event) EveryFifteenMinutes() *Event {
	return receiver.spliceIntoPosition(1, "*/15")
}

//EveryThirtyMinutes Schedule the event to run every thirty minutes.
func (receiver *Event) EveryThirtyMinutes() *Event {
	return receiver.spliceIntoPosition(1, "0,30")
}

//Hourly Schedule the event to run hourly.
func (receiver *Event) Hourly() *Event {
	return receiver.spliceIntoPosition(1, "0")
}

//HourlyAt Schedule the event to run hourly at a given offset in the hour.
func (receiver *Event) HourlyAt(offset []string) *Event {
	return receiver.spliceIntoPosition(1, strings.Join(offset, ","))
}

//EveryTwoHours Schedule the event to run every two hours.
func (receiver *Event) EveryTwoHours() *Event {
	return receiver.spliceIntoPosition(1, "0").spliceIntoPosition(2, "*/2")
}

//EveryThreeHours Schedule the event to run every three hours.
func (receiver *Event) EveryThreeHours() *Event {
	return receiver.spliceIntoPosition(1, "0").spliceIntoPosition(2, "*/3")
}

//EveryFourHours Schedule the event to run every four hours.
func (receiver *Event) EveryFourHours() *Event {
	return receiver.spliceIntoPosition(1, "0").spliceIntoPosition(2, "*/4")
}

//EverySixHours Schedule the event to run every six hours.
func (receiver *Event) EverySixHours() *Event {
	return receiver.spliceIntoPosition(1, "0").spliceIntoPosition(2, "*/6")
}

//Daily Schedule the event to run daily.
func (receiver *Event) Daily() *Event {
	return receiver.spliceIntoPosition(1, "0").spliceIntoPosition(2, "0")
}

//At Schedule the command at a given time.
func (receiver *Event) At(time string) *Event {
	return receiver.DailyAt(time)
}

//DailyAt Schedule the event to run daily at a given time (10:00, 19:30, etc).
func (receiver *Event) DailyAt(time string) *Event {
	segments := strings.Split(time, ":")
	receiver.spliceIntoPosition(2, segments[0])

	if len(segments) == 2 {
		receiver.spliceIntoPosition(1, segments[1])
	} else {
		receiver.spliceIntoPosition(1, "0")
	}

	return receiver
}

//SkipIfStillRunning Do not allow the event to overlap each other.
func (receiver *Event) SkipIfStillRunning() *Event {
	receiver.skipIfStillRunning = true

	return receiver
}

//DelayIfStillRunning Do not allow the event to overlap each other.
func (receiver *Event) DelayIfStillRunning() *Event {
	receiver.delayIfStillRunning = true

	return receiver
}

//spliceIntoPosition Splice the given value into the given position of the expression.
func (receiver *Event) spliceIntoPosition(position int, value string) *Event {
	segments := strings.Split(receiver.GetCron(), " ")

	segments[position-1] = value

	return receiver.Cron(strings.Join(segments, " "))
}
