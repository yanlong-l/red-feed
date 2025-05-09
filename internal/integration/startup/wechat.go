package startup

import (
	"os"
	"red-feed/internal/service/oauth2/wechat"
	"red-feed/pkg/logger"
)

func InitWechatService(l logger.Logger) wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		appId = ""
		// panic("没有找到环境变量 WECHAT_APP_ID ")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		appKey = ""
		// panic("没有找到环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewService(appId, appKey, l)
}
