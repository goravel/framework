package carbon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeNano_Scan(t *testing.T) {
	c := NewTimeNano(Now())

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

func TestTimeNano_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}
		v, err := NewTimeNano(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, err)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")
		v, err := NewTimeNano(c).Value()
		assert.Nil(t, v)
		assert.Error(t, err)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")
		v, err := NewTimeNano(c).Value()
		assert.Equal(t, c.StdTime(), v)
		assert.Nil(t, err)
	})
}

func TestTimeNano_MarshalJSON(t *testing.T) {
	type User struct {
		TimeNano TimeNano `json:"time_nano"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.TimeNano = NewTimeNano(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"time_nano":""}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.TimeNano = NewTimeNano(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999999999")
		user.TimeNano = NewTimeNano(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"time_nano":"13:14:15.999999999"}`, string(data))
	})
}

func TestTimeNano_UnmarshalJSON(t *testing.T) {
	type User struct {
		TimeNano TimeNano `json:"time_nano"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"time_nano":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.TimeNano.String())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"time_nano":"null"}`

		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Empty(t, user.TimeNano.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"time_nano":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))

		assert.Empty(t, user.TimeNano.String())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"time_nano":"13:14:15.999999999"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "13:14:15.999999999", user.TimeNano.String())
	})
}

func TestTimeNano_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999999999")
	assert.Equal(t, "13:14:15.999999999", NewTimeNano(c).String())
}

func TestTimeNano_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewTimeNano(Now()).GormDataType())
}
