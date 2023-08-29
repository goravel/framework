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

	newLog = fmt.Sprintf("[%s] %s.%s: %s\n", timestamp, general.config.GetString("app.env"), entry.Level, entry.Message)
	if len(entry.Data) > 0 {
		data, _ := sonic.Marshal(entry.Data)
		root, _ := sonic.Get(data, "root")

		delete(entry.Data, "root")
		if len(entry.Data) > 0 {
			removedData, _ := sonic.Marshal(entry.Data)
			newLog = fmt.Sprintf("Fields: %s\n", string(removedData))
		}

		code := root.Get("code")
		if code.Valid() {
			codeInfo, _ := code.Raw()
			newLog += fmt.Sprintf("Code: %s\n", codeInfo)
		}

		context := root.Get("context")
		if context.Valid() {
			contextInfo, _ := context.Raw()
			newLog += fmt.Sprintf("Context: %s\n", contextInfo)
		}

		domain := root.Get("domain")
		if domain.Valid() {
			domainInfo, _ := domain.Raw()
			newLog += fmt.Sprintf("Domain: %s\n", domainInfo)
		}

		hint := root.Get("hint")
		if hint.Valid() {
			hintInfo, _ := hint.Raw()
			newLog += fmt.Sprintf("Hint: %s\n", hintInfo)
		}

		owner := root.Get("owner")
		if owner.Valid() {
			ownerInfo, _ := owner.Raw()
			newLog += fmt.Sprintf("Owner: %s\n", ownerInfo)
		}

		span := root.Get("span")
		if span.Valid() {
			spanInfo, _ := span.Raw()
			newLog += fmt.Sprintf("Span: %s\n", spanInfo)
		}

		tags := root.Get("tags")
		if tags.Valid() {
			tagsInfo, _ := tags.Raw()
			newLog += fmt.Sprintf("Tags: %s\n", tagsInfo)
		}

		user := root.Get("user")
		if user.Valid() {
			userInfo, _ := user.Raw()
			newLog += fmt.Sprintf("User: %s\n", userInfo)
		}

		tracks := root.Get("stacktrace")
		if tracks.Valid() {
			tracksInfo, _ := tracks.Interface()
			newLog += general.formatStackTraces(tracksInfo)
		}
	}

	b.WriteString(newLog)

	return b.Bytes(), nil
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

func (general *General) formatStackTraces(stackTraces interface{}) string {
	var formattedTraces string
	data, _ := sonic.Marshal(stackTraces)
	var traces StackTrace
	err := sonic.Unmarshal(data, &traces)
	if err != nil {
		return ""
	}
	formattedTraces += "Trace:\n"
	root := traces.Root
	if len(root.Stack) > 0 {
		formattedTraces += "\t" + root.Message + "\n"
		for _, stackStr := range root.Stack {
			formattedTraces += fmt.Sprintf("\t\t%s\n", stackStr)
		}
	}

	for _, wrap := range traces.Wrap {
		formattedTraces += "\t" + wrap.Message + "\n"
		if wrap.Stack != "" {
			formattedTraces += fmt.Sprintf("\t\t%s\n", wrap.Stack)
		}
	}

	return formattedTraces
}
