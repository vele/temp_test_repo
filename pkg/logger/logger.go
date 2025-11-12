package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func New(level logrus.Level) *logrus.Logger {
	log := logrus.New()
	log.Out = os.Stdout
	log.SetLevel(level)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return log
}
