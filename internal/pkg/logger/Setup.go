package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	Logger logrus.Logger
)

func InitLogger() error {
	logFileName := fmt.Sprintf("app_%s.log", time.Now().Format("20060102150405"))
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	Logger = *logrus.New()
	Logger.SetOutput(file)
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	Logger.Info("Logging started")

	return nil
}
