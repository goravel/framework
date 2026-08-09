package broadcasting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractshttp "github.com/goravel/framework/contracts/http"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockshttp "github.com/goravel/framework/mocks/http"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	mocksroute "github.com/goravel/framework/mocks/route"
)

func TestBroadcastServiceProviderBoot_Middleware(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockQueue := mocksqueue.NewQueue(t)
	mockRoute := mocksroute.NewRoute(t)
	mockArtisan := mocksconsole.NewArtisan(t)
	testMiddleware := mockshttp.NewMiddleware(t)

	mockApp.EXPECT().MakeConfig().Return(mockConfig).Once()
	mockApp.EXPECT().MakeQueue().Return(mockQueue).Once()
	mockQueue.EXPECT().Register(mock.Anything).Once()

	mockConfig.EXPECT().UnmarshalKey("broadcasting", mock.MatchedBy(func(c *Config) bool {
		c.Default = "log"
		c.Auth.Enabled = true
		c.Auth.Path = "/broadcasting/auth"
		c.Auth.Middleware = []contractshttp.Middleware{testMiddleware}
		return true
	})).Return(nil).Once()

	mockApp.EXPECT().MakeRoute().Return(mockRoute).Once()
	mockApp.EXPECT().MakeBroadcast().Return(&Application{config: &Config{Default: "log"}}).Once()
	mockApp.EXPECT().MakeArtisan().Return(mockArtisan).Once()
	mockArtisan.EXPECT().Register(mock.Anything).Once()

	mockRoute.EXPECT().Middleware(testMiddleware).Return(mockRoute).Once()
	mockRoute.EXPECT().Post("/broadcasting/auth", mock.Anything).Return(mocksroute.NewAction(t)).Once()

	provider := &ServiceProvider{}
	provider.Boot(mockApp)
}

func TestBroadcastServiceProviderBoot_AuthDisabled(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockQueue := mocksqueue.NewQueue(t)
	mockArtisan := mocksconsole.NewArtisan(t)

	mockApp.EXPECT().MakeConfig().Return(mockConfig).Once()
	mockApp.EXPECT().MakeQueue().Return(mockQueue).Once()
	mockQueue.EXPECT().Register(mock.Anything).Once()

	mockConfig.EXPECT().UnmarshalKey("broadcasting", mock.MatchedBy(func(c *Config) bool {
		c.Default = "log"
		c.Auth.Enabled = false
		return true
	})).Return(nil).Once()

	mockApp.EXPECT().MakeArtisan().Return(mockArtisan).Once()
	mockArtisan.EXPECT().Register(mock.Anything).Once()

	provider := &ServiceProvider{}
	assert.NotPanics(t, func() { provider.Boot(mockApp) })
}
