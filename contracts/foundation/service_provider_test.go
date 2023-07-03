package foundation

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type test1ServiceProvider struct {
}

type test2ServiceProvider struct {
	*UnimplementedServiceProvider
}

func TestUnimplementedServiceProvider(t *testing.T) {
	var (
		sp1 = &test1ServiceProvider{}
		sp2 = &test2ServiceProvider{}
	)

	_, ok1 := interface{}(sp1).(ServiceProvider)
	_, ok2 := interface{}(sp2).(ServiceProvider)

	assert.False(t, ok1)
	assert.True(t, ok2)
}
