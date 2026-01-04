package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mocksconfig "github.com/goravel/framework/mocks/config"
)

func TestPrefix(t *testing.T) {
	t.Run("with value", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("cache.prefix").Return("myprefix").Once()
		got := prefix(mockConfig)
		assert.Equal(t, "myprefix:", got)
	})

	t.Run("empty", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("cache.prefix").Return("").Once()
		got := prefix(mockConfig)
		assert.Equal(t, ":", got)
	})
}
