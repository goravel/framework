package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateMicro_Scan(t *testing.T) {
	c := NewDateMicro(Now())

	t.Run("[]byte type", func(t *testing.T) {
		assert.Nil(t, c.Scan([]byte(Now().ToDateMicroString())))
	})

	t.Run("string type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().ToDateMicroString()))
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

func TestDateMicro_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}
		v, err := NewDateMicro(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, err)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")
		v, err := NewDateMicro(c).Value()
		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")
		v, err := NewDateMicro(c).Value()
		assert.Equal(t, c.StdTime(), v)
		assert.Nil(t, err)
	})
}

func TestDateMicro_MarshalJSON(t *testing.T) {
	type User struct {
		DateMicro DateMicro `json:"date_micro"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.DateMicro = NewDateMicro(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_micro":""}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.DateMicro = NewDateMicro(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999")
		user.DateMicro = NewDateMicro(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_micro":"2020-08-05.999999"}`, string(data))
	})
}

func TestDateMicro_UnmarshalJSON(t *testing.T) {
	type User struct {
		DateMicro DateMicro `json:"date_micro"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"date_micro":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateMicro.String())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"date_micro":"null"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateMicro.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"date_micro":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateMicro.String())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"date_micro":"2020-08-05.999999"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "2020-08-05.999999", user.DateMicro.String())
	})
}

func TestDateMicro_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")
	assert.Equal(t, "2020-08-05.999999", NewDateMicro(c).String())
}

func TestDateMicro_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewDateMicro(Now()).GormDataType())
}
