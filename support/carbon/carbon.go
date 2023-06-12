package carbon

import (
	stdtime "time"

	"github.com/golang-module/carbon/v2"
)

type Carbon = carbon.Carbon

// timezone constants
// 时区常量
const (
	Local = carbon.Local // 本地时间
	UTC   = carbon.UTC   // 世界协调时间
	GMT   = carbon.GMT   // 格林尼治标准时间
	EET   = carbon.EET   // 欧洲东部标准时间
	WET   = carbon.WET   // 欧洲西部标准时间
	CET   = carbon.CET   // 欧洲中部标准时间
	EST   = carbon.EST   // 美国东部标准时间
	MST   = carbon.MST   // 美国山地标准时间

	Cuba      = carbon.Cuba      // 古巴
	Egypt     = carbon.Egypt     // 埃及
	Eire      = carbon.Eire      // 爱尔兰
	Greenwich = carbon.Greenwich // 格林尼治
	Iceland   = carbon.Iceland   // 冰岛
	Iran      = carbon.Iran      // 伊朗
	Israel    = carbon.Israel    // 以色列
	Jamaica   = carbon.Jamaica   // 牙买加
	Japan     = carbon.Japan     // 日本
	Libya     = carbon.Libya     // 利比亚
	Poland    = carbon.Poland    // 波兰
	Portugal  = carbon.Portugal  // 葡萄牙
	PRC       = carbon.PRC       // 中国
	Singapore = carbon.Singapore // 新加坡
	Turkey    = carbon.Turkey    // 土耳其

	Shanghai   = carbon.Shanghai   // 上海
	Chongqing  = carbon.Chongqing  // 重庆
	Harbin     = carbon.Harbin     // 哈尔滨
	Urumqi     = carbon.Urumqi     // 乌鲁木齐
	HongKong   = carbon.HongKong   // 香港
	Macao      = carbon.Macao      // 澳门
	Taipei     = carbon.Taipei     // 台北
	Tokyo      = carbon.Tokyo      // 东京
	Saigon     = carbon.Saigon     // 西贡
	Seoul      = carbon.Seoul      // 首尔
	Bangkok    = carbon.Bangkok    // 曼谷
	Dubai      = carbon.Dubai      // 迪拜
	NewYork    = carbon.NewYork    // 纽约
	LosAngeles = carbon.LosAngeles // 洛杉矶
	Chicago    = carbon.Chicago    // 芝加哥
	Moscow     = carbon.Moscow     // 莫斯科
	London     = carbon.London     // 伦敦
	Berlin     = carbon.Berlin     // 柏林
	Paris      = carbon.Paris      // 巴黎
	Rome       = carbon.Rome       // 罗马
	Sydney     = carbon.Sydney     // 悉尼
	Melbourne  = carbon.Melbourne  // 墨尔本
	Darwin     = carbon.Darwin     // 达尔文
)

// month constants
// 月份常量
const (
	January   = carbon.January   // 一月
	February  = carbon.February  // 二月
	March     = carbon.March     // 三月
	April     = carbon.April     // 四月
	May       = carbon.May       // 五月
	June      = carbon.June      // 六月
	July      = carbon.July      // 七月
	August    = carbon.August    // 八月
	September = carbon.September // 九月
	October   = carbon.October   // 十月
	November  = carbon.November  // 十一月
	December  = carbon.December  // 十二月
)

// week constants
// 星期常量
const (
	Monday    = carbon.Monday    // 周一
	Tuesday   = carbon.Tuesday   // 周二
	Wednesday = carbon.Wednesday // 周三
	Thursday  = carbon.Thursday  // 周四
	Friday    = carbon.Friday    // 周五
	Saturday  = carbon.Saturday  // 周六
	Sunday    = carbon.Sunday    // 周日
)

// number constants
// 数字常量
const (
	YearsPerMillennium = 1000   // 每千年1000年
	YearsPerCentury    = 100    // 每世纪100年
	YearsPerDecade     = 10     // 每十年10年
	QuartersPerYear    = 4      // 每年4个季度
	MonthsPerYear      = 12     // 每年12月
	MonthsPerQuarter   = 3      // 每季度3月
	WeeksPerNormalYear = 52     // 每常规年52周
	weeksPerLongYear   = 53     // 每长年53周
	WeeksPerMonth      = 4      // 每月4周
	DaysPerLeapYear    = 366    // 每闰年366天
	DaysPerNormalYear  = 365    // 每常规年365天
	DaysPerWeek        = 7      // 每周7天
	HoursPerWeek       = 168    // 每周168小时
	HoursPerDay        = 24     // 每天24小时
	MinutesPerDay      = 1440   // 每天1440分钟
	MinutesPerHour     = 60     // 每小时60分钟
	SecondsPerWeek     = 604800 // 每周604800秒
	SecondsPerDay      = 86400  // 每天86400秒
	SecondsPerHour     = 3600   // 每小时3600秒
	SecondsPerMinute   = 60     // 每分钟60秒
)

