package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeMilli_Scan(t *testing.T) {
	c := NewTimeMilli(Now())

	t.Run("[]byte type", func(t *testing.T) {
		assert.Nil(t, c.Scan([]byte(Now().ToDateString())))
	})

	t.Run("string type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().ToDateString()))
	})

	t.Run("int64 type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().Timestamp()))
	})

	t.Run("time type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().StdTime()))
	})

	t.Run("unsupported type", func(t *testing.T) {
		e := c.Scan(nil)
		assert.Error(t, e)
	})
}

func TestTimeMilli_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}
		v, err := NewTimeMilli(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, err)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")
		v, err := NewTimeMilli(c).Value()
		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")
		v, err := NewTimeMilli(c).Value()
		assert.Equal(t, c.StdTime(), v)
		assert.Nil(t, err)
	})
}

func TestTimeMilli_MarshalJSON(t *testing.T) {
	type User struct {
		TimeMilli TimeMilli `json:"time_milli"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.TimeMilli = NewTimeMilli(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"time_milli":""}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.TimeMilli = NewTimeMilli(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999")
		user.TimeMilli = NewTimeMilli(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"time_milli":"13:14:15.999"}`, string(data))
	})
}

func TestTimeMilli_UnmarshalJSON(t *testing.T) {
	type User struct {
		TimeMilli TimeMilli `json:"time_milli"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"time_milli":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.TimeMilli.String())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"time_milli":"null"}`

		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.TimeMilli.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"time_milli":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))

		assert.Empty(t, user.TimeMilli.String())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"time_milli":"13:14:15.999"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "13:14:15.999", user.TimeMilli.String())
	})
}

func TestTimeMilli_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")
	assert.Equal(t, "13:14:15.999", NewTimeMilli(c).String())
}

func TestTimeMilli_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewTimeMilli(Now()).GormDataType())
}
