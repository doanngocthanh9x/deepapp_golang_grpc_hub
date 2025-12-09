package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func Init(level string) {
	log = logrus.New()
	log.SetOutput(os.Stdout)

	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

func Info(msg string, fields ...interface{}) {
	log.WithFields(logrus.Fields{"args": fields}).Info(msg)
}

func Warn(msg string, fields ...interface{}) {
	log.WithFields(logrus.Fields{"args": fields}).Warn(msg)
}

func Error(msg string, fields ...interface{}) {
	log.WithFields(logrus.Fields{"args": fields}).Error(msg)
}

func Debug(msg string, fields ...interface{}) {
	log.WithFields(logrus.Fields{"args": fields}).Debug(msg)
}