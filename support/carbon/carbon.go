package carbon

import (
	stdtime "time"

	"github.com/dromara/carbon/v2"
)

type Carbon = carbon.Carbon

// SetTimezone sets timezone.
func SetTimezone(timezone string) {
	carbon.SetTimezone(timezone)
}

// SetLocale sets language locale.
func SetLocale(locale string) {
	carbon.SetLocale(locale)
}

// SetTestNow sets the test time, remember to clean after use.
func SetTestNow(testTime Carbon) {
	carbon.SetTestNow(testTime)
}

// UnsetTestNow unsets the test time.
// Deprecated: it will be deprecated in the future, please use 'CleanTestNow' instead
func UnsetTestNow() {
	CleanTestNow()
}

// CleanTestNow cleans the test time.
func CleanTestNow() {
	carbon.CleanTestNow()
}

// IsTestNow determines if the test now time is set.
func IsTestNow() bool {
	return carbon.IsTestNow()
}

// Now returns a Carbon object of now.
func Now(timezone ...string) Carbon {
	return carbon.Now(timezone...)
}

// Parse returns a Carbon object of given value.
func Parse(value string, timezone ...string) Carbon {
	return carbon.Parse(value, timezone...)
}

// ParseByFormat returns a Carbon object by a confirmed format.
func ParseByFormat(value, format string, timezone ...string) Carbon {
	return carbon.ParseByFormat(value, format, timezone...)
}

// ParseByLayout returns a Carbon object by a confirmed layout.
func ParseByLayout(value, layout string, timezone ...string) Carbon {
	return carbon.ParseByLayout(value, layout, timezone...)
}

// ParseWithFormats returns a Carbon object with multiple fuzzy formats.
func ParseWithFormats(value string, formats []string, timezone ...string) Carbon {
	return carbon.ParseWithFormats(value, formats, timezone...)
}

// ParseWithLayouts returns a Carbon object with multiple fuzzy layouts.
func ParseWithLayouts(value string, layouts []string, timezone ...string) Carbon {
	return carbon.ParseWithLayouts(value, layouts, timezone...)
}

// FromTimestamp returns a Carbon object of given timestamp.
func FromTimestamp(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestamp(timestamp, timezone...)
}

// FromTimestampMilli returns a Carbon object of given millisecond timestamp.
func FromTimestampMilli(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestampMilli(timestamp, timezone...)
}

// FromTimestampMicro returns a Carbon object of given microsecond timestamp.
func FromTimestampMicro(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestampMicro(timestamp, timezone...)
}

// FromTimestampNano returns a Carbon object of given nanosecond timestamp.
func FromTimestampNano(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestampNano(timestamp, timezone...)
}

// FromDateTime returns a Carbon object of given date and time.
func FromDateTime(year int, month int, day int, hour int, minute int, second int, timezone ...string) Carbon {
	return carbon.CreateFromDateTime(year, month, day, hour, minute, second, timezone...)
}

// FromDateTimeMilli returns a Carbon object of given date and millisecond time.
func FromDateTimeMilli(year int, month int, day int, hour int, minute int, second int, millisecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateTimeMilli(year, month, day, hour, minute, second, millisecond, timezone...)
}

// FromDateTimeMicro returns a Carbon object of given date and microsecond time.
func FromDateTimeMicro(year int, month int, day int, hour int, minute int, second int, microsecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateTimeMicro(year, month, day, hour, minute, second, microsecond, timezone...)
}

// FromDateTimeNano returns a Carbon object of given date and nanosecond time.
func FromDateTimeNano(year int, month int, day int, hour int, minute int, second int, nanosecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateTimeNano(year, month, day, hour, minute, second, nanosecond, timezone...)
}

// FromDate returns a Carbon object of given date.
func FromDate(year int, month int, day int, timezone ...string) Carbon {
	return carbon.CreateFromDate(year, month, day, timezone...)
}

// FromDateMilli returns a Carbon object of given millisecond date.
func FromDateMilli(year int, month int, day int, millisecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateMilli(year, month, day, millisecond, timezone...)
}

// FromDateMicro returns a Carbon object of given microsecond date.
func FromDateMicro(year int, month int, day int, microsecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateMicro(year, month, day, microsecond, timezone...)
}

// FromDateNano returns a Carbon object of given nanosecond date.
func FromDateNano(year int, month int, day int, nanosecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateNano(year, month, day, nanosecond, timezone...)
}

// FromTime returns a Carbon object of given time.
func FromTime(hour int, minute int, second int, timezone ...string) Carbon {
	return carbon.CreateFromTime(hour, minute, second, timezone...)
}

// FromTimeMilli returns a Carbon object of given millisecond time.
func FromTimeMilli(hour int, minute int, second int, millisecond int, timezone ...string) Carbon {
	return carbon.CreateFromTimeMilli(hour, minute, second, millisecond, timezone...)
}

// FromTimeMicro returns a Carbon object of given microsecond time.
func FromTimeMicro(hour int, minute int, second int, microsecond int, timezone ...string) Carbon {
	return carbon.CreateFromTimeMicro(hour, minute, second, microsecond, timezone...)
}

// FromTimeNano returns a Carbon object of given nanosecond time.
func FromTimeNano(hour int, minute int, second int, nanosecond int, timezone ...string) Carbon {
	return carbon.CreateFromTimeNano(hour, minute, second, nanosecond, timezone...)
}

// FromStdTime returns a Carbon object of given time.Time object.
func FromStdTime(time stdtime.Time, timezone ...string) Carbon {
	return carbon.CreateFromStdTime(time, timezone...)
}
