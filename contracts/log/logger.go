package log

import "github.com/sirupsen/logrus"

type Logger interface {
	Handle(configPath string) (logrus.Hook, error)
}
