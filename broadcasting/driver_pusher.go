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
	"net/url"
	"strconv"
	"time"

	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/errors"
)

type PusherDriver struct {
	key     string
	secret  string
	appID   string
	options broadcasting.PusherOptions
	client  *http.Client
	baseURL string
}

func NewPusherDriver(conn broadcasting.ConnectionConfig) *PusherDriver {
	opts := conn.Options
	host := opts.Host
	if host == "" && opts.Cluster != "" {
		host = fmt.Sprintf("api-%s.pusher.com", opts.Cluster)
	}

	baseURL := fmt.Sprintf("%s://%s:%d/apps/%s",
		opts.Scheme, host, opts.Port, conn.AppID)

	return &PusherDriver{
		key:     conn.Key,
		secret:  conn.Secret,
		appID:   conn.AppID,
		options: opts,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
	}
}

func (d *PusherDriver) Broadcast(channels []broadcasting.Channel, event string, payload map[string]any) error {
	url := fmt.Sprintf("%s/events", d.baseURL)

	chanNames := make([]string, len(channels))
	for i, ch := range channels {
		chanNames[i] = ch.Name
	}

	dataJSON, err := json.Marshal(payload)
	if err != nil {
		return errors.BroadcastPusherMarshalPayloadFailed.Args(err)
	}

	body := map[string]any{
		"name":     event,
		"channels": chanNames,
		"data":     string(dataJSON),
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return errors.BroadcastPusherMarshalBodyFailed.Args(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return errors.BroadcastPusherCreateRequestFailed.Args(err)
	}
	req.Header.Set("Content-Type", "application/json")
	d.signRequest(req, bodyBytes)

	resp, err := d.client.Do(req)
	if err != nil {
		return errors.BroadcastPusherRequestFailed.Args(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return errors.BroadcastPusherHTTPError.Args(resp.StatusCode, url)
	}

	return nil
}

func (d *PusherDriver) signRequest(req *http.Request, body []byte) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	bodyMD5 := fmt.Sprintf("%x", md5.Sum(body))

	stringToSign := fmt.Sprintf("POST\n%s\nauth_key=%s&auth_timestamp=%s&auth_version=1.0&body_md5=%s",
		req.URL.Path, d.key, timestamp, bodyMD5)

	mac := hmac.New(sha256.New, []byte(d.secret))
	_, _ = mac.Write([]byte(stringToSign))
	signature := hex.EncodeToString(mac.Sum(nil))

	q := url.Values{}
	q.Set("auth_key", d.key)
	q.Set("auth_timestamp", timestamp)
	q.Set("auth_version", "1.0")
	q.Set("body_md5", bodyMD5)
	q.Set("auth_signature", signature)
	req.URL.RawQuery = q.Encode()
}
