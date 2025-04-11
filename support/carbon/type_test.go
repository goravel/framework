package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	DateTime      DateTime      `json:"date_time"`
	DateTimeMilli DateTimeMilli `json:"date_time_milli"`
	DateTimeMicro DateTimeMicro `json:"date_time_micro"`
	DateTimeNano  DateTimeNano  `json:"date_time_nano"`

	Date      Date      `json:"date"`
	DateMilli DateMilli `json:"date_milli"`
	DateMicro DateMicro `json:"date_micro"`
	DateNano  DateNano  `json:"date_nano"`

	Time      Time      `json:"time"`
	TimeMilli TimeMilli `json:"time_milli"`
	TimeMicro TimeMicro `json:"time_micro"`
	TimeNano  TimeNano  `json:"time_nano"`

	Timestamp      Timestamp      `json:"timestamp"`
	TimestampMilli TimestampMilli `json:"timestamp_milli"`
	TimestampMicro TimestampMicro `json:"timestamp_micro"`
	TimestampNano  TimestampNano  `json:"timestamp_nano"`
}

func TestType_MarshalJSON(t *testing.T) {
	var user User

	c := Parse("2020-08-05 13:14:15.999999999")

	user.DateTime = NewDateTime(c)
	user.DateTimeMilli = NewDateTimeMilli(c)
	user.DateTimeMicro = NewDateTimeMicro(c)
	user.DateTimeNano = NewDateTimeNano(c)

	user.Date = NewDate(c)
	user.DateMilli = NewDateMilli(c)
	user.DateMicro = NewDateMicro(c)
	user.DateNano = NewDateNano(c)

	user.Time = NewTime(c)
	user.TimeMilli = NewTimeMilli(c)
	user.TimeMicro = NewTimeMicro(c)
	user.TimeNano = NewTimeNano(c)

	user.Timestamp = NewTimestamp(c)
	user.TimestampMilli = NewTimestampMilli(c)
	user.TimestampMicro = NewTimestampMicro(c)
	user.TimestampNano = NewTimestampNano(c)

	data, err := json.Marshal(&user)
	assert.NoError(t, err)
	assert.Equal(t, `{"date_time":"2020-08-05 13:14:15","date_time_milli":"2020-08-05 13:14:15.999","date_time_micro":"2020-08-05 13:14:15.999999","date_time_nano":"2020-08-05 13:14:15.999999999","date":"2020-08-05","date_milli":"2020-08-05.999","date_micro":"2020-08-05.999999","date_nano":"2020-08-05.999999999","time":"13:14:15","time_milli":"13:14:15.999","time_micro":"13:14:15.999999","time_nano":"13:14:15.999999999","timestamp":1596633255,"timestamp_milli":1596633255999,"timestamp_micro":1596633255999999,"timestamp_nano":1596633255999999999}`, string(data))
}

func TestType_UnmarshalJSON(t *testing.T) {
	var user User

	value := `{"date_time":"2020-08-05 13:14:15","date_time_milli":"2020-08-05 13:14:15.999","date_time_micro":"2020-08-05 13:14:15.999999","date_time_nano":"2020-08-05 13:14:15.999999999","date":"2020-08-05","date_milli":"2020-08-05.999","date_micro":"2020-08-05.999999","date_nano":"2020-08-05.999999999","time":"13:14:15","time_milli":"13:14:15.999","time_micro":"13:14:15.999999","time_nano":"13:14:15.999999999","timestamp":1596633255,"timestamp_milli":1596633255999,"timestamp_micro":1596633255999999,"timestamp_nano":1596633255999999999}`
	assert.NoError(t, json.Unmarshal([]byte(value), &user))

	assert.Equal(t, "2020-08-05 13:14:15", user.DateTime.String())
	assert.Equal(t, "2020-08-05 13:14:15.999", user.DateTimeMilli.String())
	assert.Equal(t, "2020-08-05 13:14:15.999999", user.DateTimeMicro.String())
	assert.Equal(t, "2020-08-05 13:14:15.999999999", user.DateTimeNano.String())
	assert.Equal(t, "2020-08-05", user.Date.String())
	assert.Equal(t, "2020-08-05.999", user.DateMilli.String())
	assert.Equal(t, "2020-08-05.999999", user.DateMicro.String())
	assert.Equal(t, "2020-08-05.999999999", user.DateNano.String())
	assert.Equal(t, "13:14:15", user.Time.String())
	assert.Equal(t, "13:14:15.999", user.TimeMilli.String())
	assert.Equal(t, "13:14:15.999999", user.TimeMicro.String())
	assert.Equal(t, "13:14:15.999999999", user.TimeNano.String())
	assert.Equal(t, int64(1596633255), user.Timestamp.Int64())
	assert.Equal(t, int64(1596633255999), user.TimestampMilli.Int64())
	assert.Equal(t, int64(1596633255999999), user.TimestampMicro.Int64())
	assert.Equal(t, int64(1596633255999999999), user.TimestampNano.Int64())
}
