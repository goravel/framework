package log

import (
	"bufio"
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/log/formatters"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/facades"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"os"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	addDefaultConfig()

	app := Application{}
	instance := app.Init()

	instance.WithFields(logrus.Fields{
		"goravel": "test",
	}).Debug("debug")

	instance.WithFields(logrus.Fields{
		"goravel": "test",
	}).Error("error")

	dailyFile := "storage/logs/goravel-" + time.Now().Format("2006-01-02") + ".log"
	singleFile := "storage/logs/goravel.log"
	singleErrorFile := "storage/logs/goravel-error.log"
	customFile := "storage/logs/goravel-custom.log"

	assert.FileExists(t, dailyFile)
	assert.FileExists(t, singleFile)
	assert.FileExists(t, singleErrorFile)
	assert.FileExists(t, customFile)

	assert.Equal(t, 2, getLineNum(dailyFile))
	assert.Equal(t, 2, getLineNum(singleFile))
	assert.Equal(t, 1, getLineNum(singleErrorFile))
	assert.Equal(t, 2, getLineNum(customFile))

	os.Remove(".env")
	os.RemoveAll("storage")
}

type CustomTest struct {
}

func (single CustomTest) Handle(configPath string) logrus.Hook {
	logPath := facades.Config.GetString(configPath + ".path")

	return lfshook.NewHook(
		logPath,
		&formatters.General{},
	)
}

//getLineNum Get file line num.
func getLineNum(fileName string) int {
	total := 0
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0444)
	if err != nil {
		log.Fatalln("Open file fail:", err.Error())
	}

	buf := bufio.NewReader(file)

	for {
		_, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalln("Read file fail:", err.Error())
			}
		}

		total++
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalln("Close file fail:", err.Error())
		}
	}()

	return total
}

//addDefaultConfig Add default config for test.
func addDefaultConfig() {
	support.CreateEnv()
	configApp := config.ServiceProvider{}
	configApp.Register()

	facadesConfig := facades.Config
	facadesConfig.Add("logging", map[string]interface{}{
		"default": facadesConfig.Env("LOG_CHANNEL", "stack"),
		"channels": map[string]interface{}{
			"stack": map[string]interface{}{
				"driver":   "stack",
				"channels": []string{"daily", "single", "single-error", "custom"},
			},
			"single": map[string]interface{}{
				"driver": "single",
				"path":   "storage/logs/goravel.log",
				"level":  "debug",
			},
			"single-error": map[string]interface{}{
				"driver": "single",
				"path":   "storage/logs/goravel-error.log",
				"level":  "error",
			},
			"daily": map[string]interface{}{
				"driver": "daily",
				"path":   "storage/logs/goravel.log",
				"level":  facadesConfig.Env("LOG_LEVEL", "debug"),
				"days":   7,
			},
			"custom": map[string]interface{}{
				"driver": "custom",
				"via":    CustomTest{},
				"path":   "storage/logs/goravel-custom.log",
				"level":  facadesConfig.Env("LOG_LEVEL", "debug"),
			},
		},
	})
}