// layout constants
// 布局模板常量
const (
	ANSICLayout              = stdtime.ANSIC
	UnixDateLayout           = stdtime.UnixDate
	RubyDateLayout           = stdtime.RubyDate
	RFC822Layout             = stdtime.RFC822
	RFC822ZLayout            = stdtime.RFC822Z
	RFC850Layout             = stdtime.RFC850
	RFC1123Layout            = stdtime.RFC1123
	RFC1123ZLayout           = stdtime.RFC1123Z
	RssLayout                = stdtime.RFC1123Z
	KitchenLayout            = stdtime.Kitchen
	RFC2822Layout            = stdtime.RFC1123Z
	CookieLayout             = carbon.CookieLayout
	RFC3339Layout            = carbon.RFC3339Layout
	RFC3339MilliLayout       = carbon.RFC3339MilliLayout
	RFC3339MicroLayout       = carbon.RFC3339MicroLayout
	RFC3339NanoLayout        = carbon.RFC3339NanoLayout
	ISO8601Layout            = carbon.ISO8601Layout
	ISO8601MilliLayout       = carbon.ISO8601MilliLayout
	ISO8601MicroLayout       = carbon.ISO8601MicroLayout
	ISO8601NanoLayout        = carbon.ISO8601NanoLayout
	RFC1036Layout            = carbon.RFC1036Layout
	RFC7231Layout            = carbon.RFC7231Layout
	DayDateTimeLayout        = carbon.DayDateTimeLayout
	DateTimeLayout           = carbon.DateTimeLayout
	DateTimeMilliLayout      = carbon.DateTimeMilliLayout
	DateTimeMicroLayout      = carbon.DateTimeMicroLayout
	DateTimeNanoLayout       = carbon.DateTimeNanoLayout
	ShortDateTimeLayout      = carbon.ShortDateTimeLayout
	ShortDateTimeMilliLayout = carbon.ShortDateTimeMilliLayout
	ShortDateTimeMicroLayout = carbon.ShortDateTimeMicroLayout
	ShortDateTimeNanoLayout  = carbon.ShortDateTimeNanoLayout
	DateLayout               = carbon.DateLayout
	DateMilliLayout          = carbon.DateMilliLayout
	DateMicroLayout          = carbon.DateMicroLayout
	DateNanoLayout           = carbon.DateNanoLayout
	ShortDateLayout          = carbon.ShortDateLayout
	ShortDateMilliLayout     = carbon.ShortDateMilliLayout
	ShortDateMicroLayout     = carbon.ShortDateMicroLayout
	ShortDateNanoLayout      = carbon.ShortDateNanoLayout
	TimeLayout               = carbon.TimeLayout
	TimeMilliLayout          = carbon.TimeMilliLayout
	TimeMicroLayout          = carbon.TimeMicroLayout
	TimeNanoLayout           = carbon.TimeNanoLayout
	ShortTimeLayout          = carbon.ShortTimeLayout
	ShortTimeMilliLayout     = carbon.ShortTimeMilliLayout
	ShortTimeMicroLayout     = carbon.ShortTimeMicroLayout
	ShortTimeNanoLayout      = carbon.ShortTimeNanoLayout
)

type Clock struct {
	testNow  bool
	testTime Carbon
	timezone string
}

var clock = &Clock{}

// SetTestNow Set the test time. Remember to unset after used.
func SetTestNow(testTime Carbon) {
	clock.testNow = true
	clock.testTime = testTime
}

// UnsetTestNow Unset the test time.
func UnsetTestNow() {
	clock.testNow = false
}

// IsTestNow Determine if the test now time is set.
func IsTestNow() bool {
	return clock.testNow
}

// SetTimezone sets timezone.
// 设置时区
func SetTimezone(timezone string) {
	clock.timezone = timezone
}

// Now return a Carbon object of now.
func Now(timezone ...string) Carbon {
	if IsTestNow() {
		return clock.testTime
	}

	return carbon.Now(getTimezone(timezone))
}

