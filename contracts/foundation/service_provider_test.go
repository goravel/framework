package foundation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testRegister = 0

type test1ServiceProvider struct {
}

type test2ServiceProvider struct {
	*BaseServiceProvider
}

func (t *test2ServiceProvider) Register(Application) {
	testRegister++
}

func TestBaseServiceProvider(t *testing.T) {
	var (
		sp1 = &test1ServiceProvider{}
		sp2 = &test2ServiceProvider{}
	)

	_, ok1 := interface{}(sp1).(ServiceProvider)
	_, ok2 := interface{}(sp2).(ServiceProvider)

	assert.False(t, ok1)
	assert.True(t, ok2)

	sp2.Register(nil)
	assert.Equal(t, 1, testRegister)
	sp2.Boot(nil)
}
