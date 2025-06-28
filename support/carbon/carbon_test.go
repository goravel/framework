package carbon

import (
	"testing"
	stdtime "time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tt, _ := stdtime.ParseInLocation(DateTimeLayout, "2020-08-05 13:14:15", stdtime.UTC)
	assert.Equal(t, "2020-08-05 13:14:15 +0000 UTC", New(tt).ToString())
	assert.Equal(t, tt.String(), New(tt).ToString())
}

func TestZeroValue(t *testing.T) {
	assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", ZeroValue().ToString())
	assert.Equal(t, "January", ZeroValue().ToMonthString())
	assert.Equal(t, "一月", ZeroValue().SetLocale("zh-CN").ToMonthString())
}

func TestEpochValue(t *testing.T) {
	assert.Equal(t, "1970-01-01 00:00:00 +0000 UTC", EpochValue().ToString())
	assert.Equal(t, "January", EpochValue().ToMonthString())
	assert.Equal(t, "一月", EpochValue().SetLocale("zh-CN").ToMonthString())
}

func TestSetTimezone(t *testing.T) {
	defer SetTimezone(UTC)

	SetTimezone(PRC)
	c := Parse("2025-04-11 00:00:00")

	assert.Equal(t, PRC, c.Timezone())
	assert.Equal(t, "CST", c.ZoneName())
	assert.Equal(t, 28800, c.ZoneOffset())
	assert.Equal(t, "2025-04-11 00:00:00 +0800 CST", c.ToString())
}

func TestSetLocale(t *testing.T) {
	defer SetLocale("en")

	SetLocale("zh-CN")
	c := Parse("2025-04-11 00:00:00")

	assert.Equal(t, "zh-CN", c.Locale())
	assert.Equal(t, "白羊座", c.Constellation())
	assert.Equal(t, "春季", c.Season())
	assert.Equal(t, "四月", c.ToMonthString())
	assert.Equal(t, "4月", c.ToShortMonthString())
	assert.Equal(t, "星期五", c.ToWeekString())
	assert.Equal(t, "周五", c.ToShortWeekString())
}

func TestSetTestNow(t *testing.T) {
	SetTestNow(Now().SubHour())
	assert.True(t, IsTestNow())
	ClearTestNow()
	assert.False(t, IsTestNow())
}

func TestClearTestNow(t *testing.T) {
	now := Parse("2020-08-05")

	SetTestNow(now)
	assert.Equal(t, "2020-08-05", Now().ToDateString())
	assert.True(t, IsTestNow())

	ClearTestNow()
	assert.Equal(t, stdtime.Now().In(stdtime.UTC).Format(DateTimeLayout), Now().ToDateTimeString())
	assert.False(t, IsTestNow())
}

func TestUnsetTestNow(t *testing.T) {
	now := Parse("2020-08-05")

	SetTestNow(now)
	assert.Equal(t, "2020-08-05", Now().ToDateString())
	assert.True(t, IsTestNow())

	UnsetTestNow()
	assert.Equal(t, stdtime.Now().In(stdtime.UTC).Format(DateTimeLayout), Now().ToDateTimeString())
	assert.False(t, IsTestNow())
}

func TestNow(t *testing.T) {
	SetTestNow(Now().SubSeconds(10))
	stdtime.Sleep(2 * stdtime.Second)
	testNow := Now().Timestamp()
	ClearTestNow()
	now := Now().Timestamp()
	assert.True(t, now-testNow >= 10)

	utcNow := Now()
	shanghaiNow := Now(Shanghai)
	assert.Equal(t, shanghaiNow.SubHours(8).ToDateTimeString(), utcNow.ToDateTimeString())
}

