// internal/logging/logging.go
package logging

import (
	"github.com/sirupsen/logrus"
)

func Init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: false,
		ForceColors:   true,
	})
	logrus.SetLevel(logrus.InfoLevel)
}
