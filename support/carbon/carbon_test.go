package carbon

import (
	"testing"
	stdtime "time"

	"github.com/dromara/carbon/v2"
	"github.com/stretchr/testify/assert"
)

func TestSetTestNow(t *testing.T) {
	SetTestNow(Now().SubHour())
	assert.True(t, IsTestNow())
	UnsetTestNow()
	assert.False(t, IsTestNow())
}

func TestSetTimezone(t *testing.T) {
	defer SetTimezone(UTC)

	SetTimezone(PRC)
	c := Parse("2020-01-01 00:00:00")
	assert.Equal(t, "2020-01-01 00:00:00 +0800 CST", c.ToString())
}

func TestTimezone(t *testing.T) {
	timezones := []string{Local, UTC, GMT, EET, WET, CET, EST, MST, Cuba, Egypt, Eire, Greenwich, Iceland, Iran,
		Israel, Jamaica, Japan, Libya, Poland, Portugal, PRC, Singapore, Turkey, Shanghai, Chongqing, Harbin, Urumqi,
		HongKong, Macao, Taipei, Tokyo, Saigon, Seoul, Bangkok, Dubai, NewYork, LosAngeles, Chicago, Moscow, London,
		Berlin, Paris, Rome, Sydney, Melbourne, Darwin}

	for _, timezone := range timezones {
		now := Now(timezone)
		assert.Nil(t, now.Error, timezone)
		assert.True(t, now.Timestamp() > 0, timezone)
	}

	now := Now()
	assert.Nil(t, now.Error)
	assert.True(t, now.Timestamp() > 0)
}

func TestNow(t *testing.T) {
	SetTestNow(Now().SubSeconds(10))
	stdtime.Sleep(2 * stdtime.Second)
	testNow := Now().Timestamp()
	UnsetTestNow()
	now := Now().Timestamp()
	assert.True(t, now-testNow >= 10)

	utcNow := Now()
	shanghaiNow := Now(Shanghai)
	assert.Equal(t, shanghaiNow.SubHours(8).ToDateTimeString(), utcNow.ToDateTimeString())
}

