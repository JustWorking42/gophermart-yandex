package logger

import (
	"go.uber.org/zap"
)

func CreateLogger(textLevel string) (*zap.Logger, error) {
	level, err := zap.ParseAtomicLevel(textLevel)
	if err != nil {
		return nil, err
	}

	config := zap.NewDevelopmentConfig()
	config.Level = level
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
