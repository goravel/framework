package carbon_test

import (
	"encoding/json"
	"testing"

	"github.com/goravel/framework/support/carbon"
	"github.com/stretchr/testify/assert"
)

func TestCarbonType(t *testing.T) {
	type Model struct {
		Carbon1 carbon.Carbon  `json:"carbon1"`
		Carbon2 *carbon.Carbon `json:"carbon2"`
	}

	var model1 Model
	c := carbon.Parse("2020-08-05 13:14:15.999999999")
	model1.Carbon1 = *c
	model1.Carbon2 = c

	v, e := json.Marshal(&model1)
	assert.Nil(t, e)
	assert.Equal(t, `{"carbon1":"2020-08-05 13:14:15","carbon2":"2020-08-05 13:14:15"}`, string(v))

	var model2 Model
	assert.NoError(t, json.Unmarshal(v, &model2))

	assert.Equal(t, "2020-08-05 13:14:15", model2.Carbon1.String())
	assert.Equal(t, "2020-08-05 13:14:15", model2.Carbon2.String())

}

func TestBuiltinType(t *testing.T) {
	type Model struct {
		Date      carbon.Date      `json:"date"`
		DateMilli carbon.DateMilli `json:"date_milli"`
		DateMicro carbon.DateMicro `json:"date_micro"`
		DateNano  carbon.DateNano  `json:"date_nano"`

		Time      carbon.Time      `json:"time"`
		TimeMilli carbon.TimeMilli `json:"time_milli"`
		TimeMicro carbon.TimeMicro `json:"time_micro"`
		TimeNano  carbon.TimeNano  `json:"time_nano"`

		DateTime      carbon.DateTime      `json:"date_time"`
		DateTimeMilli carbon.DateTimeMilli `json:"date_time_milli"`
		DateTimeMicro carbon.DateTimeMicro `json:"date_time_micro"`
		DateTimeNano  carbon.DateTimeNano  `json:"date_time_nano"`

		CreatedAt *carbon.DateTime `json:"created_at"`
		UpdatedAt *carbon.DateTime `json:"updated_at"`

		Timestamp      carbon.Timestamp      `json:"timestamp"`
		TimestampMilli carbon.TimestampMilli `json:"timestamp_milli"`
		TimestampMicro carbon.TimestampMicro `json:"timestamp_micro"`
		TimestampNano  carbon.TimestampNano  `json:"timestamp_nano"`

		DeletedAt *carbon.Timestamp `json:"deleted_at"`
	}

	var model1 Model
	c := carbon.Parse("2020-08-05 13:14:15.999999999")

	model1.Date = *carbon.NewDate(c)
	model1.DateMilli = *carbon.NewDateMilli(c)
	model1.DateMicro = *carbon.NewDateMicro(c)
	model1.DateNano = *carbon.NewDateNano(c)

	model1.Time = *carbon.NewTime(c)
	model1.TimeMilli = *carbon.NewTimeMilli(c)
	model1.TimeMicro = *carbon.NewTimeMicro(c)
	model1.TimeNano = *carbon.NewTimeNano(c)

	model1.DateTime = *carbon.NewDateTime(c)
	model1.DateTimeMilli = *carbon.NewDateTimeMilli(c)
	model1.DateTimeMicro = *carbon.NewDateTimeMicro(c)
	model1.DateTimeNano = *carbon.NewDateTimeNano(c)

	model1.Timestamp = *carbon.NewTimestamp(c)
	model1.TimestampMilli = *carbon.NewTimestampMilli(c)
	model1.TimestampMicro = *carbon.NewTimestampMicro(c)
	model1.TimestampNano = *carbon.NewTimestampNano(c)

	model1.CreatedAt = carbon.NewDateTime(c)
	model1.UpdatedAt = carbon.NewDateTime(c)
	model1.DeletedAt = carbon.NewTimestamp(c)

	v, e := json.Marshal(&model1)
	assert.Nil(t, e)
	assert.Equal(t, `{"date":"2020-08-05","date_milli":"2020-08-05.999","date_micro":"2020-08-05.999999","date_nano":"2020-08-05.999999999","time":"13:14:15","time_milli":"13:14:15.999","time_micro":"13:14:15.999999","time_nano":"13:14:15.999999999","date_time":"2020-08-05 13:14:15","date_time_milli":"2020-08-05 13:14:15.999","date_time_micro":"2020-08-05 13:14:15.999999","date_time_nano":"2020-08-05 13:14:15.999999999","created_at":"2020-08-05 13:14:15","updated_at":"2020-08-05 13:14:15","timestamp":1596633255,"timestamp_milli":1596633255999,"timestamp_micro":1596633255999999,"timestamp_nano":1596633255999999999,"deleted_at":1596633255}`, string(v))

	var model2 Model
	assert.NoError(t, json.Unmarshal(v, &model2))

	assert.Equal(t, "2020-08-05", model2.Date.String())
	assert.Equal(t, "2020-08-05.999", model2.DateMilli.String())
	assert.Equal(t, "2020-08-05.999999", model2.DateMicro.String())
	assert.Equal(t, "2020-08-05.999999999", model2.DateNano.String())

	assert.Equal(t, "13:14:15", model2.Time.String())
	assert.Equal(t, "13:14:15.999", model2.TimeMilli.String())
	assert.Equal(t, "13:14:15.999999", model2.TimeMicro.String())
	assert.Equal(t, "13:14:15.999999999", model2.TimeNano.String())

	assert.Equal(t, "2020-08-05 13:14:15", model2.DateTime.String())
	assert.Equal(t, "2020-08-05 13:14:15.999", model2.DateTimeMilli.String())
	assert.Equal(t, "2020-08-05 13:14:15.999999", model2.DateTimeMicro.String())
	assert.Equal(t, "2020-08-05 13:14:15.999999999", model2.DateTimeNano.String())

	assert.Equal(t, "1596633255", model2.Timestamp.String())
	assert.Equal(t, "1596633255999", model2.TimestampMilli.String())
	assert.Equal(t, "1596633255999999", model2.TimestampMicro.String())
	assert.Equal(t, "1596633255999999999", model2.TimestampNano.String())

	assert.Equal(t, int64(1596633255), model2.Timestamp.Int64())
	assert.Equal(t, int64(1596633255999), model2.TimestampMilli.Int64())
	assert.Equal(t, int64(1596633255999999), model2.TimestampMicro.Int64())
	assert.Equal(t, int64(1596633255999999999), model2.TimestampNano.Int64())

	assert.Equal(t, "2020-08-05 13:14:15", model2.CreatedAt.String())
	assert.Equal(t, "2020-08-05 13:14:15", model2.UpdatedAt.String())
	assert.Equal(t, "1596633255", model2.DeletedAt.String())
	assert.Equal(t, int64(1596633255), model2.DeletedAt.Int64())
}

