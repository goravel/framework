package carbon

import (
	"github.com/golang-module/carbon/v2"
	"testing"
	stdtime "time"

	"github.com/stretchr/testify/assert"
)

func TestSetTestNow(t *testing.T) {
	SetTestNow(Now().SubHour())
	assert.True(t, IsTestNow())
	UnsetTestNow()
	assert.False(t, IsTestNow())
}

func TestNow(t *testing.T) {
	SetTestNow(Now().SubSeconds(10))
	stdtime.Sleep(2 * stdtime.Second)
	testNow := Now().Timestamp()
	UnsetTestNow()
	now := Now().Timestamp()
	assert.True(t, now-testNow >= 10)
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
	clock := Now(carbon.UTC).ToTimeString(carbon.UTC)
	time := FromDate(2020, 1, 1, carbon.UTC)
	assert.Equal(t, "2020-01-01 "+clock, time.ToDateTimeString(carbon.UTC))
}

func TestFromDateMilli(t *testing.T) {
	clock := Now().ToTimeString(carbon.UTC)
	time := FromDateMilli(2020, 1, 1, 999, carbon.UTC)
	assert.Equal(t, "2020-01-01 "+clock+".999", time.ToDateTimeMilliString(carbon.UTC))
}

func TestFromDateMicro(t *testing.T) {
	clock := Now().ToTimeString(carbon.UTC)
	time := FromDateMicro(2020, 1, 1, 999999, carbon.UTC)
	assert.Equal(t, "2020-01-01 "+clock+".999999", time.ToDateTimeMicroString(carbon.UTC))
}

func TestFromDateNano(t *testing.T) {
	clock := Now().ToTimeString(carbon.UTC)
	time := FromDateNano(2020, 1, 1, 999999999, carbon.UTC)
	assert.Equal(t, "2020-01-01 "+clock+".999999999", time.ToDateTimeNanoString(carbon.UTC))
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
