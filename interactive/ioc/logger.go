package ioc

import (
	"red-feed/pkg/logger"

	"go.uber.org/zap"
)

func InitLogger() logger.Logger {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	l := logger.NewZapLogger(zapLogger)
	return l
}
