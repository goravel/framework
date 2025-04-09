package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateTimeMilli_Scan(t *testing.T) {
	c := NewDateTimeMilli(Now())

	t.Run("[]byte type", func(t *testing.T) {
		assert.Nil(t, c.Scan([]byte(Now().ToDateTimeMilliString())))
	})

	t.Run("string type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().ToDateTimeMilliString()))
	})

	t.Run("int64 type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().Timestamp()))
	})

	t.Run("time type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().StdTime()))
	})

	t.Run("unsupported type", func(t *testing.T) {
		assert.Error(t, c.Scan(nil))
	})
}

func TestDateTimeMilli_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}
		v, err := NewDateTimeMilli(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, err)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")
		v, err := NewDateTimeMilli(c).Value()
		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")
		v, err := NewDateTimeMilli(c).Value()
		assert.Equal(t, c.StdTime(), v)
		assert.Nil(t, err)
	})
}

func TestDateTimeMilli_MarshalJSON(t *testing.T) {
	type User struct {
		DateTimeMilli DateTimeMilli `json:"date_time_milli"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.DateTimeMilli = NewDateTimeMilli(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_time_milli":""}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.DateTimeMilli = NewDateTimeMilli(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999")
		user.DateTimeMilli = NewDateTimeMilli(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_time_milli":"2020-08-05 13:14:15.999"}`, string(data))
	})
}

func TestDateTimeMilli_UnmarshalJSON(t *testing.T) {
	type User struct {
		DateTimeMilli DateTimeMilli `json:"date_time_milli"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"date_time_milli":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTimeMilli.String())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"date_time_milli":"null"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTimeMilli.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"date_time_milli":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTimeMilli.String())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"date_time_milli":"2020-08-05 13:14:15.999"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "2020-08-05 13:14:15.999", user.DateTimeMilli.String())
	})
}

func TestDateTimeMilli_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")
	assert.Equal(t, "2020-08-05 13:14:15.999", NewDateTimeMilli(c).String())
}

func TestDateTimeMilli_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewDateTimeMilli(Now()).GormDataType())
}
