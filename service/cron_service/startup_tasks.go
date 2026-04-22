package cron_service

import (
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/constant"
	"ops-message-unified-push/service/mq_source_service"
	"ops-message-unified-push/service/send_message_service"
	"strconv"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/unknwon/com"
)

var ClearLogsTaskId cron.EntryID
var ClearConsumeLogsTaskId cron.EntryID
var ClearLoginLogsTaskId cron.EntryID
var MQStatusProbeTaskId cron.EntryID

// CleanConfig 清理任务配置
type CleanConfig struct {
	TaskID       string
	TaskName     string
	LogPrefix    string
	SectionName  string
	EnabledKey   string
	KeepNumKey   string
	DeleteFunc   func(int) (int, error)
	ResourceName string // 资源名称，如"日志"、"托管消息"
}

// executeCleanTask 执行清理任务的通用逻辑
func executeCleanTask(config CleanConfig) {
	var errStr string
	sm := send_message_service.SendMessageService{
		TaskID: config.TaskID,
		Name:   config.TaskName,
		DefaultLogger: logrus.WithFields(logrus.Fields{
			"prefix": config.LogPrefix,
		}),
	}
	sm.Status = send_message_service.SendSuccess

	// 检查是否启用
	enabledSetting, err := models.GetSettingByKey(config.SectionName, config.EnabledKey)
	if err != nil {
		errStr = fmt.Sprintf("获取%s清理开关失败，原因：%s", config.ResourceName, err)
		sm.LogsAndStatusMark(errStr, send_message_service.SendFail)
		sm.RecordSendLog()
		return
	}

	if enabledSetting.Value != "true" {
		sm.LogsAndStatusMark(fmt.Sprintf("%s清理功能未启用，跳过执行", config.ResourceName), sm.Status)
		sm.RecordSendLog()
		return
	}

	sm.LogsAndStatusMark(fmt.Sprintf("开始清除%s", config.ResourceName), sm.Status)

	setting, err := models.GetSettingByKey(config.SectionName, config.KeepNumKey)
	if err != nil {
		errStr = fmt.Sprintf("获取%s的保留数失败，原因：%s", config.ResourceName, err)
		sm.LogsAndStatusMark(errStr, send_message_service.SendFail)
	}

	keepNum := com.StrTo(setting.Value).MustInt()
	affectedRows, err := config.DeleteFunc(keepNum)
	if err != nil {
		errStr = fmt.Sprintf("删除%s失败，原因：%s", config.ResourceName, err)
		sm.LogsAndStatusMark(errStr, send_message_service.SendFail)
	} else {
		errStr = fmt.Sprintf("删除%s成功，删除条数：%d，保留数目：%d", config.ResourceName, affectedRows, keepNum)
		sm.LogsAndStatusMark(errStr, sm.Status)
	}

	sm.RecordSendLog()
}

// ClearLogs 清除任务日志的定时任务
func ClearLogs() {
	executeCleanTask(CleanConfig{
		TaskID:       constant.CleanLogsTaskId,
		TaskName:     "任务日志定时清除",
		LogPrefix:    "[Cron Clear Task Logs]",
		SectionName:  constant.LogsCleanSectionName,
		EnabledKey:   constant.LogsCleanEnabledKeyName,
		KeepNumKey:   constant.LogsCleanKeepKeyName,
		DeleteFunc:   models.DeleteOutDateLogs,
		ResourceName: "任务日志",
	})
}

// ClearConsumeLogs 清除消费日志的定时任务
func ClearConsumeLogs() {
	executeCleanTask(CleanConfig{
		TaskID:       constant.CleanConsumeLogsTaskId,
		TaskName:     "消费日志定时清除",
		LogPrefix:    "[Cron Clear Consume Logs]",
		SectionName:  constant.ConsumeLogsCleanSectionName,
		EnabledKey:   constant.LogsCleanEnabledKeyName,
		KeepNumKey:   constant.LogsCleanKeepKeyName,
		DeleteFunc:   models.DeleteOutDateConsumeLogs,
		ResourceName: "消费日志",
	})
}

