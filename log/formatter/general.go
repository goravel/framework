package formatter

import (
	"bytes"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/config"
)

type General struct {
	config config.Config
}

func NewGeneral(config config.Config) *General {
	return &General{
		config: config,
	}
}

func (general *General) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	cstSh, err := time.LoadLocation(general.config.GetString("app.timezone"))
	if err != nil {
		return nil, err
	}

	timestamp := entry.Time.In(cstSh).Format("2006-01-02 15:04:05")
	var newLog string

	if len(entry.Data) > 0 {
		data, _ := sonic.Marshal(entry.Data)
		newLog = fmt.Sprintf("[%s] %s.%s: %s %s\n", timestamp, general.config.GetString("app.env"), entry.Level, entry.Message, string(data))
	} else {
		newLog = fmt.Sprintf("[%s] %s.%s: %s\n", timestamp, general.config.GetString("app.env"), entry.Level, entry.Message)
	}

	b.WriteString(newLog)

	return b.Bytes(), nil
}
