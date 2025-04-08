package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeMicro_Scan(t *testing.T) {
	c := NewTimeMicro(Now())

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

func TestTimeMicro_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}
		v, err := NewTimeMicro(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, err)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")
		v, err := NewTimeMicro(c).Value()
		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")
		v, err := NewTimeMicro(c).Value()
		assert.Equal(t, c.StdTime(), v)
		assert.Nil(t, err)
	})
}

func TestTimeMicro_MarshalJSON(t *testing.T) {
	type User struct {
		TimeMicro TimeMicro `json:"time_micro"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.TimeMicro = NewTimeMicro(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"time_micro":""}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.TimeMicro = NewTimeMicro(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999999")
		user.TimeMicro = NewTimeMicro(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"time_micro":"13:14:15.999999"}`, string(data))
	})
}

func TestTimeMicro_UnmarshalJSON(t *testing.T) {
	type User struct {
		TimeMicro TimeMicro `json:"time_micro"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"time_micro":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.TimeMicro.String())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"time_micro":"null"}`

		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.TimeMicro.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"time_micro":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))

		assert.Empty(t, user.TimeMicro.String())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"time_micro":"13:14:15.999999"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "13:14:15.999999", user.TimeMicro.String())
	})
}

func TestTimeMicro_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999999")
	assert.Equal(t, "13:14:15.999999", NewTimeMicro(c).String())
}

func TestTimeMicro_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewTimeMicro(Now()).GormDataType())
}
