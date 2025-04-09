package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateTime_Scan(t *testing.T) {
	c := NewDateTime(Now())

	t.Run("[]byte type", func(t *testing.T) {
		assert.Nil(t, c.Scan([]byte(Now().ToDateTimeString())))
	})

	t.Run("string type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().ToDateTimeString()))
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

func TestDateTime_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}
		v, err := NewDateTime(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, err)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")
		v, err := NewDateTime(c).Value()
		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")
		v, err := NewDateTime(c).Value()
		assert.Equal(t, c.StdTime(), v)
		assert.Nil(t, err)
	})
}

func TestDateTime_MarshalJSON(t *testing.T) {
	type User struct {
		DateTime DateTime `json:"date_time"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.DateTime = NewDateTime(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_time":""}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.DateTime = NewDateTime(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999")
		user.DateTime = NewDateTime(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_time":"2020-08-05 13:14:15"}`, string(data))
	})
}

func TestDateTime_UnmarshalJSON(t *testing.T) {
	type User struct {
		DateTime DateTime `json:"date_time"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"date_time":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTime.String())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"date_time":"null"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTime.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"date_time":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateTime.String())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"date_time":"2020-08-05 13:14:15"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "2020-08-05 13:14:15", user.DateTime.String())
	})
}

func TestDateTime_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")
	assert.Equal(t, "2020-08-05 13:14:15", NewDateTime(c).String())
}

func TestDateTime_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewDateTime(Now()).GormDataType())
}
