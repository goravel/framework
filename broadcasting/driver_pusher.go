package broadcasting

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/goravel/framework/contracts/broadcasting"
)

type PusherDriver struct {
	key     string
	secret  string
	appID   string
	options broadcasting.PusherOptions
	client  *http.Client
}

func NewPusherDriver(conn broadcasting.ConnectionConfig) *PusherDriver {
	return &PusherDriver{
		key:     conn.Key,
		secret:  conn.Secret,
		appID:   conn.AppID,
		options: conn.Options,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (d *PusherDriver) Broadcast(channels []broadcasting.Channel, event string, payload map[string]any) error {
	url := fmt.Sprintf("%s://%s:%d/apps/%s/events",
		d.options.Scheme, d.options.Host, d.options.Port, d.appID)

	chanNames := make([]string, len(channels))
	for i, ch := range channels {
		chanNames[i] = ch.Name
	}

	dataJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("pusher: failed to marshal payload: %w", err)
	}

	body := map[string]any{
		"name":     event,
		"channels": chanNames,
		"data":     string(dataJSON),
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pusher: failed to marshal body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("pusher: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	d.signRequest(req, bodyBytes)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("pusher: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("pusher: HTTP %d: request to %s failed", resp.StatusCode, url)
	}

	return nil
}

func (d *PusherDriver) signRequest(req *http.Request, body []byte) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	bodyMD5 := fmt.Sprintf("%x", md5.Sum(body))

	stringToSign := fmt.Sprintf("POST\n%s\nauth_key=%s&auth_timestamp=%s&auth_version=1.0&body_md5=%s",
		req.URL.Path, d.key, timestamp, bodyMD5)

	mac := hmac.New(sha256.New, []byte(d.secret))
	mac.Write([]byte(stringToSign))
	signature := hex.EncodeToString(mac.Sum(nil))

	req.Header.Set("X-Pusher-Key", d.key)
	req.Header.Set("X-Pusher-Signature", signature)
	req.Header.Set("X-Pusher-Timestamp", timestamp)
}
