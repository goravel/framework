package broadcasting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsconsole "github.com/goravel/framework/contracts/console"
	contractshttp "github.com/goravel/framework/contracts/http"
	mocksconfig "github.com/goravel/framework/mocks/config"
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
	mockApp.EXPECT().Commands(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		return len(commands) == 1 && commands[0].Signature() == "make:channel"
	})).Once()

	mockRoute.EXPECT().Middleware(testMiddleware).Return(mockRoute).Once()
	mockRoute.EXPECT().Post("/broadcasting/auth", mock.Anything).Return(mocksroute.NewAction(t)).Once()

	provider := &ServiceProvider{}
	provider.Boot(mockApp)
}

func TestBroadcastServiceProviderBoot_AuthDisabled(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockQueue := mocksqueue.NewQueue(t)

	mockApp.EXPECT().MakeConfig().Return(mockConfig).Once()
	mockApp.EXPECT().MakeQueue().Return(mockQueue).Once()
	mockQueue.EXPECT().Register(mock.Anything).Once()

	mockConfig.EXPECT().UnmarshalKey("broadcasting", mock.MatchedBy(func(c *Config) bool {
		c.Default = "log"
		c.Auth.Enabled = false
		return true
	})).Return(nil).Once()

	mockApp.EXPECT().Commands(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		return len(commands) == 1 && commands[0].Signature() == "make:channel"
	})).Once()

	provider := &ServiceProvider{}
	assert.NotPanics(t, func() { provider.Boot(mockApp) })
}
