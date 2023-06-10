package time

import (
	"sync"
	stdtime "time"

	"github.com/golang-module/carbon/v2"
)

var once sync.Once

var internalClock *Clock

type Time = carbon.Carbon

type Clock struct {
	testNow bool
	time    Time
}

// getInstance Get a singleton Clock object.
func getInstance(timezone ...string) *Clock {
	once.Do(func() {
		internalClock = &Clock{
			testNow: false,
			time:    carbon.Now(timezone...),
		}
	})

	return internalClock
}

// SetTestNow Set the test now time.
func SetTestNow(testNow Time) {
	t := getInstance()
	t.testNow = true
	t.time = testNow
}

// UnsetTestNow Unset the test now time.
func UnsetTestNow() {
	t := getInstance()
	t.testNow = false
	t.time = carbon.Now()
}

// IsTestNow Determine if the test now time is set.
func IsTestNow() bool {
	t := getInstance()

	return t.testNow
}

// Now return a Time object of now.
func Now(timezone ...string) Time {
	t := getInstance(timezone...)

	if t.testNow {
		return t.time
	}

	now := carbon.Now(timezone...)
	t.time = now

	return t.time
}

// Parse return a Time object of given value.
func Parse(value string, timezone ...string) Time {
	t := getInstance(timezone...)
	parse := carbon.Parse(value, timezone...)
	t.time = parse

	return t.time
}

// ParseByFormat return a Time object of given value and format.
func ParseByFormat(value, format string, timezone ...string) Time {
	t := getInstance(timezone...)
	parseByFormat := carbon.ParseByFormat(value, format, timezone...)
	t.time = parseByFormat

	return t.time
}

// ParseByLayout return a Time object of given value and layout.
func ParseByLayout(value, layout string, timezone ...string) Time {
	t := getInstance(timezone...)
	parseByLayout := carbon.ParseByLayout(value, layout, timezone...)
	t.time = parseByLayout

	return t.time
}

// FromTimestamp return a Time object of given timestamp.
func FromTimestamp(timestamp int64, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromTimestamp := carbon.CreateFromTimestamp(timestamp, timezone...)
	t.time = createFromTimestamp

	return t.time
}

// FromTimestampMilli return a Time object of given millisecond timestamp.
func FromTimestampMilli(timestamp int64, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromTimestampMilli := carbon.CreateFromTimestampMilli(timestamp, timezone...)
	t.time = createFromTimestampMilli

	return t.time
}

// FromTimestampMicro return a Time object of given microsecond timestamp.
func FromTimestampMicro(timestamp int64, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromTimestampMicro := carbon.CreateFromTimestampMicro(timestamp, timezone...)
	t.time = createFromTimestampMicro

	return t.time
}

// FromTimestampNano return a Time object of given nanosecond timestamp.
func FromTimestampNano(timestamp int64, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromTimestampNano := carbon.CreateFromTimestampNano(timestamp, timezone...)
	t.time = createFromTimestampNano

	return t.time
}

// FromDateTime return a Time object of given date and time.
func FromDateTime(year int, month int, day int, hour int, minute int, second int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromDateTime := carbon.CreateFromDateTime(year, month, day, hour, minute, second, timezone...)
	t.time = createFromDateTime

	return t.time
}

// FromDateTimeMilli return a Time object of given date and millisecond time.
func FromDateTimeMilli(year int, month int, day int, hour int, minute int, second int, millisecond int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromDateTimeMilli := carbon.CreateFromDateTimeMilli(year, month, day, hour, minute, second, millisecond, timezone...)
	t.time = createFromDateTimeMilli

	return t.time
}

// FromDateTimeMicro return a Time object of given date and microsecond time.
func FromDateTimeMicro(year int, month int, day int, hour int, minute int, second int, microsecond int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromDateTimeMicro := carbon.CreateFromDateTimeMicro(year, month, day, hour, minute, second, microsecond, timezone...)
	t.time = createFromDateTimeMicro

	return t.time
}

// FromDateTimeNano return a Time object of given date and nanosecond time.
func FromDateTimeNano(year int, month int, day int, hour int, minute int, second int, nanosecond int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromDateTimeNano := carbon.CreateFromDateTimeNano(year, month, day, hour, minute, second, nanosecond, timezone...)
	t.time = createFromDateTimeNano

	return t.time
}

// FromDate return a Time object of given date.
func FromDate(year int, month int, day int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromDate := carbon.CreateFromDate(year, month, day, timezone...)
	t.time = createFromDate

	return t.time
}

// FromDateMilli return a Time object of given millisecond date.
func FromDateMilli(year int, month int, day int, millisecond int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromDateMilli := carbon.CreateFromDateMilli(year, month, day, millisecond, timezone...)
	t.time = createFromDateMilli

	return t.time
}

// FromDateMicro return a Time object of given microsecond date.
func FromDateMicro(year int, month int, day int, microsecond int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromDateMicro := carbon.CreateFromDateMicro(year, month, day, microsecond, timezone...)
	t.time = createFromDateMicro

	return t.time
}

// FromDateNano return a Time object of given nanosecond date.
func FromDateNano(year int, month int, day int, nanosecond int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromDateNano := carbon.CreateFromDateNano(year, month, day, nanosecond, timezone...)
	t.time = createFromDateNano

	return t.time
}

// FromTime return a Time object of given time.
func FromTime(hour int, minute int, second int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromTime := carbon.CreateFromTime(hour, minute, second, timezone...)
	t.time = createFromTime

	return t.time
}

// FromTimeMilli return a Time object of given millisecond time.
func FromTimeMilli(hour int, minute int, second int, millisecond int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromTimeMilli := carbon.CreateFromTimeMilli(hour, minute, second, millisecond, timezone...)
	t.time = createFromTimeMilli

	return t.time
}

// FromTimeMicro return a Time object of given microsecond time.
func FromTimeMicro(hour int, minute int, second int, microsecond int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromTimeMicro := carbon.CreateFromTimeMicro(hour, minute, second, microsecond, timezone...)
	t.time = createFromTimeMicro

	return t.time
}

// FromTimeNano return a Time object of given nanosecond time.
func FromTimeNano(hour int, minute int, second int, nanosecond int, timezone ...string) Time {
	t := getInstance(timezone...)
	createFromTimeNano := carbon.CreateFromTimeNano(hour, minute, second, nanosecond, timezone...)
	t.time = createFromTimeNano

	return t.time
}

// FromStdTime return a Time object of given time.Time object.
func FromStdTime(time stdtime.Time) Time {
	t := getInstance()
	createFromStdTime := carbon.FromStdTime(time)
	t.time = createFromStdTime

	return t.time
}
