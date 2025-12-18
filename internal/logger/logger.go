package logger

import "github.com/sirupsen/logrus"

// NewLogger creates a logrus logger with JSON formatting.
// Returns configured logger or error for invalid level.
func NewLogger(level string) (*logrus.Logger, error) {
	var err error

	logLevel := logrus.InfoLevel
	if level != "" {
		logLevel, err = logrus.ParseLevel(level)
		if err != nil {
			return nil, err
		}
	}

	logger := logrus.New()
	logger.SetLevel(logLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	return logger, nil
}