// Parse return a Carbon object of given value.
func Parse(value string, timezone ...string) Carbon {
	return carbon.Parse(value, getTimezone(timezone))
}

// ParseByFormat return a Carbon object of given value and format.
func ParseByFormat(value, format string, timezone ...string) Carbon {
	return carbon.ParseByFormat(value, format, getTimezone(timezone))
}

// ParseByLayout return a Carbon object of given value and layout.
func ParseByLayout(value, layout string, timezone ...string) Carbon {
	return carbon.ParseByLayout(value, layout, getTimezone(timezone))
}

// FromTimestamp return a Carbon object of given timestamp.
func FromTimestamp(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestamp(timestamp, getTimezone(timezone))
}

// FromTimestampMilli return a Carbon object of given millisecond timestamp.
func FromTimestampMilli(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestampMilli(timestamp, getTimezone(timezone))
}

// FromTimestampMicro return a Carbon object of given microsecond timestamp.
func FromTimestampMicro(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestampMicro(timestamp, getTimezone(timezone))
}

// FromTimestampNano return a Carbon object of given nanosecond timestamp.
func FromTimestampNano(timestamp int64, timezone ...string) Carbon {
	return carbon.CreateFromTimestampNano(timestamp, getTimezone(timezone))
}

// FromDateTime return a Carbon object of given date and time.
func FromDateTime(year int, month int, day int, hour int, minute int, second int, timezone ...string) Carbon {
	return carbon.CreateFromDateTime(year, month, day, hour, minute, second, getTimezone(timezone))
}

// FromDateTimeMilli return a Carbon object of given date and millisecond time.
func FromDateTimeMilli(year int, month int, day int, hour int, minute int, second int, millisecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateTimeMilli(year, month, day, hour, minute, second, millisecond, getTimezone(timezone))
}

// FromDateTimeMicro return a Carbon object of given date and microsecond time.
func FromDateTimeMicro(year int, month int, day int, hour int, minute int, second int, microsecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateTimeMicro(year, month, day, hour, minute, second, microsecond, getTimezone(timezone))
}

// FromDateTimeNano return a Carbon object of given date and nanosecond time.
func FromDateTimeNano(year int, month int, day int, hour int, minute int, second int, nanosecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateTimeNano(year, month, day, hour, minute, second, nanosecond, getTimezone(timezone))
}

// FromDate return a Carbon object of given date.
func FromDate(year int, month int, day int, timezone ...string) Carbon {
	return carbon.CreateFromDate(year, month, day, getTimezone(timezone))
}

// FromDateMilli return a Carbon object of given millisecond date.
func FromDateMilli(year int, month int, day int, millisecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateMilli(year, month, day, millisecond, getTimezone(timezone))
}

// FromDateMicro return a Carbon object of given microsecond date.
func FromDateMicro(year int, month int, day int, microsecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateMicro(year, month, day, microsecond, getTimezone(timezone))
}

// FromDateNano return a Carbon object of given nanosecond date.
func FromDateNano(year int, month int, day int, nanosecond int, timezone ...string) Carbon {
	return carbon.CreateFromDateNano(year, month, day, nanosecond, getTimezone(timezone))
}

// FromTime return a Carbon object of given time.
func FromTime(hour int, minute int, second int, timezone ...string) Carbon {
	return carbon.CreateFromTime(hour, minute, second, getTimezone(timezone))
}

// FromTimeMilli return a Carbon object of given millisecond time.
func FromTimeMilli(hour int, minute int, second int, millisecond int, timezone ...string) Carbon {
	return carbon.CreateFromTimeMilli(hour, minute, second, millisecond, getTimezone(timezone))
}

// FromTimeMicro return a Carbon object of given microsecond time.
func FromTimeMicro(hour int, minute int, second int, microsecond int, timezone ...string) Carbon {
	return carbon.CreateFromTimeMicro(hour, minute, second, microsecond, getTimezone(timezone))
}

// FromTimeNano return a Carbon object of given nanosecond time.
func FromTimeNano(hour int, minute int, second int, nanosecond int, timezone ...string) Carbon {
	return carbon.CreateFromTimeNano(hour, minute, second, nanosecond, getTimezone(timezone))
}

// FromStdTime return a Carbon object of given time.Time object.
func FromStdTime(time stdtime.Time) Carbon {
	return carbon.FromStdTime(time)
}

func getTimezone(timezone []string) string {
	if len(timezone) == 0 {
		if clock.timezone == "" {
			return UTC
		} else {
			return clock.timezone
		}
	}

	return timezone[0]
}
