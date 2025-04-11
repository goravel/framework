package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDateNano_Scan(t *testing.T) {
	c := NewDateNano(Now())

	t.Run("[]byte type", func(t *testing.T) {
		assert.Nil(t, c.Scan([]byte(Now().ToDateNanoString())))
	})

	t.Run("string type", func(t *testing.T) {
		assert.Nil(t, c.Scan(Now().ToDateNanoString()))
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

func TestDateNano_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}
		v, err := NewDateNano(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, err)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")
		v, err := NewDateNano(c).Value()
		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")
		v, err := NewDateNano(c).Value()
		assert.Equal(t, c.StdTime(), v)
		assert.Nil(t, err)
	})
}

func TestDateNano_MarshalJSON(t *testing.T) {
	type User struct {
		DateNano DateNano `json:"date_nano"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.DateNano = NewDateNano(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_nano":""}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.DateNano = NewDateNano(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999")
		user.DateNano = NewDateNano(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"date_nano":"2020-08-05.999999999"}`, string(data))
	})
}

func TestDateNano_UnmarshalJSON(t *testing.T) {
	type User struct {
		DateNano DateNano `json:"date_nano"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"date_nano":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateNano.String())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"date_nano":"null"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateNano.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"date_nano":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.DateNano.String())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"date_nano":"2020-08-05.999999999"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "2020-08-05.999999999", user.DateNano.String())
	})
}

func TestDateNano_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")
	assert.Equal(t, "2020-08-05.999999999", NewDateNano(c).String())
}

func TestDateNano_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewDateNano(Now()).GormDataType())
}
