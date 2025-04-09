package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTime_Scan(t *testing.T) {
	c := NewTime(Now())

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

func TestTime_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}
		v, err := NewTime(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, err)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")
		v, err := NewTime(c).Value()
		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")
		v, err := NewTime(c).Value()
		assert.Equal(t, c.StdTime(), v)
		assert.Nil(t, err)
	})
}

func TestTime_MarshalJSON(t *testing.T) {
	type User struct {
		Time Time `json:"time"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.Time = NewTime(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"time":""}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.Time = NewTime(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999")
		user.Time = NewTime(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"time":"13:14:15"}`, string(data))
	})
}

func TestTime_UnmarshalJSON(t *testing.T) {
	type User struct {
		Time Time `json:"time"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"time":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.Time.String())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"time":"null"}`

		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.Time.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"time":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))

		assert.Empty(t, user.Time.String())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"time":"13:14:15"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "13:14:15", user.Time.String())
	})
}

func TestTime_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")
	assert.Equal(t, "13:14:15", NewTime(c).String())
}

func TestTime_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewTime(Now()).GormDataType())
}
