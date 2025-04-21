package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	_logger, err := zap.NewProduction(
		zap.AddStacktrace(zap.PanicLevel),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		panic(err)
	}
	logger = _logger
}

func Errorf(msg string, args ...any) {
	logger.Error(fmt.Sprintf(msg, args...))
}

func Fatalf(msg string, args ...any) {
	logger.Fatal(fmt.Sprintf(msg, args...))
}

func Info(msg string) {
	logger.Info(msg)
}

func Infof(msg string, args ...any) {
	logger.Info(fmt.Sprintf(msg, args...))
}
