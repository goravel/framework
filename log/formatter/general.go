package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/facades"
)

type General struct {
}

func (general *General) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	cstSh, err := time.LoadLocation(facades.Config.GetString("app.timezone"))
	if err != nil {
		return nil, err
	}

	timestamp := entry.Time.In(cstSh).Format("2006-01-02 15:04:05")
	var newLog string

	if len(entry.Data) > 0 {
		data, _ := json.Marshal(entry.Data)
		newLog = fmt.Sprintf("[%s] %s.%s: %s %s\n", timestamp, facades.Config.GetString("app.env"), entry.Level, entry.Message, string(data))
	} else {
		newLog = fmt.Sprintf("[%s] %s.%s: %s\n", timestamp, facades.Config.GetString("app.env"), entry.Level, entry.Message)
	}

	b.WriteString(newLog)

	return b.Bytes(), nil
}
