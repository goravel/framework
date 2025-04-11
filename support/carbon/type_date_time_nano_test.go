package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateTimeNano_Scan(t *testing.T) {
	c := NewDateTimeNano(Now())

	t.Run("[]byte type", func(t *testing.T) {
		assert.Nil(t, c.Scan([]byte(Now().ToDateTimeNanoString())))
	})

	t.Run("string type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().ToDateTimeNanoString()))
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

func TestDateTimeNano_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}
		v, err := NewDateTimeNano(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, err)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")
		v, err := NewDateTimeNano(c).Value()
		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")
		v, err := NewDateTimeNano(c).Value()
		assert.Equal(t, c.StdTime(), v)
		assert.Nil(t, err)
	})
}

func TestDateTimeNano_MarshalJSON(t *testing.T) {
	type User struct {
		DateTimeNano DateTimeNano `json:"date_time_nano"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.DateTimeNano = NewDateTimeNano(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_time_nano":""}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.DateTimeNano = NewDateTimeNano(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999999")
		user.DateTimeNano = NewDateTimeNano(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_time_nano":"2020-08-05 13:14:15.999999999"}`, string(data))
	})
}

func TestDateTimeNano_UnmarshalJSON(t *testing.T) {
	type User struct {
		DateTimeNano DateTimeNano `json:"date_time_nano"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"date_time_nano":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTimeNano.String())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"date_time_nano":"null"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTimeNano.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"date_time_nano":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTimeNano.String())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"date_time_nano":"2020-08-05 13:14:15.999999999"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "2020-08-05 13:14:15.999999999", user.DateTimeNano.String())
	})
}

func TestDateTimeNano_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999999")
	assert.Equal(t, "2020-08-05 13:14:15.999999999", NewDateTimeNano(c).String())
}

func TestDateTimeNano_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewDateTimeNano(Now()).GormDataType())
}
