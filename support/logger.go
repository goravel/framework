package support

import "github.com/sirupsen/logrus"

type Logger interface {
	Handle(configPath string) logrus.Hook
}