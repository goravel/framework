package broadcasters

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/broadcasting"
	mockshttp "github.com/goravel/framework/mocks/http/client"
)

func TestPushOptionsFromConfig(t *testing.T) {
	opts := PushOptionsFromConfig(map[string]any{
		"cluster": "mt1",
		"host":    "api.example.com",
		"port":    float64(8080),
		"scheme":  "http",
	})
	assert.Equal(t, "mt1", opts.Cluster)
	assert.Equal(t, "api.example.com", opts.Host)
	assert.Equal(t, 8080, opts.Port)
	assert.Equal(t, "http", opts.Scheme)
}

func TestPushOptionsFromConfig_Defaults(t *testing.T) {
	opts := PushOptionsFromConfig(map[string]any{
		"cluster": "eu",
	})
	assert.Equal(t, "eu", opts.Cluster)
	assert.Equal(t, 443, opts.Port)
	assert.Equal(t, "https", opts.Scheme)
	assert.Equal(t, "", opts.Host)
}

func TestPushOptionsFromConfig_HostNormalization(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{name: "bare host", host: "api.pusher.com", want: "api.pusher.com"},
		{name: "scheme stripped", host: "https://api.pusher.com", want: "api.pusher.com"},
		{name: "scheme + port stripped", host: "https://api.pusher.com:443", want: "api.pusher.com"},
		{name: "explicit port stripped", host: "api.pusher.com:8080", want: "api.pusher.com"},
		{name: "whitespace trimmed", host: "  api.pusher.com  ", want: "api.pusher.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := PushOptionsFromConfig(map[string]any{"host": tt.host, "port": float64(443)})
			assert.Equal(t, tt.want, opts.Host)
		})
	}
}

func TestPusherDriver_ClusterFallback(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver: "pusher",
		Key:    "key",
		Secret: "secret",
		AppID:  "app",
		Options: map[string]any{
			"cluster": "eu",
			"port":    float64(443),
			"scheme":  "https",
		},
	}

	mockClient := mockshttp.NewFactory(t)
	driver, err := NewPusherDriver(conn, mockClient)
	assert.NoError(t, err)
	assert.Contains(t, driver.baseURL, "api-eu.pusher.com")
}

func TestPusherDriver_SignParams(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver: "pusher",
		Key:    "test-key",
		Secret: "test-secret",
		AppID:  "test-app",
		Options: map[string]any{
			"cluster": "mt1",
			"host":    "api-mt1.pusher.com",
			"port":    float64(443),
			"scheme":  "https",
		},
	}

	mockClient := mockshttp.NewFactory(t)
	driver, err := NewPusherDriver(conn, mockClient)
	assert.NoError(t, err)

	body := []byte(`{"name":"test","channels":["my-channel"],"data":"{}"}`)
	params := driver.signParams(body)

	assert.Equal(t, "test-key", params["auth_key"])
	assert.NotEmpty(t, params["auth_signature"])
	assert.NotEmpty(t, params["auth_timestamp"])
	assert.Equal(t, "1.0", params["auth_version"])

	timestamp := params["auth_timestamp"]
	bodyMD5 := fmt.Sprintf("%x", md5.Sum(body))
	stringToSign := fmt.Sprintf("POST\n/apps/test-app/events\nauth_key=test-key&auth_timestamp=%s&auth_version=1.0&body_md5=%s",
		timestamp, bodyMD5)

	mac := hmac.New(sha256.New, []byte("test-secret"))
	_, _ = mac.Write([]byte(stringToSign))
	expected := hex.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expected, params["auth_signature"])
}

