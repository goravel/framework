package time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetTestNow(t *testing.T) {
	testNow := time.Now().Add(-1 * time.Hour)
	SetTestNow(testNow)
	assert.NotNil(t, now)
	SetTestNow()
	assert.Nil(t, now)
}

func TestNow(t *testing.T) {
	testNow := time.Now().Add(-10 * time.Second)
	SetTestNow(testNow)
	time.Sleep(2 * time.Second)
	assert.True(t, time.Now().Add(-9*time.Second).After(Now()))
	assert.True(t, time.Now().Add(-11*time.Second).Before(Now()))
}
