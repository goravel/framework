package carbon

import (
	"github.com/dromara/carbon/v2"
)

type DateTime = carbon.DateTime

func NewDateTime(c *Carbon) *DateTime {
	return &DateTime{Carbon: c}
}

type DateTimeMilli = carbon.DateTimeMilli

func NewDateTimeMilli(c *Carbon) *DateTimeMilli {
	return &DateTimeMilli{Carbon: c}
}

type DateTimeMicro = carbon.DateTimeMicro

func NewDateTimeMicro(c *Carbon) *DateTimeMicro {
	return &DateTimeMicro{Carbon: c}
}

type DateTimeNano = carbon.DateTimeNano

func NewDateTimeNano(c *Carbon) *DateTimeNano {
	return &DateTimeNano{Carbon: c}
}

type Date = carbon.Date

func NewDate(c *Carbon) *Date {
	return &Date{Carbon: c}
}

type DateMilli = carbon.DateMilli

func NewDateMilli(c *Carbon) *DateMilli {
	return &DateMilli{Carbon: c}
}

type DateMicro = carbon.DateMicro

func NewDateMicro(c *Carbon) *DateMicro {
	return &DateMicro{Carbon: c}
}

type DateNano = carbon.DateNano

func NewDateNano(c *Carbon) *DateNano {
	return &DateNano{Carbon: c}
}

type Time = carbon.Time

func NewTime(c *Carbon) *Time {
	return &Time{Carbon: c}
}

type TimeMilli = carbon.TimeMilli

func NewTimeMilli(c *Carbon) *TimeMilli {
	return &TimeMilli{Carbon: c}
}

type TimeMicro = carbon.TimeMicro

func NewTimeMicro(c *Carbon) *TimeMicro {
	return &TimeMicro{Carbon: c}
}

type TimeNano = carbon.TimeNano

func NewTimeNano(c *Carbon) *TimeNano {
	return &TimeNano{Carbon: c}
}

type Timestamp = carbon.Timestamp

func NewTimestamp(c *Carbon) *Timestamp {
	return &Timestamp{Carbon: c}
}

type TimestampMilli = carbon.TimestampMilli

func NewTimestampMilli(c *Carbon) *TimestampMilli {
	return &TimestampMilli{Carbon: c}
}

type TimestampMicro = carbon.TimestampMicro

func NewTimestampMicro(c *Carbon) *TimestampMicro {
	return &TimestampMicro{Carbon: c}
}

type TimestampNano = carbon.TimestampNano

func NewTimestampNano(c *Carbon) *TimestampNano {
	return &TimestampNano{Carbon: c}
}