func TestPusherDriver_Broadcast_Success(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver: "pusher",
		Key:    "test-key",
		Secret: "test-secret",
		AppID:  "test-app",
		Options: map[string]any{
			"host":   "api.example.com",
			"port":   float64(443),
			"scheme": "https",
		},
	}

	mockResponse := mockshttp.NewResponse(t)
	mockResponse.EXPECT().Failed().Return(false).Once()

	mockRequest := mockshttp.NewRequest(t)
	mockRequest.EXPECT().WithHeader("Content-Type", "application/json").Return(mockRequest).Once()
	mockRequest.EXPECT().WithQueryParameters(mock.MatchedBy(func(p map[string]string) bool {
		return p["auth_key"] == "test-key" && p["auth_version"] == "1.0"
	})).Return(mockRequest).Once()
	mockRequest.EXPECT().WithContext(context.Background()).Return(mockRequest).Once()
	mockRequest.EXPECT().Post("https://api.example.com:443/apps/test-app/events", mock.MatchedBy(func(r any) bool { return r != nil })).Return(mockResponse, nil).Once()

	mockClient := mockshttp.NewFactory(t)
	mockClient.EXPECT().Client().Return(mockRequest).Once()

	driver, err := NewPusherDriver(conn, mockClient)
	assert.NoError(t, err)

	err = driver.Broadcast(
		context.Background(),
		[]string{"my-channel"},
		"test-event",
		map[string]any{"message": "hello"},
	)
	assert.NoError(t, err)
}

func TestPusherDriver_Broadcast_Error(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver: "pusher",
		Key:    "test-key",
		Secret: "test-secret",
		AppID:  "test-app",
		Options: map[string]any{
			"host":   "api.example.com",
			"port":   float64(443),
			"scheme": "https",
		},
	}

	mockResponse := mockshttp.NewResponse(t)
	mockResponse.EXPECT().Failed().Return(true).Once()
	mockResponse.EXPECT().Status().Return(http.StatusInternalServerError).Once()

	mockRequest := mockshttp.NewRequest(t)
	mockRequest.EXPECT().WithHeader("Content-Type", "application/json").Return(mockRequest).Once()
	mockRequest.EXPECT().WithQueryParameters(mock.MatchedBy(func(p map[string]string) bool {
		return p["auth_key"] == "test-key" && p["auth_version"] == "1.0"
	})).Return(mockRequest).Once()
	mockRequest.EXPECT().WithContext(context.Background()).Return(mockRequest).Once()
	mockRequest.EXPECT().Post("https://api.example.com:443/apps/test-app/events", mock.MatchedBy(func(r any) bool { return r != nil })).Return(mockResponse, nil).Once()

	mockClient := mockshttp.NewFactory(t)
	mockClient.EXPECT().Client().Return(mockRequest).Once()

	driver, err := NewPusherDriver(conn, mockClient)
	assert.NoError(t, err)

	err = driver.Broadcast(
		context.Background(),
		[]string{"my-channel"},
		"test-event",
		map[string]any{"message": "hello"},
	)
	assert.Error(t, err)
}

func TestPusherDriver_ErrorOnEmptyHost(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver: "pusher",
		Key:    "key",
		Secret: "secret",
		AppID:  "app",
		Options: map[string]any{
			"port":   float64(443),
			"scheme": "https",
		},
	}

	mockClient := mockshttp.NewFactory(t)

	driver, err := NewPusherDriver(conn, mockClient)
	assert.Error(t, err)
	assert.Nil(t, driver)
}

func TestBroadcasterNewPusherDriver(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver: "pusher",
		Key:    "key",
		Secret: "secret",
		AppID:  "app",
		Options: map[string]any{
			"cluster": "mt1",
			"port":    float64(443),
			"scheme":  "https",
		},
	}

	mockClient := mockshttp.NewFactory(t)
	driver, err := NewPusherDriver(conn, mockClient)
	assert.NoError(t, err)
	assert.Equal(t, "key", driver.key)
	assert.Equal(t, "secret", driver.secret)
	assert.Equal(t, "app", driver.appID)
	assert.Equal(t, "mt1", driver.options.Cluster)
}

func TestBroadcasterNewLogDriver(t *testing.T) {
	driver := new(LogDriver)
	assert.NotNil(t, driver)
}

func TestBroadcasterNewNullDriver(t *testing.T) {
	driver := NewNullDriver()
	assert.NotNil(t, driver)
	err := driver.Broadcast(context.Background(), nil, "", nil)
	assert.NoError(t, err)
}
