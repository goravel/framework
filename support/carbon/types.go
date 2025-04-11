package carbon

import (
	"github.com/dromara/carbon/v2"
)

type DateTime = carbon.DateTime

func NewDateTime(carbon Carbon) DateTime {
	return DateTime{Carbon: carbon}
}

type DateTimeMilli = carbon.DateTimeMilli

func NewDateTimeMilli(carbon Carbon) DateTimeMilli {
	return DateTimeMilli{Carbon: carbon}
}

type DateTimeMicro = carbon.DateTimeMicro

func NewDateTimeMicro(carbon Carbon) DateTimeMicro {
	return DateTimeMicro{Carbon: carbon}
}

type DateTimeNano = carbon.DateTimeNano

func NewDateTimeNano(carbon Carbon) DateTimeNano {
	return DateTimeNano{Carbon: carbon}
}

type Date = carbon.Date

func NewDate(carbon Carbon) Date {
	return Date{Carbon: carbon}
}

type DateMilli = carbon.DateMilli

func NewDateMilli(carbon Carbon) DateMilli {
	return DateMilli{Carbon: carbon}
}

type DateMicro = carbon.DateMicro

func NewDateMicro(carbon Carbon) DateMicro {
	return DateMicro{Carbon: carbon}
}

type DateNano = carbon.DateNano

func NewDateNano(carbon Carbon) DateNano {
	return DateNano{Carbon: carbon}
}

type Time = carbon.Time

func NewTime(carbon Carbon) Time {
	return Time{Carbon: carbon}
}

type TimeMilli = carbon.TimeMilli

func NewTimeMilli(carbon Carbon) TimeMilli {
	return TimeMilli{Carbon: carbon}
}

type TimeMicro = carbon.TimeMicro

func NewTimeMicro(carbon Carbon) TimeMicro {
	return TimeMicro{Carbon: carbon}
}

type TimeNano = carbon.TimeNano

func NewTimeNano(carbon Carbon) TimeNano {
	return TimeNano{Carbon: carbon}
}

type Timestamp = carbon.Timestamp

func NewTimestamp(carbon Carbon) Timestamp {
	return Timestamp{Carbon: carbon}
}

type TimestampMilli = carbon.TimestampMilli

func NewTimestampMilli(carbon Carbon) TimestampMilli {
	return TimestampMilli{Carbon: carbon}
}

type TimestampMicro = carbon.TimestampMicro

func NewTimestampMicro(carbon Carbon) TimestampMicro {
	return TimestampMicro{Carbon: carbon}
}

type TimestampNano = carbon.TimestampNano

func NewTimestampNano(carbon Carbon) TimestampNano {
	return TimestampNano{Carbon: carbon}
}
