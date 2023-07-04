package foundation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testRegister = 0

type testServiceProvider struct {
	*BaseServiceProvider
}

func (t *testServiceProvider) Register(Application) {
	testRegister++
}

func TestBaseServiceProvider(t *testing.T) {
	var sp = &testServiceProvider{}

	_, ok := interface{}(sp).(ServiceProvider)

	assert.True(t, ok)

	sp.Register(nil)
	sp.Boot(nil)

	assert.Equal(t, 1, testRegister)
}
