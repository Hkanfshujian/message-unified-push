package main

import (
	"embed"
	"fmt"
	"ops-message-unified-push/migrate"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/constant"
	"ops-message-unified-push/pkg/logging"
	"ops-message-unified-push/pkg/setting"
	"ops-message-unified-push/routers"
	"ops-message-unified-push/service/cron_msg_service"
	"ops-message-unified-push/service/cron_service"
	"ops-message-unified-push/service/mq_consumer"
	"ops-message-unified-push/service/subscription_service"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	//go:embed web/dist/*
	f embed.FS

	//go:embed .release*
	rf embed.FS
)

func init() {
	constant.InitReleaseInfo(rf)
	setting.Setup()
	logging.Setup()
	models.Setup()
	go func() {
		// 完成model的迁移之后，需要加载异步任务
		migrate.Setup()
		cron_service.StartTasksRunOnStartup()
		// 加载用户自己设定的定时消息任务
		cron_msg_service.StartUpUserSetupMsgCronTask()
		
		// 初始化 MQ Consumer Manager
		subscription_service.GlobalConsumerManager = mq_consumer.NewConsumerManager()
		// 启动所有运行中的订阅
		subscription_service.GlobalConsumerManager.StartAllRunning()
	}()
}

func main() {
	gin.SetMode(setting.ServerSetting.RunMode)
	routersInit := routers.InitRouter(f)
	readTimeout := setting.ServerSetting.ReadTimeout
	writeTimeout := setting.ServerSetting.WriteTimeout
	endPoint := fmt.Sprintf(":%d", setting.ServerSetting.HttpPort)
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           endPoint,
		Handler:        routersInit,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	startInfo := ""
	if setting.ServerSetting.RunMode == "debug" {
		startInfo = fmt.Sprintf("run mode: %s, start message server @ http://localhost%s", setting.ServerSetting.RunMode, endPoint)
	} else {
		startInfo = fmt.Sprintf("run mode: %s, start message server @ http://0.0.0.0%s", setting.ServerSetting.RunMode, endPoint)
	}

	logrus.WithFields(logrus.Fields{
		"prefix": fmt.Sprintf("[PID:%d]", os.Getpid()),
	}).Info(startInfo)

	// 优雅关闭
	go func() {
		// 监听系统信号
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logrus.Info("正在关闭服务...")
		
		// 停止所有 MQ 订阅
		if subscription_service.GlobalConsumerManager != nil {
			logrus.Info("正在停止所有 MQ 订阅...")
			subscription_service.GlobalConsumerManager.StopAll()
		}
		
		os.Exit(0)
	}()

	err := server.ListenAndServe()
	if err != nil {
		logrus.Errorf("Server err: %v", err)
	}
}