type ISO8601Type string

func (t ISO8601Type) Format() string {
	return carbon.ISO8601Format
}

type RFC3339Type string

func (t RFC3339Type) Layout() string {
	return carbon.RFC3339Layout
}

func TestCustomerType(t *testing.T) {
	type Model struct {
		Customer1 carbon.FormatType[ISO8601Type] `json:"customer1"`
		Customer2 carbon.LayoutType[RFC3339Type] `json:"customer2"`

		CreatedAt *carbon.FormatType[ISO8601Type] `json:"created_at"`
		UpdatedAt *carbon.LayoutType[RFC3339Type] `json:"updated_at"`
	}

	var model1 Model
	c := carbon.Parse("2020-08-05 13:14:15.999999999")

	model1.Customer1 = *carbon.NewFormatType[ISO8601Type](c)
	model1.Customer2 = *carbon.NewLayoutType[RFC3339Type](c)

	model1.CreatedAt = carbon.NewFormatType[ISO8601Type](c)
	model1.UpdatedAt = carbon.NewLayoutType[RFC3339Type](c)

	v, e := json.Marshal(&model1)
	assert.Nil(t, e)
	assert.Equal(t, `{"customer1":"2020-08-05T13:14:15+00:00","customer2":"2020-08-05T13:14:15Z","created_at":"2020-08-05T13:14:15+00:00","updated_at":"2020-08-05T13:14:15Z"}`, string(v))

	var model2 Model
	assert.NoError(t, json.Unmarshal(v, &model2))

	assert.Equal(t, "2020-08-05T13:14:15+00:00", model2.Customer1.String())
	assert.Equal(t, "2020-08-05T13:14:15Z", model2.Customer2.String())
	assert.Equal(t, "2020-08-05T13:14:15+00:00", model2.CreatedAt.String())
	assert.Equal(t, "2020-08-05T13:14:15Z", model2.UpdatedAt.String())
}
