package main

import (
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

func main() {
	initLogger()
	initViper()
	// server := InitWebServer()
	// err := server.Run(":8080")
	// if err != nil {
	// 	return
	// }
}

func initViper() {
	viper.SetConfigFile("config/dev.yaml")
	viper.ReadInConfig()
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	zap.L().Info("init logger")
}
