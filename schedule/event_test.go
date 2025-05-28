package schedule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type EventTestSuite struct {
	suite.Suite
	event *Event
}

func TestEventTestSuite(t *testing.T) {
	suite.Run(t, new(EventTestSuite))
}

func (s *EventTestSuite) SetupTest() {
	s.event = &Event{cron: "* * * * *"}
}

func (s *EventTestSuite) TestAt() {
	s.Equal("30 10 * * *", s.event.At("10:30").GetCron())
}

func (s *EventTestSuite) TestCron() {
	s.Equal("* * * * 1", s.event.Cron("* * * * 1").GetCron())
}

func (s *EventTestSuite) TestDaily() {
	s.Equal("0 0 * * *", s.event.Daily().GetCron())
}

func (s *EventTestSuite) TestDailyAt() {
	s.Equal("30 10 * * *", s.event.DailyAt("10:30").GetCron())
	s.Equal("0 10 * * *", s.event.DailyAt("10").GetCron())
}

func (s *EventTestSuite) TestTwiceDaily() {
	s.Equal("* 1,13 * * *", s.event.TwiceDaily().GetCron())
	s.Equal("* 5,15 * * *", s.event.TwiceDaily(5, 15).GetCron())
}

func (s *EventTestSuite) TestEverySecond() {
	s.Equal("* * * * * *", s.event.EverySecond().GetCron())
}

func (s *EventTestSuite) TestEveryTwoSeconds() {
	s.Equal("*/2 * * * * *", s.event.EveryTwoSeconds().GetCron())
}

func (s *EventTestSuite) TestEveryFiveSeconds() {
	s.Equal("*/5 * * * * *", s.event.EveryFiveSeconds().GetCron())
}

func (s *EventTestSuite) TestEveryTenSeconds() {
	s.Equal("*/10 * * * * *", s.event.EveryTenSeconds().GetCron())
}

func (s *EventTestSuite) TestEveryFifteenSeconds() {
	s.Equal("*/15 * * * * *", s.event.EveryFifteenSeconds().GetCron())
}

func (s *EventTestSuite) TestEveryTwentySeconds() {
	s.Equal("*/20 * * * * *", s.event.EveryTwentySeconds().GetCron())
}

func (s *EventTestSuite) TestEveryThirtySeconds() {
	s.Equal("*/30 * * * * *", s.event.EveryThirtySeconds().GetCron())
}

func (s *EventTestSuite) TestEveryMinute() {
	s.Equal("* * * * *", s.event.EveryMinute().GetCron())
}

func (s *EventTestSuite) TestEveryTwoMinutes() {
	s.Equal("*/2 * * * *", s.event.EveryTwoMinutes().GetCron())
}

func (s *EventTestSuite) TestEveryThreeMinutes() {
	s.Equal("*/3 * * * *", s.event.EveryThreeMinutes().GetCron())
}

func (s *EventTestSuite) TestEveryFourMinutes() {
	s.Equal("*/4 * * * *", s.event.EveryFourMinutes().GetCron())
}

func (s *EventTestSuite) TestEveryFiveMinutes() {
	s.Equal("*/5 * * * *", s.event.EveryFiveMinutes().GetCron())
}

func (s *EventTestSuite) TestEveryTenMinutes() {
	s.Equal("*/10 * * * *", s.event.EveryTenMinutes().GetCron())
}

func (s *EventTestSuite) TestEveryFifteenMinutes() {
	s.Equal("*/15 * * * *", s.event.EveryFifteenMinutes().GetCron())
}

func (s *EventTestSuite) TestEveryThirtyMinutes() {
	s.Equal("*/30 * * * *", s.event.EveryThirtyMinutes().GetCron())
}

func (s *EventTestSuite) TestEveryTwoHours() {
	s.Equal("0 */2 * * *", s.event.EveryTwoHours().GetCron())
}

func (s *EventTestSuite) TestEveryThreeHours() {
	s.Equal("0 */3 * * *", s.event.EveryThreeHours().GetCron())
}

func (s *EventTestSuite) TestEveryFourHours() {
	s.Equal("0 */4 * * *", s.event.EveryFourHours().GetCron())
}

func (s *EventTestSuite) TestEverySixHours() {
	s.Equal("0 */6 * * *", s.event.EverySixHours().GetCron())
}

func (s *EventTestSuite) TestGetCron() {
	s.Equal("* * * * *", s.event.GetCron())

	s.event.cron = "* * * * 1"
	s.Equal("* * * * 1", s.event.GetCron())
}

func (s *EventTestSuite) TestHourly() {
	s.Equal("0 * * * *", s.event.Hourly().GetCron())
}

func (s *EventTestSuite) TestHourlyAt() {
	s.Equal("10,20 * * * *", s.event.HourlyAt([]string{"10", "20"}).GetCron())
}

func (s *EventTestSuite) TestDays() {
	s.Equal("* * * * 1,5", s.event.Days(time.Monday, time.Friday).GetCron())
	s.Equal("* * * * 0-1,5-6", s.event.Days(time.Sunday, time.Monday, time.Friday, time.Saturday).GetCron())
}

func (s *EventTestSuite) TestWeekdays() {
	s.Equal("* * * * 1-5", s.event.Weekdays().GetCron())
}

func (s *EventTestSuite) TestWeekends() {
	s.Equal("* * * * 0,6", s.event.Weekends().GetCron())
}

func (s *EventTestSuite) TestMondays() {
	s.Equal("* * * * 1", s.event.Mondays().GetCron())
}

func (s *EventTestSuite) TestTuesdays() {
	s.Equal("* * * * 2", s.event.Tuesdays().GetCron())
}

func (s *EventTestSuite) TestWednesdays() {
	s.Equal("* * * * 3", s.event.Wednesdays().GetCron())
}

func (s *EventTestSuite) TestThursdays() {
	s.Equal("* * * * 4", s.event.Thursdays().GetCron())
}

func (s *EventTestSuite) TestFridays() {
	s.Equal("* * * * 5", s.event.Fridays().GetCron())
}

func (s *EventTestSuite) TestSaturdays() {
	s.Equal("* * * * 6", s.event.Saturdays().GetCron())
}

func (s *EventTestSuite) TestSundays() {
	s.Equal("* * * * 0", s.event.Sundays().GetCron())
}

func (s *EventTestSuite) TestWeekly() {
	s.Equal("0 0 * * 0", s.event.Weekly().GetCron())
}

func (s *EventTestSuite) TestMonthly() {
	s.Equal("0 0 1 * *", s.event.Monthly().GetCron())
}

func (s *EventTestSuite) TestQuarterly() {
	s.Equal("0 0 1 1-12/3 *", s.event.Quarterly().GetCron())
}

func (s *EventTestSuite) TestYearly() {
	s.Equal("0 0 1 1 *", s.event.Yearly().GetCron())
}

func (s *EventTestSuite) TestSkipIfStillRunning() {
	s.event.SkipIfStillRunning()
	s.True(s.event.GetSkipIfStillRunning())
}

func (s *EventTestSuite) TestDelayIfStillRunning() {
	s.event.DelayIfStillRunning()
	s.True(s.event.GetDelayIfStillRunning())
}
