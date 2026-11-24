package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mocksschedule "github.com/goravel/framework/mocks/schedule"
)

func TestCronSpec(t *testing.T) {
	event := mocksschedule.NewEvent(t)
	event.EXPECT().GetCron().Return("0 9 * * *").Twice()
	event.EXPECT().GetTimezone().Return("").Once()
	event.EXPECT().GetTimezone().Return("America/Santiago").Once()

	assert.Equal(t, "0 9 * * *", CronSpec(event))
	assert.Equal(t, "CRON_TZ=America/Santiago 0 9 * * *", CronSpec(event))
}
