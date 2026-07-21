package broadcasting

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/broadcasting"
)

func TestPusherDriver_SignRequest(t *testing.T) {
	driver := &PusherDriver{
		key:    "test-key",
		secret: "test-secret",
		appID:  "test-app",
		options: broadcasting.PusherOptions{
			Cluster: "mt1",
			Host:    "api-mt1.pusher.com",
			Port:    443,
			Scheme:  "https",
		},
		client: &http.Client{},
	}

	body := []byte(`{"name":"test","channels":["my-channel"],"data":"{}"}`)

	req, err := http.NewRequest("POST", "https://api-mt1.pusher.com:443/apps/test-app/events", nil)
	assert.NoError(t, err)

	driver.signRequest(req, body)

	assert.Equal(t, "test-key", req.Header.Get("X-Pusher-Key"))
	assert.NotEmpty(t, req.Header.Get("X-Pusher-Signature"))
	assert.NotEmpty(t, req.Header.Get("X-Pusher-Timestamp"))

	timestamp := req.Header.Get("X-Pusher-Timestamp")
	bodyMD5 := fmt.Sprintf("%x", md5.Sum(body))
	stringToSign := fmt.Sprintf("POST\n%s\nauth_key=test-key&auth_timestamp=%s&auth_version=1.0&body_md5=%s",
		req.URL.Path, timestamp, bodyMD5)

	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write([]byte(stringToSign))
	expected := hex.EncodeToString(mac.Sum(nil))

	assert.Equal(t, expected, req.Header.Get("X-Pusher-Signature"))
}

func TestPusherDriver_Broadcast(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/apps/test-app/events", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("X-Pusher-Key"))
		assert.NotEmpty(t, r.Header.Get("X-Pusher-Signature"))

		w.WriteHeader(200)
	}))
	defer server.Close()

	// Parse the test server URL to get host and port
	// The URL format is http://127.0.0.1:PORT
	driver := &PusherDriver{
		key:    "test-key",
		secret: "test-secret",
		appID:  "test-app",
		options: broadcasting.PusherOptions{
			Host:   "127.0.0.1",
			Port:   parsePortFromURL(t, server.URL),
			Scheme: "http",
		},
		client: server.Client(),
	}

	err := driver.Broadcast(
		[]broadcasting.Channel{{Name: "my-channel"}},
		"test-event",
		map[string]any{"message": "hello"},
	)
	assert.NoError(t, err)
}

func TestPusherDriver_Broadcast_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	driver := &PusherDriver{
		key:    "test-key",
		secret: "test-secret",
		appID:  "test-app",
		options: broadcasting.PusherOptions{
			Host:   "127.0.0.1",
			Port:   parsePortFromURL(t, server.URL),
			Scheme: "http",
		},
		client: server.Client(),
	}

	err := driver.Broadcast(
		[]broadcasting.Channel{{Name: "my-channel"}},
		"test-event",
		map[string]any{"message": "hello"},
	)
	assert.Error(t, err)
}

func parsePortFromURL(t *testing.T, urlStr string) int {
	t.Helper()
	var port int
	fmt.Sscanf(urlStr, "http://127.0.0.1:%d", &port)
	return port
}
