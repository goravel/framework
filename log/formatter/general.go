package formatter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

type General struct {
	config config.Config
	json   foundation.Json
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

func NewGeneral(config config.Config, json foundation.Json) *General {
	return &General{
		config: config,
		json:   json,
	}
}

func (general *General) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := carbon.FromStdTime(entry.Time).ToDateTimeMilliString()
	fmt.Fprintf(b, "[%s] %s.%s: %s\n", timestamp, general.config.GetString("app.env"), entry.Level, entry.Message)
	data := entry.Data
	if len(data) > 0 {
		formattedData, err := general.formatData(data)
		if err != nil {
			return nil, err
		}
		b.WriteString(formattedData)
	}

	return b.Bytes(), nil
}

func (general *General) formatData(data logrus.Fields) (string, error) {
	var builder strings.Builder

	if len(data) > 0 {
		removedData := deleteKey(data, "root")
		if len(removedData) > 0 {
			removedDataStr, err := general.json.MarshalString(removedData)
			if err != nil {
				return "", err
			}

			builder.WriteString(fmt.Sprintf("fields: %s\n", removedDataStr))
		}

		root, err := cast.ToStringMapE(data["root"])
		if err != nil {
			return "", err
		}

		for _, key := range []string{"hint", "tags", "owner", "context", "with", "domain", "code", "request", "response", "user"} {
			if value, exists := root[key]; exists && value != nil {
				builder.WriteString(fmt.Sprintf("[%s] %+v\n", str.Of(key).UcFirst().String(), value))
			}
		}

		if stackTraceValue, exists := root["stacktrace"]; exists && stackTraceValue != nil {
			traces, err := general.formatStackTraces(stackTraceValue)
			if err != nil {
				return "", err
			}

			builder.WriteString(traces)
		}
	}

	return builder.String(), nil
}

func (general *General) formatStackTraces(stackTraces any) (string, error) {
	var formattedTraces strings.Builder
	data, err := general.json.Marshal(stackTraces)

	if err != nil {
		return "", err
	}
	var traces StackTrace
	err = general.json.Unmarshal(data, &traces)
	if err != nil {
		return "", err
	}
	formattedTraces.WriteString("[Trace]\n")
	root := traces.Root
	if len(root.Stack) > 0 {
		for _, stackStr := range root.Stack {
			formattedTraces.WriteString(formatStackTrace(stackStr))
		}
	}

	return formattedTraces.String(), nil
}

func formatStackTrace(stackStr string) string {
	lastColon := strings.LastIndex(stackStr, ":")
	if lastColon > 0 && lastColon < len(stackStr)-1 {
		secondLastColon := strings.LastIndex(stackStr[:lastColon], ":")
		if secondLastColon > 0 {
			fileLine := stackStr[secondLastColon+1:]
			method := stackStr[:secondLastColon]
			return fmt.Sprintf("%s [%s]\n", fileLine, method)
		}
	}
	return fmt.Sprintf("%s\n", stackStr)
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
