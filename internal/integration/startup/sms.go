package startup

import (
	"red-feed/internal/service/sms"
	"red-feed/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 暂时先用内存来实现
	return memory.NewService()
}
