package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	Carbon1 Carbon  `json:"carbon1"`
	Carbon2 *Carbon `json:"carbon2"`

	Date      Date      `json:"date"`
	DateMilli DateMilli `json:"date_milli"`
	DateMicro DateMicro `json:"date_micro"`
	DateNano  DateNano  `json:"date_nano"`

	Time      Time      `json:"time"`
	TimeMilli TimeMilli `json:"time_milli"`
	TimeMicro TimeMicro `json:"time_micro"`
	TimeNano  TimeNano  `json:"time_nano"`

	DateTime      DateTime      `json:"date_time"`
	DateTimeMilli DateTimeMilli `json:"date_time_milli"`
	DateTimeMicro DateTimeMicro `json:"date_time_micro"`
	DateTimeNano  DateTimeNano  `json:"date_time_nano"`

	Timestamp      Timestamp      `json:"timestamp"`
	TimestampMilli TimestampMilli `json:"timestamp_milli"`
	TimestampMicro TimestampMicro `json:"timestamp_micro"`
	TimestampNano  TimestampNano  `json:"timestamp_nano"`

	CreatedAt *DateTime  `json:"created_at"`
	UpdatedAt *DateTime  `json:"updated_at"`
	DeletedAt *Timestamp `json:"deleted_at"`
}

var user User

func TestMarshalJSON(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")

	user.Carbon1 = *c
	user.Carbon2 = c

	user.Date = *NewDate(c)
	user.DateMilli = *NewDateMilli(c)
	user.DateMicro = *NewDateMicro(c)
	user.DateNano = *NewDateNano(c)

	user.Time = *NewTime(c)
	user.TimeMilli = *NewTimeMilli(c)
	user.TimeMicro = *NewTimeMicro(c)
	user.TimeNano = *NewTimeNano(c)

	user.DateTime = *NewDateTime(c)
	user.DateTimeMilli = *NewDateTimeMilli(c)
	user.DateTimeMicro = *NewDateTimeMicro(c)
	user.DateTimeNano = *NewDateTimeNano(c)

	user.Timestamp = *NewTimestamp(c)
	user.TimestampMilli = *NewTimestampMilli(c)
	user.TimestampMicro = *NewTimestampMicro(c)
	user.TimestampNano = *NewTimestampNano(c)

	user.CreatedAt = NewDateTime(c)
	user.UpdatedAt = NewDateTime(c)
	user.DeletedAt = NewTimestamp(c)
	data, err := json.Marshal(&user)
	assert.Nil(t, err)
	assert.Equal(t, `{"carbon1":"2020-08-05 13:14:15","carbon2":"2020-08-05 13:14:15","date":"2020-08-05","date_milli":"2020-08-05.999","date_micro":"2020-08-05.999999","date_nano":"2020-08-05.999999999","time":"13:14:15","time_milli":"13:14:15.999","time_micro":"13:14:15.999999","time_nano":"13:14:15.999999999","date_time":"2020-08-05 13:14:15","date_time_milli":"2020-08-05 13:14:15.999","date_time_micro":"2020-08-05 13:14:15.999999","date_time_nano":"2020-08-05 13:14:15.999999999","timestamp":1596633255,"timestamp_milli":1596633255999,"timestamp_micro":1596633255999999,"timestamp_nano":1596633255999999999,"created_at":"2020-08-05 13:14:15","updated_at":"2020-08-05 13:14:15","deleted_at":1596633255}`, string(data))
}

func TestUnmarshalJSON(t *testing.T) {
	str := `{
		"carbon1":"2020-08-05 13:14:15",
		"carbon2":"2020-08-05 13:14:15",
		"date":"2020-08-05",
		"date_milli":"2020-08-05.999",
		"date_micro":"2020-08-05.999999",
		"date_nano":"2020-08-05.999999999",
		"time":"13:14:15",
		"time_milli":"13:14:15.999",
		"time_micro":"13:14:15.999999",
		"time_nano":"13:14:15.999999999",
		"date_time":"2020-08-05 13:14:15",
		"date_time_milli":"2020-08-05 13:14:15.999",
		"date_time_micro":"2020-08-05 13:14:15.999999",
		"date_time_nano":"2020-08-05 13:14:15.999999999",
		"timestamp":1596633255,
		"timestamp_milli":1596633255999,
		"timestamp_micro":1596633255999999,
		"timestamp_nano":1596633255999999999,
		"created_at":"2020-08-05 13:14:15",
		"updated_at":"2020-08-05 13:14:15",
		"deleted_at":1596633255
	}`

	assert.NoError(t, json.Unmarshal([]byte(str), &user))

	assert.Equal(t, "2020-08-05 13:14:15", user.Carbon1.String())
	assert.Equal(t, "2020-08-05 13:14:15", user.Carbon2.String())

	assert.Equal(t, "2020-08-05", user.Date.String())
	assert.Equal(t, "2020-08-05.999", user.DateMilli.String())
	assert.Equal(t, "2020-08-05.999999", user.DateMicro.String())
	assert.Equal(t, "2020-08-05.999999999", user.DateNano.String())

	assert.Equal(t, "13:14:15", user.Time.String())
	assert.Equal(t, "13:14:15.999", user.TimeMilli.String())
	assert.Equal(t, "13:14:15.999999", user.TimeMicro.String())
	assert.Equal(t, "13:14:15.999999999", user.TimeNano.String())

	assert.Equal(t, "2020-08-05 13:14:15", user.DateTime.String())
	assert.Equal(t, "2020-08-05 13:14:15.999", user.DateTimeMilli.String())
	assert.Equal(t, "2020-08-05 13:14:15.999999", user.DateTimeMicro.String())
	assert.Equal(t, "2020-08-05 13:14:15.999999999", user.DateTimeNano.String())

	assert.Equal(t, "1596633255", user.Timestamp.String())
	assert.Equal(t, "1596633255999", user.TimestampMilli.String())
	assert.Equal(t, "1596633255999999", user.TimestampMicro.String())
	assert.Equal(t, "1596633255999999999", user.TimestampNano.String())

	assert.Equal(t, int64(1596633255), user.Timestamp.Int64())
	assert.Equal(t, int64(1596633255999), user.TimestampMilli.Int64())
	assert.Equal(t, int64(1596633255999999), user.TimestampMicro.Int64())
	assert.Equal(t, int64(1596633255999999999), user.TimestampNano.Int64())

	assert.Equal(t, "2020-08-05 13:14:15", user.CreatedAt.String())
	assert.Equal(t, "2020-08-05 13:14:15", user.UpdatedAt.String())
	assert.Equal(t, "1596633255", user.DeletedAt.String())
	assert.Equal(t, int64(1596633255), user.DeletedAt.Int64())
}
