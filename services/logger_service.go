package services

import "github.com/sirupsen/logrus"

type Fields map[string]interface{}

type Logger interface {
	Debug(message string, fields Fields)
	Info(message string, fields Fields)
	Warn(message string, fields Fields)
	Error(message string, fields Fields)
}

type LogrusLogger struct {
	Logger *logrus.Logger
}

func (*LogrusLogger) Debug(message string, fields Fields) {
	logrus.WithFields(logrus.Fields(fields)).Debug(message)
}

func (*LogrusLogger) Info(message string, fields Fields) {
	logrus.WithFields(logrus.Fields(fields)).Info(message)
}

func (*LogrusLogger) Warn(message string, fields Fields) {
	logrus.WithFields(logrus.Fields(fields)).Warn(message)
}

func (*LogrusLogger) Error(message string, fields Fields) {
	logrus.WithFields(logrus.Fields(fields)).Error(message)
}

func NewLogrusLogger(logger *logrus.Logger) *LogrusLogger {
	return &LogrusLogger{
		Logger: logger,
	}
}