func TestParse(t *testing.T) {
	time := Parse("2020-01-01 00:00:00", UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
}

func TestParseByLayout(t *testing.T) {
	t.Run("string type layout", func(t *testing.T) {
		time := ParseByLayout("2020-01-01 00:00:00", "2006-01-02 15:04:05", UTC)
		assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
	})

	t.Run("[]string type layout", func(t *testing.T) {
		c := ParseByLayout("2020|08|05 13|14|15", []string{"2006|01|02 15|04|05", "2006|1|2 3|4|5"}, PRC)
		assert.Equal(t, "2020-08-05 13:14:15 +0800 CST", c.ToString())
		assert.Equal(t, "2006|01|02 15|04|05", c.CurrentLayout())
	})
}

func TestParseByFormat(t *testing.T) {
	t.Run("string type layout", func(t *testing.T) {
		time := ParseByFormat("2020-01-01 00:00:00", "Y-m-d H:i:s", UTC)
		assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
	})

	t.Run("[]string type layout", func(t *testing.T) {
		c := ParseByFormat("2020|08|05 13|14|15", []string{"Y|m|d H|i|s", "y|m|d h|i|s"}, PRC)
		assert.Equal(t, "2020-08-05 13:14:15 +0800 CST", c.ToString())
		assert.Equal(t, "2006|01|02 15|04|05", c.CurrentLayout())
	})
}

func TestParseWithLayouts(t *testing.T) {
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

func TestFromTimestamp(t *testing.T) {
	time := FromTimestamp(1577836800, UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
}

func TestFromTimestampMilli(t *testing.T) {
	time := FromTimestampMilli(1577836800000, UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
}

func TestFromTimestampMicro(t *testing.T) {
	time := FromTimestampMicro(1577836800000000, UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
}

func TestFromTimestampNano(t *testing.T) {
	time := FromTimestampNano(1577836800000000000, UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
}

func TestFromDateTime(t *testing.T) {
	time := FromDateTime(2020, 1, 1, 0, 0, 0, UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
}

func TestFromDateTimeMilli(t *testing.T) {
	time := FromDateTimeMilli(2020, 1, 1, 0, 0, 0, 0, UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
}

func TestFromDateTimeMicro(t *testing.T) {
	time := FromDateTimeMicro(2020, 1, 1, 0, 0, 0, 0, UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
}

func TestFromDateTimeNano(t *testing.T) {
	time := FromDateTimeNano(2020, 1, 1, 0, 0, 0, 0, UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
}

func TestFromDate(t *testing.T) {
	time := FromDate(2020, 1, 1, UTC)
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
}

func TestFromDateMilli(t *testing.T) {
	time := FromDateMilli(2020, 1, 1, 999, UTC)
	assert.Equal(t, "2020-01-01 00:00:00.999", time.ToDateTimeMilliString(UTC))
}

func TestFromDateMicro(t *testing.T) {
	time := FromDateMicro(2020, 1, 1, 999999, UTC)
	assert.Equal(t, "2020-01-01 00:00:00.999999", time.ToDateTimeMicroString(UTC))
}

func TestFromDateNano(t *testing.T) {
	time := FromDateNano(2020, 1, 1, 999999999, UTC)
	assert.Equal(t, "2020-01-01 00:00:00.999999999", time.ToDateTimeNanoString(UTC))
}

func TestFromTime(t *testing.T) {
	date := Now().ToDateString(UTC)
	time := FromTime(0, 0, 0, UTC)
	assert.Equal(t, date+" 00:00:00", time.ToDateTimeString(UTC))
}

func TestFromTimeMilli(t *testing.T) {
	date := Now().ToDateString(UTC)
	time := FromTimeMilli(0, 0, 0, 999, UTC)
	assert.Equal(t, date+" 00:00:00.999", time.ToDateTimeMilliString(UTC))
}

func TestFromTimeMicro(t *testing.T) {
	date := Now().ToDateString(UTC)
	time := FromTimeMicro(0, 0, 0, 999999, UTC)
	assert.Equal(t, date+" 00:00:00.999999", time.ToDateTimeMicroString(UTC))
}

func TestFromTimeNano(t *testing.T) {
	date := Now().ToDateString(UTC)
	time := FromTimeNano(0, 0, 0, 999999999, UTC)
	assert.Equal(t, date+" 00:00:00.999999999", time.ToDateTimeNanoString(UTC))
}

func TestFromStdTime(t *testing.T) {
	time := FromStdTime(stdtime.Date(2020, 1, 1, 0, 0, 0, 0, stdtime.UTC))
	assert.Equal(t, "2020-01-01 00:00:00", time.ToDateTimeString(UTC))
}

func TestFromJulian(t *testing.T) {
	assert.Equal(t, "2024-01-23 13:14:15 +0000 UTC", FromJulian(2460333.051563).ToString())
	assert.Equal(t, "2024-01-23 13:14:15 +0000 UTC", FromJulian(60332.551563).ToString())

	assert.Equal(t, "2024-01-23 13:14:18 +0000 UTC", FromJulian(2460333.0516).ToString())
	assert.Equal(t, "2024-01-23 13:14:18 +0000 UTC", FromJulian(60332.5516).ToString())
}

func TestFromLunar(t *testing.T) {
	assert.Equal(t, "2024-01-21 00:00:00 +0800 CST", FromLunar(2023, 12, 11, false).ToString(PRC))
	assert.Equal(t, "2024-01-18 00:00:00 +0800 CST", FromLunar(2023, 12, 8, false).ToString(PRC))
	assert.Equal(t, "2024-01-24 00:00:00 +0800 CST", FromLunar(2023, 12, 14, false).ToString(PRC))
}

func TestFromPersian(t *testing.T) {
	assert.Equal(t, "1800-01-01 00:00:00", FromPersian(1178, 10, 11).ToDateTimeString())
	assert.Equal(t, "2024-01-01 00:00:00", FromPersian(1402, 10, 11).ToDateTimeString())
	assert.Equal(t, "2024-08-05 00:00:00", FromPersian(1403, 5, 15).ToDateTimeString())
}
