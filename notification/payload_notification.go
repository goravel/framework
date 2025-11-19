package notification

import (
    contractsnotification "github.com/goravel/framework/contracts/notification"
)

type PayloadNotification struct {
    Channels []string
    Payloads map[string]map[string]interface{} // channel -> payload map
}

func (n PayloadNotification) Via(_ contractsnotification.Notifiable) []string {
    return n.Channels
}

func (n PayloadNotification) PayloadFor(channel string, _ contractsnotification.Notifiable) (map[string]interface{}, error) {
    if n.Payloads == nil {
        return map[string]interface{}{}, nil
    }
    m := n.Payloads[channel]
    if m == nil {
        return map[string]interface{}{}, nil
    }
    return m, nil
}