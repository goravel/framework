package mail

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/mail/template"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

var _ contractsfoundation.ServiceProviderWithFlushState = (*ServiceProvider)(nil)

func TestServiceProviderFlushState(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)

	t.Run("builds the engine again from the current configuration", func(t *testing.T) {
		template.FlushEngines()
		t.Cleanup(template.FlushEngines)

		first := mocksconfig.NewConfig(t)
		first.EXPECT().GetString("mail.template.default", "html").Return("html").Once()
		first.EXPECT().GetString("mail.template.engines.html.driver", "html").Return("html").Once()
		first.EXPECT().GetString("mail.template.engines.html.path", "resources/views/mail").Return("/first").Once()
		engine, err := template.Get(first)
		assert.NoError(t, err)

		provider.FlushState(app)

		second := mocksconfig.NewConfig(t)
		second.EXPECT().GetString("mail.template.default", "html").Return("html").Once()
		second.EXPECT().GetString("mail.template.engines.html.driver", "html").Return("html").Once()
		second.EXPECT().GetString("mail.template.engines.html.path", "resources/views/mail").Return("/second").Once()
		rebuilt, err := template.Get(second)
		assert.NoError(t, err)
		assert.NotSame(t, engine, rebuilt)
	})

	t.Run("is safe on empty state", func(t *testing.T) {
		template.FlushEngines()

		assert.NotPanics(t, func() {
			provider.FlushState(app)
			provider.FlushState(app)
		})
	})
}
