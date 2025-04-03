package carbon

import (
	stdtime "time"

	"github.com/dromara/carbon/v2"
)

// SetTestNow Set the test time. Remember to unset after used.
func SetTestNow(testTime Carbon) {
	carbon.SetTestNow(testTime)
}

// UnsetTestNow Unset the test time.
func UnsetTestNow() {
	carbon.CleanTestNow()
}

// IsTestNow Determine if the test now time is set.
func IsTestNow() bool {
	return carbon.IsTestNow()
}

// SetTimezone sets timezone.
// 设置时区
func SetTimezone(timezone string) {
	carbon.SetTimezone(timezone)
}

// Now return a Carbon object of now.
func Now(timezone ...string) Carbon {
	return carbon.Now(timezone...)
}

// Parse return a Carbon object of given value.
func Parse(value string, timezone ...string) Carbon {
	return carbon.Parse(value, timezone...)
}

// ParseByFormat return a Carbon object of given value and format.
func ParseByFormat(value, format string, timezone ...string) Carbon {
	return carbon.ParseByFormat(value, format, timezone...)
}

// ParseByLayout return a Carbon object of given value and layout.
func ParseByLayout(value, layout string, timezone ...string) Carbon {
	return carbon.ParseByLayout(value, layout, timezone...)
}

// FromTimestamp return a Carbon object of given timestamp.
func FromTimestamp(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestamp(timestamp, timezone...)
}

// FromTimestampMilli return a Carbon object of given millisecond timestamp.
func FromTimestampMilli(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestampMilli(timestamp, timezone...)
}

// FromTimestampMicro return a Carbon object of given microsecond timestamp.
func FromTimestampMicro(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestampMicro(timestamp, timezone...)
}

// FromTimestampNano return a Carbon object of given nanosecond timestamp.
func FromTimestampNano(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestampNano(timestamp, timezone...)
}

// FromDateTime return a Carbon object of given date and time.
func FromDateTime(year int, month int, day int, hour int, minute int, second int, timezone ...string) Carbon {
	return carbon.CreateFromDateTime(year, month, day, hour, minute, second, timezone...)
}

// FromDateTimeMilli return a Carbon object of given date and millisecond time.
func FromDateTimeMilli(year int, month int, day int, hour int, minute int, second int, millisecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateTimeMilli(year, month, day, hour, minute, second, millisecond, timezone...)
}

// FromDateTimeMicro return a Carbon object of given date and microsecond time.
func FromDateTimeMicro(year int, month int, day int, hour int, minute int, second int, microsecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateTimeMicro(year, month, day, hour, minute, second, microsecond, timezone...)
}

// FromDateTimeNano return a Carbon object of given date and nanosecond time.
func FromDateTimeNano(year int, month int, day int, hour int, minute int, second int, nanosecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateTimeNano(year, month, day, hour, minute, second, nanosecond, timezone...)
}

// FromDate return a Carbon object of given date.
func FromDate(year int, month int, day int, timezone ...string) Carbon {
	return carbon.CreateFromDate(year, month, day, timezone...)
}

// FromDateMilli return a Carbon object of given millisecond date.
func FromDateMilli(year int, month int, day int, millisecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateMilli(year, month, day, millisecond, timezone...)
}

// FromDateMicro return a Carbon object of given microsecond date.
func FromDateMicro(year int, month int, day int, microsecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateMicro(year, month, day, microsecond, timezone...)
}

// FromDateNano return a Carbon object of given nanosecond date.
func FromDateNano(year int, month int, day int, nanosecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateNano(year, month, day, nanosecond, timezone...)
}

// FromTime return a Carbon object of given time.
func FromTime(hour int, minute int, second int, timezone ...string) Carbon {
	return carbon.CreateFromTime(hour, minute, second, timezone...)
}

// FromTimeMilli return a Carbon object of given millisecond time.
func FromTimeMilli(hour int, minute int, second int, millisecond int, timezone ...string) Carbon {
	return carbon.CreateFromTimeMilli(hour, minute, second, millisecond, timezone...)
}

// FromTimeMicro return a Carbon object of given microsecond time.
func FromTimeMicro(hour int, minute int, second int, microsecond int, timezone ...string) Carbon {
	return carbon.CreateFromTimeMicro(hour, minute, second, microsecond, timezone...)
}

// FromTimeNano return a Carbon object of given nanosecond time.
func FromTimeNano(hour int, minute int, second int, nanosecond int, timezone ...string) Carbon {
	return carbon.CreateFromTimeNano(hour, minute, second, nanosecond, timezone...)
}

// FromStdTime return a Carbon object of given time.Time object.
func FromStdTime(time stdtime.Time) Carbon {
	return carbon.CreateFromStdTime(time)
}