func TestParse(t *testing.T) {
	time := Parse("2020-01-01 00:00:00", carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestParseByFormat(t *testing.T) {
	time := ParseByFormat("2020-01-01 00:00:00", "Y-m-d H:i:s", carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestParseByLayout(t *testing.T) {
	time := ParseByLayout("2020-01-01 00:00:00", "2006-01-02 15:04:05", carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestFromTimestamp(t *testing.T) {
	time := FromTimestamp(1577836800, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestFromTimestampMilli(t *testing.T) {
	time := FromTimestampMilli(1577836800000, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestFromTimestampMicro(t *testing.T) {
	time := FromTimestampMicro(1577836800000000, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestFromTimestampNano(t *testing.T) {
	time := FromTimestampNano(1577836800000000000, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestFromDateTime(t *testing.T) {
	time := FromDateTime(2020, 1, 1, 0, 0, 0, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestFromDateTimeMilli(t *testing.T) {
	time := FromDateTimeMilli(2020, 1, 1, 0, 0, 0, 0, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestFromDateTimeMicro(t *testing.T) {
	time := FromDateTimeMicro(2020, 1, 1, 0, 0, 0, 0, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestFromDateTimeNano(t *testing.T) {
	time := FromDateTimeNano(2020, 1, 1, 0, 0, 0, 0, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestFromDate(t *testing.T) {
	time := FromDate(2020, 1, 1, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestFromDateMilli(t *testing.T) {
	time := FromDateMilli(2020, 1, 1, 999, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00.999", time.ToDateTimeMilliString(carbon.UTC))
}

func TestFromDateMicro(t *testing.T) {
	time := FromDateMicro(2020, 1, 1, 999999, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00.999999", time.ToDateTimeMicroString(carbon.UTC))
}

func TestFromDateNano(t *testing.T) {
	time := FromDateNano(2020, 1, 1, 999999999, carbon.UTC)
	assert.Equal(t, "2020-01-01 00:00:00.999999999", time.ToDateTimeNanoString(carbon.UTC))
}

func TestFromTime(t *testing.T) {
	date := Now().ToDateString(carbon.UTC)
	time := FromTime(0, 0, 0, carbon.UTC)
	assert.Equal(t, date+" 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestFromTimeMilli(t *testing.T) {
	date := Now().ToDateString(carbon.UTC)
	time := FromTimeMilli(0, 0, 0, 999, carbon.UTC)
	assert.Equal(t, date+" 00:00:00.999", time.ToDateTimeMilliString(carbon.UTC))
}

func TestFromTimeMicro(t *testing.T) {
	date := Now().ToDateString(carbon.UTC)
	time := FromTimeMicro(0, 0, 0, 999999, carbon.UTC)
	assert.Equal(t, date+" 00:00:00.999999", time.ToDateTimeMicroString(carbon.UTC))
}

func TestFromTimeNano(t *testing.T) {
	date := Now().ToDateString(carbon.UTC)
	time := FromTimeNano(0, 0, 0, 999999999, carbon.UTC)
	assert.Equal(t, date+" 00:00:00.999999999", time.ToDateTimeNanoString(carbon.UTC))
}

func TestFromStdTime(t *testing.T) {
	time := FromStdTime(stdtime.Date(2020, 1, 1, 0, 0, 0, 0, stdtime.UTC))
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(carbon.UTC))
}

func TestErrorTime(t *testing.T) {
	year, month, day, hour, minute, second, timestamp, timezone := 2020, 8, 5, 13, 14, 15, int64(1577855655), "xxx"

	assert.NotNil(t, FromDateTime(year, month, day, hour, minute, second, timezone).Error, "It should catch an exception in CreateFromDateTime()")
	assert.NotNil(t, FromDate(year, month, day, timezone).Error, "It should catch an exception in CreateFromDate()")
	assert.NotNil(t, FromTime(hour, minute, second, timezone).Error, "It should catch an exception in CreateFromTime()")
	assert.NotNil(t, FromTimestamp(timestamp, timezone).Error, "It should catch an exception in CreateFromTime()")
	assert.NotNil(t, FromTimestampMilli(timestamp, timezone).Error, "It should catch an exception in CreateFromTimestampMilli()")
	assert.NotNil(t, FromTimestampMicro(timestamp, timezone).Error, "It should catch an exception in CreateFromTimestampMicro()")
	assert.NotNil(t, FromTimestampNano(timestamp, timezone).Error, "It should catch an exception in CreateFromTimestampNano()")
}

func TestParseWithLayouts(t *testing.T) {
	t.Run("empty time", func(t *testing.T) {
		assert.Empty(t, ParseWithLayouts("", []string{DateTimeLayout}).ToString())
	})

	t.Run("invalid time", func(t *testing.T) {
		assert.True(t, ParseWithLayouts("0", []string{DateTimeLayout}).HasError())
		assert.True(t, ParseWithLayouts("xxx", []string{DateTimeLayout}).HasError())

		assert.Empty(t, ParseWithLayouts("0", []string{DateTimeLayout}).ToString())
		assert.Empty(t, ParseWithLayouts("xxx", []string{DateTimeLayout}).ToString())
	})

	t.Run("empty layouts", func(t *testing.T) {
		assert.Equal(t, "2020-08-05 13:14:15 +0000 UTC", ParseWithLayouts("2020-08-05 13:14:15", []string{}).ToString())
		assert.Equal(t, "2006-01-02 15:04:05", ParseWithLayouts("2020-08-05 13:14:15", []string{}).CurrentLayout())
	})

	t.Run("invalid timezone", func(t *testing.T) {
		assert.True(t, ParseWithLayouts("2020-08-05 13:14:15", []string{DateLayout}, "").HasError())
		assert.True(t, ParseWithLayouts("2020-08-05 13:14:15", []string{DateLayout}, "xxx").HasError())

		assert.Empty(t, ParseWithLayouts("2020-08-05 13:14:15", []string{DateLayout}, "").ToString())
		assert.Empty(t, ParseWithLayouts("2020-08-05 13:14:15", []string{DateLayout}, "xxx").ToString())
	})

	t.Run("without timezone", func(t *testing.T) {
		c := ParseWithLayouts("2020|08|05 13|14|15", []string{"2006|01|02 15|04|05", "2006|1|2 3|4|5"})
		assert.Equal(t, "2020-08-05 13:14:15 +0000 UTC", c.ToString())
		assert.Equal(t, "2006|01|02 15|04|05", c.CurrentLayout())
	})

	t.Run("with timezone", func(t *testing.T) {
		c := ParseWithLayouts("2020|08|05 13|14|15", []string{"2006|01|02 15|04|05", "2006|1|2 3|4|5"}, PRC)
		assert.Equal(t, "2020-08-05 13:14:15 +0800 CST", c.ToString())
		assert.Equal(t, "2006|01|02 15|04|05", c.CurrentLayout())
	})
}

func TestParseWithFormats(t *testing.T) {
	t.Run("empty time", func(t *testing.T) {
		assert.Empty(t, ParseWithFormats("", []string{DateTimeFormat}).ToString())
	})

	t.Run("invalid time", func(t *testing.T) {
		assert.True(t, ParseWithFormats("0", []string{DateTimeLayout}).HasError())
		assert.True(t, ParseWithFormats("xxx", []string{DateTimeLayout}).HasError())

		assert.Empty(t, ParseWithFormats("0", []string{DateTimeLayout}).ToString())
		assert.Empty(t, ParseWithFormats("xxx", []string{DateTimeLayout}).ToString())
	})

	t.Run("empty layouts", func(t *testing.T) {
		assert.Equal(t, "2020-08-05 13:14:15 +0000 UTC", ParseWithFormats("2020-08-05 13:14:15", []string{}).ToString())
		assert.Equal(t, "2006-01-02 15:04:05", ParseWithFormats("2020-08-05 13:14:15", []string{}).CurrentLayout())
	})

	t.Run("invalid timezone", func(t *testing.T) {
		assert.True(t, ParseWithFormats("2020-08-05 13:14:15", []string{DateFormat}, "").HasError())
		assert.True(t, ParseWithFormats("2020-08-05 13:14:15", []string{DateFormat}, "xxx").HasError())

		assert.Empty(t, ParseWithFormats("2020-08-05 13:14:15", []string{DateFormat}, "").ToString())
		assert.Empty(t, ParseWithFormats("2020-08-05 13:14:15", []string{DateFormat}, "xxx").ToString())
	})

	t.Run("without timezone", func(t *testing.T) {
		c := ParseWithFormats("2020|08|05 13|14|15", []string{"Y|m|d H|i|s", "y|m|d h|i|s"})
		assert.Equal(t, "2020-08-05 13:14:15 +0000 UTC", c.ToString())
		assert.Equal(t, "2006|01|02 15|04|05", c.CurrentLayout())
	})

	t.Run("with timezone", func(t *testing.T) {
		c := ParseWithFormats("2020|08|05 13|14|15", []string{"Y|m|d H|i|s", "y|m|d h|i|s"}, PRC)
		assert.Equal(t, "2020-08-05 13:14:15 +0800 CST", c.ToString())
		assert.Equal(t, "2006|01|02 15|04|05", c.CurrentLayout())
	})
}
