package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	Date      LayoutType[Date]      `json:"date"`
	DateMilli LayoutType[DateMilli] `json:"date_milli"`
	DateMicro LayoutType[DateMicro] `json:"date_micro"`
	DateNano  LayoutType[DateNano]  `json:"date_nano"`

	Time      LayoutType[Time]      `json:"time"`
	TimeMilli LayoutType[TimeMilli] `json:"time_milli"`
	TimeMicro LayoutType[TimeMicro] `json:"time_micro"`
	TimeNano  LayoutType[TimeNano]  `json:"time_nano"`

	DateTime      LayoutType[DateTime]      `json:"date_time"`
	DateTimeMilli LayoutType[DateTimeMilli] `json:"date_time_milli"`
	DateTimeMicro LayoutType[DateTimeMicro] `json:"date_time_micro"`
	DateTimeNano  LayoutType[DateTimeNano]  `json:"date_time_nano"`

	Timestamp      TimestampType[Timestamp]      `json:"timestamp"`
	TimestampMilli TimestampType[TimestampMilli] `json:"timestamp_milli"`
	TimestampMicro TimestampType[TimestampMicro] `json:"timestamp_micro"`
	TimestampNano  TimestampType[TimestampNano]  `json:"timestamp_nano"`
}

func TestType_MarshalJSON(t *testing.T) {
	var user User

	c := Parse("2020-08-05 13:14:15.999999999")

	user.Date = NewLayoutType[Date](c)
	user.DateMilli = NewLayoutType[DateMilli](c)
	user.DateMicro = NewLayoutType[DateMicro](c)
	user.DateNano = NewLayoutType[DateNano](c)

	user.Time = NewLayoutType[Time](c)
	user.TimeMilli = NewLayoutType[TimeMilli](c)
	user.TimeMicro = NewLayoutType[TimeMicro](c)
	user.TimeNano = NewLayoutType[TimeNano](c)

	user.DateTime = NewLayoutType[DateTime](c)
	user.DateTimeMilli = NewLayoutType[DateTimeMilli](c)
	user.DateTimeMicro = NewLayoutType[DateTimeMicro](c)
	user.DateTimeNano = NewLayoutType[DateTimeNano](c)

	user.Timestamp = NewTimestampType[Timestamp](c)
	user.TimestampMilli = NewTimestampType[TimestampMilli](c)
	user.TimestampMicro = NewTimestampType[TimestampMicro](c)
	user.TimestampNano = NewTimestampType[TimestampNano](c)

	data, err := json.Marshal(&user)
	assert.NoError(t, err)
	assert.Equal(t, `{"date":"2020-08-05","date_milli":"2020-08-05.999","date_micro":"2020-08-05.999999","date_nano":"2020-08-05.999999999","time":"13:14:15","time_milli":"13:14:15.999","time_micro":"13:14:15.999999","time_nano":"13:14:15.999999999","date_time":"2020-08-05 13:14:15","date_time_milli":"2020-08-05 13:14:15.999","date_time_micro":"2020-08-05 13:14:15.999999","date_time_nano":"2020-08-05 13:14:15.999999999","timestamp":1596633255,"timestamp_milli":1596633255999,"timestamp_micro":1596633255999999,"timestamp_nano":1596633255999999999}`, string(data))
}

func TestType_UnmarshalJSON(t *testing.T) {
	var user User

	value := `{"date":"2020-08-05","date_milli":"2020-08-05.999","date_micro":"2020-08-05.999999","date_nano":"2020-08-05.999999999","time":"13:14:15","time_milli":"13:14:15.999","time_micro":"13:14:15.999999","time_nano":"13:14:15.999999999","date_time":"2020-08-05 13:14:15","date_time_milli":"2020-08-05 13:14:15.999","date_time_micro":"2020-08-05 13:14:15.999999","date_time_nano":"2020-08-05 13:14:15.999999999","timestamp":1596633255,"timestamp_milli":1596633255999,"timestamp_micro":1596633255999999,"timestamp_nano":1596633255999999999}`
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
