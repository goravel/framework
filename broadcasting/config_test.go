package broadcasting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/broadcasting"
	contractshttp "github.com/goravel/framework/contracts/http"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockshttp "github.com/goravel/framework/mocks/http"
)

func TestNewConfig(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().UnmarshalKey("broadcasting", mock.MatchedBy(func(c *Config) bool { return true })).Return(nil).Once()

	cfg, err := NewConfig(mockConfig)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestConfig_DefaultConnection(t *testing.T) {
	cfg := &Config{
		Default: "pusher",
		Connections: map[string]broadcasting.ConnectionConfig{
			"pusher": {Driver: "pusher"},
		},
	}
	assert.Equal(t, "pusher", cfg.DefaultConnection())

	empty := &Config{}
	assert.Equal(t, "log", empty.DefaultConnection())
}

func TestConfig_Connection(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver: "pusher",
		Key:    "app-key",
		Secret: "app-secret",
		AppID:  "app-id",
		Options: map[string]any{
			"cluster": "mt1",
			"host":    "api-mt1.pusher.com",
			"port":    float64(443),
			"scheme":  "https",
		},
	}

	cfg := &Config{
		Connections: map[string]broadcasting.ConnectionConfig{
			"pusher": conn,
		},
	}

	got, err := cfg.Connection("pusher")
	assert.NoError(t, err)
	assert.Equal(t, conn.Driver, got.Driver)
	assert.Equal(t, conn.Key, got.Key)
	assert.Equal(t, conn.Secret, got.Secret)
	assert.Equal(t, conn.AppID, got.AppID)
	assert.Equal(t, "mt1", got.Options["cluster"])
	assert.Equal(t, "https", got.Options["scheme"])
}

func TestConfig_Connection_NotFound(t *testing.T) {
	cfg := &Config{
		Connections: map[string]broadcasting.ConnectionConfig{},
	}

	_, err := cfg.Connection("missing")
	assert.Error(t, err)
}

func TestConfig_Auth(t *testing.T) {
	cfg := &Config{
		Auth: broadcasting.AuthConfig{
			Enabled: true,
			Path:    "/custom/auth",
		},
	}
	assert.True(t, cfg.Auth.Enabled)
	assert.Equal(t, "/custom/auth", cfg.Auth.Path)
}

func TestConfig_DefaultValues(t *testing.T) {
	cfg := &Config{}
	assert.Equal(t, "log", cfg.DefaultConnection())

	_, err := cfg.Connection("anything")
	assert.Error(t, err)
}

func TestConfig_AuthMiddleware(t *testing.T) {
	mw := mockshttp.NewMiddleware(t)
	cfg := &Config{
		Auth: broadcasting.AuthConfig{
			Enabled:    true,
			Path:       "/broadcasting/auth",
			Middleware: []contractshttp.Middleware{mw},
		},
	}
	assert.True(t, cfg.Auth.Enabled)
	assert.Len(t, cfg.Auth.Middleware, 1)
	assert.Same(t, mw, cfg.Auth.Middleware[0])
}
