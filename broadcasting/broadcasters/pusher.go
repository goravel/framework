package broadcasters

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/errors"
)

type PusherDriver struct {
	key     string
	secret  string
	appID   string
	options broadcasting.PusherOptions
	client  client.Factory
	baseURL string
}

func NewPusherDriver(conn broadcasting.ConnectionConfig, httpClient client.Factory) (*PusherDriver, error) {
	opts := PushOptionsFromConfig(conn.Options)

	host := opts.Host
	if host == "" && opts.Cluster != "" {
		host = fmt.Sprintf("api-%s.pusher.com", opts.Cluster)
	}
	if host == "" {
		return nil, errors.BroadcastPusherHostRequired
	}

	baseURL := fmt.Sprintf("%s://%s:%d/apps/%s",
		opts.Scheme, host, opts.Port, conn.AppID)

	return &PusherDriver{
		key:     conn.Key,
		secret:  conn.Secret,
		appID:   conn.AppID,
		options: opts,
		client:  httpClient,
		baseURL: baseURL,
	}, nil
}

func PushOptionsFromConfig(options map[string]any) broadcasting.PusherOptions {
	opts := broadcasting.PusherOptions{
		Port:   443,
		Scheme: "https",
	}

	if v, ok := options["cluster"].(string); ok {
		opts.Cluster = v
	}
	if v, ok := options["host"].(string); ok {
		opts.Host = v
	}
	if v, ok := options["port"].(float64); ok {
		opts.Port = int(v)
	}
	if v, ok := options["scheme"].(string); ok {
		opts.Scheme = v
	}

	return opts
}

func (d *PusherDriver) Broadcast(channels []broadcasting.Channel, event string, payload map[string]any) error {
	urlStr := fmt.Sprintf("%s/events", d.baseURL)

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

	req := d.client.Client().
		WithHeader("Content-Type", "application/json").
		WithQueryParameters(d.signParams(bodyBytes))

	resp, err := req.Post(urlStr, bytes.NewReader(bodyBytes))
	if err != nil {
		return errors.BroadcastPusherRequestFailed.Args(err)
	}

	if resp.Failed() {
		return errors.BroadcastPusherHTTPError.Args(resp.Status(), urlStr)
	}

	return nil
}

func (d *PusherDriver) signParams(body []byte) map[string]string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	bodyMD5 := fmt.Sprintf("%x", md5.Sum(body))

	path := fmt.Sprintf("/apps/%s/events", d.appID)
	stringToSign := fmt.Sprintf("POST\n%s\nauth_key=%s&auth_timestamp=%s&auth_version=1.0&body_md5=%s",
		path, d.key, timestamp, bodyMD5)

	mac := hmac.New(sha256.New, []byte(d.secret))
	_, _ = mac.Write([]byte(stringToSign))
	signature := hex.EncodeToString(mac.Sum(nil))

	return map[string]string{
		"auth_key":        d.key,
		"auth_timestamp":  timestamp,
		"auth_version":    "1.0",
		"body_md5":        bodyMD5,
		"auth_signature":  signature,
	}
}
