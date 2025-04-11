package carbon

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimestampMicro_Scan(t *testing.T) {
	t.Run("[]byte type", func(t *testing.T) {
		ts := NewTimestampMicro(Now())
		assert.Nil(t, ts.Scan([]byte(strconv.Itoa(int(ts.Timestamp())))))
		assert.Error(t, ts.Scan([]byte("xxx")))
	})

	t.Run("string type", func(t *testing.T) {
		ts := NewTimestampMicro(Now())
		assert.Nil(t, ts.Scan(strconv.Itoa(int(ts.Timestamp()))))
		assert.Error(t, ts.Scan("xxx"))
	})

	t.Run("int64 type", func(t *testing.T) {
		ts := NewTimestampMicro(Now())
		assert.Nil(t, ts.Scan(Now().Timestamp()))
	})

	t.Run("time type", func(t *testing.T) {
		ts := NewTimestampMicro(Now())
		assert.Nil(t, ts.Scan(Now().StdTime()))
	})

	t.Run("unsupported type", func(t *testing.T) {
		ts := NewTimestampMicro(Now())
		assert.Error(t, ts.Scan(nil))
	})
}

func TestTimestampMicro_Value(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		c := Carbon{}

		v, e := NewTimestampMicro(c).Value()
		assert.Nil(t, v)
		assert.Nil(t, e)
	})

	t.Run("invalid time", func(t *testing.T) {
		c := Parse("xxx")

		v, e := NewTimestampMicro(c).Value()
		assert.Nil(t, v)
		assert.Error(t, e)
	})

	t.Run("valid time", func(t *testing.T) {
		c := Parse("2020-08-05")

		v, e := NewTimestampMicro(c).Value()
		assert.Equal(t, c.TimestampMicro(), v)
		assert.Nil(t, e)
	})
}

func TestTimestampMicro_MarshalJSON(t *testing.T) {
	type User struct {
		TimestampMicro TimestampMicro `json:"timestamp_micro"`
	}

	t.Run("zero time", func(t *testing.T) {
		var user User

		c := Carbon{}
		user.TimestampMicro = NewTimestampMicro(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"timestamp_micro":0}`, string(data))
	})

	t.Run("invalid time", func(t *testing.T) {
		var user User

		c := Parse("xxx")
		user.TimestampMicro = NewTimestampMicro(c)

		data, err := json.Marshal(&user)
		assert.Error(t, err)
		assert.Empty(t, string(data))
	})

	t.Run("valid time", func(t *testing.T) {
		var user User

		c := Parse("2020-08-05 13:14:15.999999999")
		user.TimestampMicro = NewTimestampMicro(c)

		data, err := json.Marshal(&user)
		assert.NoError(t, err)
		assert.Equal(t, `{"timestamp_micro":1596633255999999}`, string(data))
	})
}

func TestTimestampMicro_UnmarshalJSON(t *testing.T) {
	type User struct {
		TimestampMicro TimestampMicro `json:"timestamp_micro"`
	}

	t.Run("empty value", func(t *testing.T) {
		var user User

		value := `{"timestamp_micro":""}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))
		assert.Equal(t, "0", user.TimestampMicro.String())
		assert.Zero(t, user.TimestampMicro.Int64())
	})

	t.Run("null value", func(t *testing.T) {
		var user User

		value := `{"timestamp_micro":"null"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))

		assert.Equal(t, "0", user.TimestampMicro.String())
		assert.Zero(t, user.TimestampMicro.Int64())
	})

	t.Run("zero value", func(t *testing.T) {
		var user User

		value := `{"timestamp_micro":"0"}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))

		assert.Equal(t, "0", user.TimestampMicro.String())
		assert.Zero(t, user.TimestampMicro.Int64())
	})

	t.Run("invalid value", func(t *testing.T) {
		var user User

		value := `{"timestamp_micro":"xxx"}`
		assert.Error(t, json.Unmarshal([]byte(value), &user))

		assert.Equal(t, "0", user.TimestampMicro.String())

		assert.Zero(t, user.TimestampMicro.Int64())
	})

	t.Run("valid value", func(t *testing.T) {
		var user User

		value := `{"timestamp_micro":1596633255999999}`
		assert.NoError(t, json.Unmarshal([]byte(value), &user))

		assert.Equal(t, "1596633255999999", user.TimestampMicro.String())
		assert.Equal(t, int64(1596633255999999), user.TimestampMicro.Int64())
	})
}

func TestTimestampMicro_String(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")
	assert.Equal(t, "1596633255999999", NewTimestampMicro(c).String())
}

func TestTimestampMicro_Int64(t *testing.T) {
	c := Parse("2020-08-05 13:14:15.999999999")
	assert.Equal(t, int64(1596633255999999), NewTimestampMicro(c).Int64())
}

func TestTimestampMicro_GormDataType(t *testing.T) {
	assert.Equal(t, "time", NewTimestampMicro(Now()).GormDataType())
}