// ClearLoginLogs 清除登录日志的定时任务
func ClearLoginLogs() {
	executeCleanTask(CleanConfig{
		TaskID:       constant.CleanLoginLogsTaskId,
		TaskName:     "登录日志定时清除",
		LogPrefix:    "[Cron Clear Login Logs]",
		SectionName:  constant.LoginLogsCleanSectionName,
		EnabledKey:   constant.LogsCleanEnabledKeyName,
		KeepNumKey:   constant.LogsCleanKeepKeyName,
		DeleteFunc:   models.DeleteOutDateLoginLogs,
		ResourceName: "登录日志",
	})
}

type CronService struct {
}

// startCleanCronTask 启动清理任务的通用逻辑
func startCleanCronTask(sectionName, enabledKey, cronKey, resourceName string, job func(), taskId *cron.EntryID) {
	// 检查是否启用
	enabledSetting, err := models.GetSettingByKey(sectionName, enabledKey)
	if err != nil {
		logrus.Error(fmt.Sprintf("获取[%s]清理开关失败，原因：%s", resourceName, err))
		return
	}

	if enabledSetting.Value != "true" {
		logrus.Info(fmt.Sprintf("[%s]清理功能未启用", resourceName))
		return
	}

	// 注册任务
	setting, err := models.GetSettingByKey(sectionName, cronKey)
	if err != nil {
		logrus.Error(fmt.Sprintf("获取[%s]的cron失败，原因：%s", resourceName, err))
		return
	}
	*taskId = AddTask(ScheduledTask{
		Schedule: setting.Value,
		Job:      job,
	})
	logrus.Info(fmt.Sprintf("[%s]清理任务已启动", resourceName))
}

// updateCleanCronTask 更新清理任务的通用逻辑
func updateCleanCronTask(cron string, enabled bool, resourceName string, job func(), taskId *cron.EntryID) {
	// 先移除旧任务
	if *taskId > 0 {
		RemoveTask(*taskId)
		*taskId = 0
	}

	// 如果启用，则添加新任务
	if enabled {
		*taskId = AddTask(ScheduledTask{
			Schedule: cron,
			Job:      job,
		})
		logrus.Info(fmt.Sprintf("更新%s的cron成功，%s", resourceName, cron))
	} else {
		logrus.Info(fmt.Sprintf("%s清理任务已停止", resourceName))
	}
	logrus.Info(fmt.Sprintf("所有的定时任务总数： %d", len(TaskList)))
}

// StartLogsCronRun 启动注册任务日志清理定时任务
func (cs *CronService) StartLogsCronRun() {
	startCleanCronTask(
		constant.LogsCleanSectionName,
		constant.LogsCleanEnabledKeyName,
		constant.LogsCleanCronKeyName,
		"任务日志",
		ClearLogs,
		&ClearLogsTaskId,
	)
}

// UpdateLogsCronRun 更新任务日志清理定时任务
func (cs *CronService) UpdateLogsCronRun(cron string, enabled bool) {
	updateCleanCronTask(cron, enabled, "任务日志", ClearLogs, &ClearLogsTaskId)
}

// StartConsumeLogsCronRun 启动注册消费日志清理定时任务
func (cs *CronService) StartConsumeLogsCronRun() {
	startCleanCronTask(
		constant.ConsumeLogsCleanSectionName,
		constant.LogsCleanEnabledKeyName,
		constant.LogsCleanCronKeyName,
		"消费日志",
		ClearConsumeLogs,
		&ClearConsumeLogsTaskId,
	)
}

// UpdateConsumeLogsCronRun 更新消费日志清理定时任务
func (cs *CronService) UpdateConsumeLogsCronRun(cron string, enabled bool) {
	updateCleanCronTask(cron, enabled, "消费日志", ClearConsumeLogs, &ClearConsumeLogsTaskId)
}

// StartLoginLogsCronRun 启动注册登录日志清理定时任务
func (cs *CronService) StartLoginLogsCronRun() {
	startCleanCronTask(
		constant.LoginLogsCleanSectionName,
		constant.LogsCleanEnabledKeyName,
		constant.LogsCleanCronKeyName,
		"登录日志",
		ClearLoginLogs,
		&ClearLoginLogsTaskId,
	)
}

