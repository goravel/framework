package schedule

import (
	"testing"

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
	s.event = &Event{}
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
	s.Equal("0,30 * * * *", s.event.EveryThirtyMinutes().GetCron())
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

func (s *EventTestSuite) TestSkipIfStillRunning() {
	s.event.SkipIfStillRunning()
	s.True(s.event.GetSkipIfStillRunning())
}

func (s *EventTestSuite) TestDelayIfStillRunning() {
	s.event.DelayIfStillRunning()
	s.True(s.event.GetDelayIfStillRunning())
}
