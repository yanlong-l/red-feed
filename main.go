package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"red-feed/ioc"
	"syscall"
	"time"
)

func main() {
	// 初始化配置
	initViper()
	// 初始化日志
	initLogger()

	// 初始化Opentelemetry
	shutdownOTEL := ioc.InitOTEL()
	defer shutdownOTEL(context.Background())

	// 初始化Prometheus
	initPrometheus()

	// 初始化APP
	app := InitApp()

	// 开启所有consumers
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	// 开启定时任务
	app.cron.Start()

	server := app.web
	server.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "ok",
		})
	})

	// 启动HTTP服务器
	srv := &http.Server{
		Addr:    ":8100",
		Handler: server,
	}

	// 在goroutine中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Shutting down server...")

	// 优雅关闭服务器，最多等待5秒
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server forced to shutdown", zap.Error(err))
	}
	zap.L().Info("Server exiting")

	// 关闭定时任务
	zap.L().Info("Cron shutting down...")
	ctx = app.cron.Stop()
	// 这边可以考虑超时强制退出，防止有些任务，执行特别长的时间
	tm := time.NewTimer(time.Minute * 10)
	select {
	case <-tm.C:
	case <-ctx.Done():
	}
	zap.L().Info("Cron exiting")
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8003", nil)
		if err != nil {
			panic(err)
		}
	}()
}

func initViper() {
	viper.SetConfigFile("config/dev.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	zap.L().Info("init logger")
}