// UpdateLoginLogsCronRun 更新登录日志清理定时任务
func (cs *CronService) UpdateLoginLogsCronRun(cron string, enabled bool) {
	updateCleanCronTask(cron, enabled, "登录日志", ClearLoginLogs, &ClearLoginLogsTaskId)
}

// StartLogsCronRunOnStartup 启动的时候开启任务日志清理任务
func StartLogsCronRunOnStartup() {
	logrus.Infof("开始注册定时清除任务日志任务...")
	cs := CronService{}
	cs.StartLogsCronRun()
}

// StartConsumeLogsCronRunOnStartup 启动的时候开启消费日志清理任务
func StartConsumeLogsCronRunOnStartup() {
	logrus.Infof("开始注册定时清除消费日志任务...")
	cs := CronService{}
	cs.StartConsumeLogsCronRun()
}

// StartLoginLogsCronRunOnStartup 启动的时候开启登录日志清理任务
func StartLoginLogsCronRunOnStartup() {
	logrus.Infof("开始注册定时清除登录日志任务...")
	cs := CronService{}
	cs.StartLoginLogsCronRun()
}

func ProbeMQSourceStatus() {
	sources, err := models.GetMQSources(0, 0, "", "", "")
	if err != nil {
		logrus.Errorf("获取消息队列数据源失败: %v", err)
		return
	}

	service := mq_source_service.MQSourceService{}
	total := 0
	success := 0
	failed := 0
	for _, source := range sources {
		if source.Enabled != 1 {
			continue
		}
		total++
		if err := service.TestConnectionByID(source.ID); err != nil {
			failed++
			continue
		}
		success++
	}
	logrus.Infof("消息队列状态自动更新完成: total=%d success=%d failed=%d", total, success, failed)
}

func buildMQStatusProbeSchedule(intervalSeconds int) string {
	if intervalSeconds < 10 {
		intervalSeconds = 10
	}
	return fmt.Sprintf("@every %ds", intervalSeconds)
}

func (cs *CronService) StartMQStatusPolicyRun() {
	enabledSetting, err := models.GetSettingByKey(constant.MQStatusPolicySectionName, constant.MQStatusPolicyEnabledKeyName)
	if err != nil {
		logrus.Errorf("获取消息队列状态更新策略开关失败: %v", err)
		return
	}
	if enabledSetting.Value != "true" {
		logrus.Info("消息队列状态自动更新未启用")
		return
	}

	intervalSetting, err := models.GetSettingByKey(constant.MQStatusPolicySectionName, constant.MQStatusPolicyIntervalSecondsKeyName)
	if err != nil {
		logrus.Errorf("获取消息队列状态更新频率失败: %v", err)
		return
	}
	seconds, _ := strconv.Atoi(intervalSetting.Value)
	if seconds <= 0 {
		seconds = 300
	}
	MQStatusProbeTaskId = AddTask(ScheduledTask{
		Schedule: buildMQStatusProbeSchedule(seconds),
		Job:      ProbeMQSourceStatus,
	})
	logrus.Infof("消息队列状态自动更新任务已启动，频率=%ds", seconds)
}

func (cs *CronService) UpdateMQStatusPolicyRun(enabled bool, intervalSeconds int) {
	if MQStatusProbeTaskId > 0 {
		RemoveTask(MQStatusProbeTaskId)
		MQStatusProbeTaskId = 0
	}
	if enabled {
		MQStatusProbeTaskId = AddTask(ScheduledTask{
			Schedule: buildMQStatusProbeSchedule(intervalSeconds),
			Job:      ProbeMQSourceStatus,
		})
		logrus.Infof("消息队列状态自动更新任务已更新，频率=%ds", intervalSeconds)
	} else {
		logrus.Info("消息队列状态自动更新任务已停止")
	}
}

func StartMQStatusPolicyRunOnStartup() {
	logrus.Infof("开始注册消息队列状态自动更新任务...")
	cs := CronService{}
	cs.StartMQStatusPolicyRun()
}

func StartTasksRunOnStartup() {
	StartLogsCronRunOnStartup()
	StartConsumeLogsCronRunOnStartup()
	StartLoginLogsCronRunOnStartup()
	StartMQStatusPolicyRunOnStartup()
}
