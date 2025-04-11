package carbon

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimestampNano_Scan(t *testing.T) {
	t.Run("[]byte type", func(t *testing.T) {
		ts := NewTimestampNano(Now())
		assert.Nil(t, ts.Scan([]byte(strconv.Itoa(int(ts.Timestamp())))))
		assert.Error(t, ts.Scan([]byte("xxx")))
	})

	t.Run("string type", func(t *testing.T) {
		ts := NewTimestampNano(Now())
		assert.Nil(t, ts.Scan(strconv.Itoa(int(ts.Timestamp()))))
		assert.Error(t, ts.Scan("xxx"))
	})

	t.Run("int64 type", func(t *testing.T) {
		ts := NewTimestampNano(Now())
		assert.Nil(t, ts.Scan(Now().Timestamp()))
	})

	t.Run("time type", func(t *testing.T) {
		ts := NewTimestampNano(Now())
		assert.Nil(t, ts.Scan(Now().StdTime()))
	})

	t.Run("unsupported type", func(t *testing.T) {
		ts := NewTimestampNano(Now())
		assert.Error(t, ts.Scan(nil))
	})
}

func TestTimestampNano_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}

		v, e := NewTimestampNano(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, e)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")

		v, e := NewTimestampNano(c).Value()
		assert.Nil(t, v)
		assert.Error(t, e)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")

		v, e := NewTimestampNano(c).Value()
		assert.Equal(t, c.TimestampNano(), v)
		assert.Nil(t, e)
	})
}

func TestTimestampNano_MarshalJSON(t *testing.T) {
	type User struct {
		TimestampNano TimestampNano `json:"timestamp_nano"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.TimestampNano = NewTimestampNano(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"timestamp_nano":0}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.TimestampNano = NewTimestampNano(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999")
		user.TimestampNano = NewTimestampNano(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"timestamp_nano":1596633255999999999}`, string(data))
	})
}

func TestTimestampNano_UnmarshalJSON(t *testing.T) {
	type User struct {
		TimestampNano TimestampNano `json:"timestamp_nano"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"timestamp_nano":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "0", user.TimestampNano.String())
		assert.Zero(t, user.TimestampNano.Int64())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"timestamp_nano":"null"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))

		assert.Equal(t, "0", user.TimestampNano.String())
		assert.Zero(t, user.TimestampNano.Int64())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"timestamp_nano":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))

		assert.Equal(t, "0", user.TimestampNano.String())
		assert.Zero(t, user.TimestampNano.Int64())
	})

	t.Run("invalid value", func(t *testing.T) {
		var user User

		value := `{"timestamp_nano":"xxx"}`
		assert.Error(t, json.Unmarshal([]byte(value), &user))

		assert.Equal(t, "0", user.TimestampNano.String())

		assert.Zero(t, user.TimestampNano.Int64())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"timestamp_nano":1596633255999999999}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))

		assert.Equal(t, "1596633255999999999", user.TimestampNano.String())
		assert.Equal(t, int64(1596633255999999999), user.TimestampNano.Int64())
	})
}

func TestTimestampNano_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")
	assert.Equal(t, "1596633255999999999", NewTimestampNano(c).String())
}

func TestTimestampNano_Int64(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")
	assert.Equal(t, int64(1596633255999999999), NewTimestampNano(c).Int64())
}

func TestTimestampNano_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewTimestampNano(Now()).GormDataType())
}
