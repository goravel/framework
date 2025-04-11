package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateTimeMicro_Scan(t *testing.T) {
	c := NewDateTimeMicro(Now())

	t.Run("[]byte type", func(t *testing.T) {
		assert.Nil(t, c.Scan([]byte(Now().ToDateTimeMicroString())))
	})

	t.Run("string type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().ToDateTimeMicroString()))
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

func TestDateTimeMicro_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}
		v, err := NewDateTimeMicro(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, err)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")
		v, err := NewDateTimeMicro(c).Value()
		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")
		v, err := NewDateTimeMicro(c).Value()
		assert.Equal(t, c.StdTime(), v)
		assert.Nil(t, err)
	})
}

func TestDateTimeMicro_MarshalJSON(t *testing.T) {
	type User struct {
		DateTimeMicro DateTimeMicro `json:"date_time_micro"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.DateTimeMicro = NewDateTimeMicro(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_time_micro":""}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.DateTimeMicro = NewDateTimeMicro(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999999")
		user.DateTimeMicro = NewDateTimeMicro(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_time_micro":"2020-08-05 13:14:15.999999"}`, string(data))
	})
}

func TestDateTimeMicro_UnmarshalJSON(t *testing.T) {
	type User struct {
		DateTimeMicro DateTimeMicro `json:"date_time_micro"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"date_time_micro":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTimeMicro.String())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"date_time_micro":"null"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTimeMicro.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"date_time_micro":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTimeMicro.String())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"date_time_micro":"2020-08-05 13:14:15.999999"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "2020-08-05 13:14:15.999999", user.DateTimeMicro.String())
	})
}

func TestDateTimeMicro_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999999")
	assert.Equal(t, "2020-08-05 13:14:15.999999", NewDateTimeMicro(c).String())
}

func TestDateTimeMicro_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewDateTimeMicro(Now()).GormDataType())
}
