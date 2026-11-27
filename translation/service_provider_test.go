package translation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockstranslation "github.com/goravel/framework/mocks/translation"
)

var _ contractsfoundation.ServiceProviderWithFlushState = (*ServiceProvider)(nil)

func TestServiceProviderFlushState(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)

	t.Run("discards loaded groups so a changed loader is used", func(t *testing.T) {
		loaded.Clear()
		t.Cleanup(loaded.Clear)

		oldLoader := mockstranslation.NewLoader(t)
		oldLoader.EXPECT().Load("en", "test").Return(map[string]any{"key": "old"}, nil).Once()
		translator := NewTranslator(t.Context(), nil, oldLoader, "en", "en", nil)
		lines, err := translator.load("en", "test")
		assert.NoError(t, err)
		assert.Equal(t, "old", lines["key"])

		provider.FlushState(app)

		newLoader := mockstranslation.NewLoader(t)
		newLoader.EXPECT().Load("en", "test").Return(map[string]any{"key": "new"}, nil).Once()
		translator = NewTranslator(t.Context(), nil, newLoader, "en", "en", nil)
		lines, err = translator.load("en", "test")
		assert.NoError(t, err)
		assert.Equal(t, "new", lines["key"])
	})

	t.Run("is safe on empty state", func(t *testing.T) {
		loaded.Clear()

		assert.NotPanics(t, func() {
			provider.FlushState(app)
			provider.FlushState(app)
		})
	})
}
