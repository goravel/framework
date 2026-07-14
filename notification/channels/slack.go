package channels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	contractsnotification "github.com/goravel/framework/contracts/notification"

	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/log"
)

// SlackChannel delivers notifications to Slack via incoming webhooks.
type SlackChannel struct {
	client *http.Client
	log    log.Log
	config contractsconfig.Config
}

func NewSlackChannel(logger log.Log, config contractsconfig.Config) *SlackChannel {
	return &SlackChannel{
		client: &http.Client{Timeout: 10 * time.Second},
		log:    logger,
		config: config,
	}
}

func (c *SlackChannel) Name() string { return "slack" }

func (c *SlackChannel) Send(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	return c.SendNow(notifiable, n)
}

// SendNow is identical to Send — SlackChannel has no queued mode of its
// own, only Manager does. Exists so callers can bypass Manager's queue
// routing entirely: facades.Notification().Channel("slack").SendNow(u, n).
func (c *SlackChannel) SendNow(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) error {
	route, payload, err := c.Resolve(notifiable, n)
	if err != nil {
		return err
	}
	return c.Deliver(route, payload)
}

func (c *SlackChannel) Resolve(
	notifiable contractsnotification.Notifiable,
	n contractsnotification.Notification,
) (string, []byte, error) {
	// RouteNotificationFor is the per-notifiable webhook. If the notifiable
	// doesn't provide one, fall back to the app-wide default configured in
	// config/notification.go (notification.channels.slack.webhook).
	webhookURL := notifiable.RouteNotificationFor("slack")
	if webhookURL == "" && c.config != nil {
		webhookURL = c.config.GetString("notification.channels.slack.webhook")
	}
	if webhookURL == "" {
		return "", nil, fmt.Errorf("slack channel: %T.RouteNotificationFor(\"slack\") returned empty webhook URL, and no default is configured", notifiable)
	}

	var msg contractsnotification.SlackMessage
	if sn, ok := n.(contractsnotification.SlackNotification); ok {
		msg = sn.ToSlack(notifiable)
	} else {
		msg = c.defaultMessage(n)
	}

	// Fill in bot identity from config when the notification didn't set one.
	if c.config != nil {
		if msg.Username == "" {
			msg.Username = c.config.GetString("notification.channels.slack.username")
		}
		if msg.IconEmoji == "" {
			msg.IconEmoji = c.config.GetString("notification.channels.slack.icon_emoji")
		}
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return "", nil, fmt.Errorf("slack channel: failed to marshal payload for %T: %w", n, err)
	}

	return webhookURL, payload, nil
}

func (c *SlackChannel) Deliver(route string, payload []byte) error {
	if route == "" {
		return nil
	}

	var msg contractsnotification.SlackMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return fmt.Errorf("slack channel: failed to unmarshal payload: %w", err)
	}

	body, err := json.Marshal(slackPayload(msg))
	if err != nil {
		return fmt.Errorf("slack channel: failed to marshal wire payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, route, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack channel: failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack channel: HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack channel: webhook returned non-2xx status %d", resp.StatusCode)
	}

	c.log.Debugf("notifications: delivered to Slack (status=%d)", resp.StatusCode)
	return nil
}

func (c *SlackChannel) defaultMessage(n contractsnotification.Notification) contractsnotification.SlackMessage {
	return contractsnotification.SlackMessage{
		Text: fmt.Sprintf("New notification: *%T*", n),
	}
}

// ---- internal wire types for the Slack Incoming Webhook API ----

type slackWirePayload struct {
	Text        string                `json:"text,omitempty"`
	Username    string                `json:"username,omitempty"`
	IconEmoji   string                `json:"icon_emoji,omitempty"`
	Channel     string                `json:"channel,omitempty"`
	Attachments []slackWireAttachment `json:"attachments,omitempty"`
}

type slackWireAttachment struct {
	Title  string           `json:"title,omitempty"`
	Text   string           `json:"text,omitempty"`
	Color  string           `json:"color,omitempty"`
	Fields []slackWireField `json:"fields,omitempty"`
	Footer string           `json:"footer,omitempty"`
	Ts     int64            `json:"ts,omitempty"`
}

type slackWireField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func slackPayload(msg contractsnotification.SlackMessage) slackWirePayload {
	p := slackWirePayload{
		Text:      msg.Text,
		Username:  msg.Username,
		IconEmoji: msg.IconEmoji,
		Channel:   msg.Channel,
	}
	for _, a := range msg.Attachments {
		wa := slackWireAttachment{
			Title:  a.Title,
			Text:   a.Text,
			Color:  a.Color,
			Footer: a.Footer,
			Ts:     a.Timestamp,
		}
		for _, f := range a.Fields {
			wa.Fields = append(wa.Fields, slackWireField{Title: f.Title, Value: f.Value, Short: f.Short})
		}
		p.Attachments = append(p.Attachments, wa)
	}
	return p
}

var _ contractsnotification.ResolvableChannel = (*SlackChannel)(nil)
