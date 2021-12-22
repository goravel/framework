package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func getInstance() *Redis {
	return &Redis{Redis: redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	}),
		Prefix: "goravel_cache:",
	}
}

func TestGet(t *testing.T) {
	r := getInstance()

	assert.Equal(t, "default", r.Get("test-get", "default").(string))
}

func TestHas(t *testing.T) {
	r := getInstance()

	assert.False(t, r.Has("test-has"))
	err := r.Put("test-has", "goravel", 5*time.Second)
	assert.Nil(t, err)
	assert.True(t, r.Has("test-has"))
}

func TestPut(t *testing.T) {
	r := getInstance()

	assert.Nil(t, r.Put("test-put", "goravel", 5*time.Second))
	assert.True(t, r.Has("test-put"))
	assert.Equal(t, "goravel", r.Get("test-put", "default"))
}

func TestPull(t *testing.T) {
	r := getInstance()

	assert.Nil(t, r.Put("test-put", "goravel", 5*time.Second))
	assert.True(t, r.Has("test-put"))
	assert.Equal(t, "goravel", r.Get("test-put", "default"))
}

func TestAdd(t *testing.T) {
	r := getInstance()

	assert.True(t, r.Add("test-add", "goravel", 5*time.Second))
	assert.True(t, r.Has("test-put"))
	assert.False(t, r.Add("test-add", "goravel", 5*time.Second))
}

func TestRemember(t *testing.T) {
	r := getInstance()

	val, err := r.Remember("test-remember", 5*time.Second, func() interface{} {
		return "goravel"
	})

	assert.Nil(t, err)
	assert.Equal(t, "goravel", val.(string))
}

func TestRememberForever(t *testing.T) {
	r := getInstance()

	val, err := r.RememberForever("test-remember-forever", func() interface{} {
		return "goravel"
	})

	assert.Nil(t, err)
	assert.Equal(t, "goravel", val.(string))
}

func TestForever(t *testing.T) {
	r := getInstance()

	val := r.Forever("test-forever", "goravel")

	assert.True(t, val)
	assert.Equal(t, "goravel", r.Get("test-forever", nil))
}

func TestForget(t *testing.T) {
	r := getInstance()

	val := r.Forget("test-forget")
	assert.True(t, val)

	err := r.Put("test-forget", "goravel", 5*time.Second)
	assert.Nil(t, err)
	assert.True(t, r.Forget("test-forget"))
}

func TestFlush(t *testing.T) {
	r := getInstance()

	err := r.Put("test-flush", "goravel", 5*time.Second)
	assert.Nil(t, err)
	assert.Equal(t, "goravel", r.Get("test-flush", nil).(string))

	r.Flush()
	assert.False(t, r.Has("test-flush"))
}
