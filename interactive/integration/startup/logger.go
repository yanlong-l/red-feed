package startup

import (
	"red-feed/pkg/logger"
)

func InitTestLogger() logger.Logger {
	return &logger.NopLogger{}
}
