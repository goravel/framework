package carbon

import (
	"github.com/dromara/carbon/v2"
)

type LayoutFactory = carbon.LayoutFactory

type LayoutType[T LayoutFactory] = carbon.LayoutType[T]

func NewLayoutType[T LayoutFactory](c *Carbon) LayoutType[T] {
	return LayoutType[T](carbon.NewLayoutType[T](c))
}

type FormatFactory = carbon.FormatFactory

type FormatType[T FormatFactory] = carbon.FormatType[T]

func NewFormatType[T FormatFactory](c *Carbon) FormatType[T] {
	return FormatType[T](carbon.NewFormatType[T](c))
}

type TimestampFactory = carbon.TimestampFactory

type TimestampType[T TimestampFactory] = carbon.TimestampType[T]

func NewTimestampType[T TimestampFactory](c *Carbon) TimestampType[T] {
	return TimestampType[T](carbon.NewTimestampType[T](c))
}

type DateTime = carbon.DateTime
type DateTimeMilli = carbon.DateTimeMilli
type DateTimeMicro = carbon.DateTimeMicro
type DateTimeNano = carbon.DateTimeNano

type Date = carbon.Date
type DateMilli = carbon.DateMilli
type DateMicro = carbon.DateMicro
type DateNano = carbon.DateNano

type Time = carbon.Time
type TimeMilli = carbon.TimeMilli
type TimeMicro = carbon.TimeMicro
type TimeNano = carbon.TimeNano

type Timestamp = carbon.Timestamp
type TimestampMilli = carbon.TimestampMilli
type TimestampMicro = carbon.TimestampMicro
type TimestampNano = carbon.TimestampNano
