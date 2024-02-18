package formatter

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/support/json"
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
	b.WriteString(fmt.Sprintf("[%s] %s.%s: %s\n", timestamp, general.config.GetString("app.env"), entry.Level, entry.Message))
	data := entry.Data
	if len(data) > 0 {
		formattedData, err := formatData(data)
		if err != nil {
			return nil, err
		}
		b.WriteString(formattedData)
	}

	return b.Bytes(), nil
}

func formatData(data logrus.Fields) (string, error) {
	var builder strings.Builder

	if len(data) > 0 {
		removedData := deleteKey(data, "root")
		if len(removedData) > 0 {
			removedDataBytes, err := json.Marshal(removedData)
			if err != nil {
				return "", err
			}

			builder.WriteString(fmt.Sprintf("fields: %s\n", string(removedDataBytes)))
		}

		root, err := cast.ToStringMapE(data["root"])
		if err != nil {
			return "", err
		}

		for _, key := range []string{"code", "context", "domain", "hint", "owner", "request", "response", "tags", "user"} {
			if value, exists := root[key]; exists && value != nil {
				v, err := json.Marshal(value)
				if err != nil {
					return "", err
				}

				builder.WriteString(fmt.Sprintf(`%s: %v"\n`, key, string(v)))
			}
		}

		if stackTraceValue, exists := root["stacktrace"]; exists && stackTraceValue != nil {
			traces, err := formatStackTraces(stackTraceValue)
			if err != nil {
				return "", err
			}

			builder.WriteString(traces)
		}
	}

	return builder.String(), nil
}

func deleteKey(data logrus.Fields, keyToDelete string) logrus.Fields {
	dataCopy := make(logrus.Fields)
	for key, value := range data {
		if key != keyToDelete {
			dataCopy[key] = value
		}
	}

	return dataCopy
}

type StackTrace struct {
	Root struct {
		Message string   `json:"message"`
		Stack   []string `json:"stack"`
	} `json:"root"`
	Wrap []struct {
		Message string `json:"message"`
		Stack   string `json:"stack"`
	} `json:"wrap"`
}

func formatStackTraces(stackTraces any) (string, error) {
	var formattedTraces strings.Builder
	data, err := json.Marshal(stackTraces)

	if err != nil {
		return "", err
	}
	var traces StackTrace
	err = json.Unmarshal(data, &traces)
	if err != nil {
		return "", err
	}
	formattedTraces.WriteString("trace:\n")
	root := traces.Root
	if len(root.Stack) > 0 {
		for _, stackStr := range root.Stack {
			formattedTraces.WriteString(fmt.Sprintf("\t%s\n", stackStr))
		}
	}

	return formattedTraces.String(), nil
}
