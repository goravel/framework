package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateMilli_Scan(t *testing.T) {
	c := NewDateMilli(Now())

	t.Run("[]byte type", func(t *testing.T) {
		assert.Nil(t, c.Scan([]byte(Now().ToDateMilliString())))
	})

	t.Run("string type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().ToDateMilliString()))
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

func TestDateMilli_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}
		v, err := NewDateMilli(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, err)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")
		v, err := NewDateMilli(c).Value()
		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")
		v, err := NewDateMilli(c).Value()
		assert.Equal(t, c.StdTime(), v)
		assert.Nil(t, err)
	})
}

func TestDateMilli_MarshalJSON(t *testing.T) {
	type User struct {
		DateMilli DateMilli `json:"date_milli"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.DateMilli = NewDateMilli(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_milli":""}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.DateMilli = NewDateMilli(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999")
		user.DateMilli = NewDateMilli(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_milli":"2020-08-05.999"}`, string(data))
	})
}

func TestDateMilli_UnmarshalJSON(t *testing.T) {
	type User struct {
		DateMilli DateMilli `json:"date_milli"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"date_milli":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateMilli.String())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"date_milli":"null"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateMilli.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"date_milli":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateMilli.String())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"date_milli":"2020-08-05.999"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "2020-08-05.999", user.DateMilli.String())
	})
}

func TestDateMilli_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")
	assert.Equal(t, "2020-08-05.999", NewDateMilli(c).String())
}

func TestDateMilli_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewDateMilli(Now()).GormDataType())
}
