package broadcasting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconfig "github.com/goravel/framework/mocks/config"
)

func TestConfig_DefaultConnection(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("broadcasting.default", "log").Return("pusher").Once()

	cfg := NewConfig(mockConfig)
	assert.Equal(t, "pusher", cfg.DefaultConnection())
}

func TestConfig_Connection(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().Get("broadcasting.connections.pusher").Return(map[string]any{
		"driver":  "pusher",
		"key":     "app-key",
		"secret":  "app-secret",
		"app_id":  "app-id",
		"options": map[string]any{
			"cluster": "mt1",
			"host":    "api-mt1.pusher.com",
			"port":    float64(443),
			"scheme":  "https",
		},
	}).Once()

	cfg := NewConfig(mockConfig)
	conn, err := cfg.Connection("pusher")
	assert.NoError(t, err)
	assert.Equal(t, "pusher", conn.Driver)
	assert.Equal(t, "app-key", conn.Key)
	assert.Equal(t, "app-secret", conn.Secret)
	assert.Equal(t, "app-id", conn.AppID)
	assert.Equal(t, "mt1", conn.Options.Cluster)
	assert.Equal(t, "api-mt1.pusher.com", conn.Options.Host)
	assert.Equal(t, 443, conn.Options.Port)
	assert.Equal(t, "https", conn.Options.Scheme)
}

func TestConfig_Connection_NotFound(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().Get("broadcasting.connections.missing").Return(nil).Once()

	cfg := NewConfig(mockConfig)
	_, err := cfg.Connection("missing")
	assert.Error(t, err)
}

func TestConfig_Auth(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetBool("broadcasting.auth.enabled", true).Return(true).Once()
	mockConfig.EXPECT().GetString("broadcasting.auth.path", "/broadcasting/auth").Return("/custom/auth").Once()
	mockConfig.EXPECT().GetStringSlice("broadcasting.auth.middleware", mock.Anything).Return([]string{"web"}).Once()

	cfg := NewConfig(mockConfig)
	auth := cfg.Auth()
	assert.True(t, auth.Enabled)
	assert.Equal(t, "/custom/auth", auth.Path)
	assert.Equal(t, []string{"web"}, auth.Middleware)
}

func TestConfig_DefaultValues(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("broadcasting.default", "log").Return("log").Once()
	mockConfig.EXPECT().GetBool("broadcasting.auth.enabled", true).Return(false).Once()
	mockConfig.EXPECT().GetString("broadcasting.auth.path", "/broadcasting/auth").Return("/broadcasting/auth").Once()
	mockConfig.EXPECT().GetStringSlice("broadcasting.auth.middleware", mock.Anything).Return([]string{"web"}).Once()

	cfg := NewConfig(mockConfig)
	assert.Equal(t, "log", cfg.DefaultConnection())

	auth := cfg.Auth()
	assert.False(t, auth.Enabled)
	assert.Equal(t, "/broadcasting/auth", auth.Path)
	assert.Equal(t, []string{"web"}, auth.Middleware)
}
