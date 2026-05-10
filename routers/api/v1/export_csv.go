package v1

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/service/consume_log_service"
	"ops-message-unified-push/service/cron_msg_service"
	"ops-message-unified-push/service/message_template_service"
	"ops-message-unified-push/service/mq_source_service"
	"ops-message-unified-push/service/send_logs_service"
	"ops-message-unified-push/service/send_way_service"
	"ops-message-unified-push/service/statistic_service"
	"ops-message-unified-push/service/subscription_service"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultExportPage = 1
	defaultExportSize = 1000
	maxExportSize     = 5000
)

func parseExportPageAndSize(c *gin.Context, pageKey, sizeKey string) (int, int) {
	page := defaultExportPage
	size := defaultExportSize

	if raw := strings.TrimSpace(c.DefaultQuery(pageKey, strconv.Itoa(defaultExportPage))); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			page = v
		}
	}
	if raw := strings.TrimSpace(c.DefaultQuery(sizeKey, strconv.Itoa(defaultExportSize))); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			size = v
		}
	}
	if size > maxExportSize {
		size = maxExportSize
	}
	return page, size
}

func pageToOffset(page, size int) int {
	if page <= 1 || size <= 0 {
		return 0
	}
	return (page - 1) * size
}

func buildExportFilename(prefix string) string {
	return fmt.Sprintf("%s-report-%s.csv", prefix, time.Now().Format("2006-01-02"))
}

func writeCsvAttachment(c *gin.Context, filename string, headers []string, rows [][]string) error {
	escapedName := strings.ReplaceAll(url.QueryEscape(filename), "+", "%20")
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", filename, escapedName))
	c.Status(http.StatusOK)

	if _, err := c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}
	writer := csv.NewWriter(c.Writer)
	if err := writer.Write(headers); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func ExportDashboardStatisticCsv(c *gin.Context) {
	appG := app.Gin{C: c}
	days := 30
	if raw := strings.TrimSpace(c.DefaultQuery("days", "30")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			days = v
		}
	}
	if days > 90 {
		days = 90
	}

	service := statistic_service.StatisticService{Days: days}
	basic, err := service.GetBasicStatisticData()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出统计失败", nil)
		return
	}
	trend, err := service.GetTrendStatisticData()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出统计失败", nil)
		return
	}
	channels, err := service.GetChannelStatisticData()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出统计失败", nil)
		return
	}

	rows := make([][]string, 0, len(trend.LatestSendData)+len(channels.WayCateData)+10)
	rows = append(rows,
		[]string{"模块", "指标", "值"},
		[]string{"概览", "发送日志数", strconv.Itoa(basic.MessageTotalNum)},
		[]string{"概览", "今日发送数", strconv.Itoa(basic.TodayTotalNum)},
		[]string{"概览", "今日成功数", strconv.Itoa(basic.TodaySuccNum)},
		[]string{"概览", "今日失败数", strconv.Itoa(basic.TodayFailedNum)},
		[]string{},
		[]string{"趋势", "日期", "发送总数", "发送成功数", "发送失败数"},
	)
	for _, item := range trend.LatestSendData {
		rows = append(rows, []string{
			"趋势",
			item.Day,
			strconv.Itoa(item.Num),
			strconv.Itoa(item.DaySuccNum),
			strconv.Itoa(item.DayFailedNum),
		})
	}
	rows = append(rows, []string{}, []string{"渠道", "渠道名称", "发送数"})
	for _, item := range channels.WayCateData {
		rows = append(rows, []string{
			"渠道",
			item.WayName,
			strconv.Itoa(item.CountNum),
		})
	}

	if err := writeCsvAttachment(c, buildExportFilename("dashboard"),
		[]string{"模块", "指标", "值"},
		rows[1:],
	); err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出统计失败", nil)
	}
}

func ExportTaskSendLogsCsv(c *gin.Context) {
	appG := app.Gin{C: c}
	page, size := parseExportPageAndSize(c, "page", "size")
	startTime, endTime := pickDateRangeQuery(c)

	service := send_logs_service.SendTaskLogsService{
		TaskId:    c.Query("taskid"),
		Name:      c.Query("name"),
		Query:     c.Query("query"),
		StartTime: startTime,
		EndTime:   endTime,
		PageNum:   pageToOffset(page, size),
		PageSize:  size,
	}
	logs, err := service.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出日志失败", nil)
		return
	}

	rows := make([][]string, 0, len(logs))
	for _, item := range logs {
		status := "失败"
		if item.Status == 1 {
			status = "成功"
		}
		rows = append(rows, []string{
			strconv.Itoa(item.ID),
			item.Type,
			item.Name,
			item.Log,
			item.CreatedOn.String(),
			status,
		})
	}
	if err := writeCsvAttachment(c, buildExportFilename("sendlogs"),
		[]string{"ID", "类型", "名称", "日志", "发送时间", "状态"},
		rows,
	); err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出日志失败", nil)
	}
}

func ExportCronMsgCsv(c *gin.Context) {
	appG := app.Gin{C: c}
	page, size := parseExportPageAndSize(c, "page", "size")
	startTime, endTime := pickDateRangeQuery(c)

	service := cron_msg_service.CronMsgService{
		Name:      c.Query("name"),
		PageNum:   pageToOffset(page, size),
		PageSize:  size,
		Status:    c.Query("status"),
		StartTime: startTime,
		EndTime:   endTime,
	}
	tasks, err := service.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出定时任务失败", nil)
		return
	}

	rows := make([][]string, 0, len(tasks))
	for _, item := range tasks {
		channelNames := ""
		if len(item.ChannelNames) > 0 {
			channelNames = strings.Join(item.ChannelNames, "、")
		}
		enableStatus := "停用"
		if item.Enable == 1 {
			enableStatus = "启用"
		}
		rows = append(rows, []string{
			item.ID,
			item.Name,
			item.TemplateName,
			channelNames,
			item.Cron,
			item.NextTime,
			item.CreatedAt.String(),
			enableStatus,
		})
	}
	if err := writeCsvAttachment(c, buildExportFilename("cronmessages"),
		[]string{"ID", "名称", "模板", "渠道", "Cron表达式", "下次执行时间", "创建时间", "启用状态"},
		rows,
	); err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出定时任务失败", nil)
	}
}

func ExportSubscriptionCsv(c *gin.Context) {
	appG := app.Gin{C: c}
	page, size := parseExportPageAndSize(c, "page", "page_size")
	startTime, endTime := pickDateRangeQuery(c)

	service := subscription_service.SubscriptionService{}
	list, _, err := service.GetAll(c.Query("name"), c.Query("status"), c.Query("source_id"), startTime, endTime, page, size)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出订阅失败", nil)
		return
	}

	rows := make([][]string, 0, len(list))
	for _, item := range list {
		status := "已停止"
		if item.Status == "running" {
			status = "运行中"
		}
		tag := item.Tag
		if strings.TrimSpace(tag) == "" {
			tag = "*"
		}
		rows = append(rows, []string{
			item.ID,
			item.Name,
			item.SourceName,
			item.Topic,
			tag,
			status,
			strconv.Itoa(item.TotalConsumed),
			strconv.Itoa(item.TotalSent),
			strconv.Itoa(item.TotalFailed),
			item.LastConsumeTime,
		})
	}
	if err := writeCsvAttachment(c, buildExportFilename("subscriptions"),
		[]string{"ID", "订阅名称", "数据源", "Topic", "Tag", "状态", "消费总数", "发送总数", "失败总数", "最后消费时间"},
		rows,
	); err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出订阅失败", nil)
	}
}

func ExportTemplateCsv(c *gin.Context) {
	appG := app.Gin{C: c}
	page, size := parseExportPageAndSize(c, "page", "size")
	startTime, endTime := pickDateRangeQuery(c)
	text := c.Query("text")
	if text == "" {
		text = c.Query("name")
	}

	service := message_template_service.TemplateService{
		Text:      text,
		Status:    c.Query("status"),
		StartTime: startTime,
		EndTime:   endTime,
		PageNum:   pageToOffset(page, size),
		PageSize:  size,
	}
	templates, err := service.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出模板失败", nil)
		return
	}

	rows := make([][]string, 0, len(templates))
	for _, item := range templates {
		status := "禁用"
		if item.Status == "enabled" {
			status = "启用"
		}
		rows = append(rows, []string{
			item.ID,
			item.Name,
			item.Description,
			status,
			item.CreatedOn.String(),
			item.ModifiedOn.String(),
		})
	}
	if err := writeCsvAttachment(c, buildExportFilename("templates"),
		[]string{"ID", "模板名称", "描述", "状态", "创建时间", "更新时间"},
		rows,
	); err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出模板失败", nil)
	}
}

func ExportSendWayCsv(c *gin.Context) {
	appG := app.Gin{C: c}
	page, size := parseExportPageAndSize(c, "page", "size")
	startTime, endTime := pickDateRangeQuery(c)

	channelType := c.Query("type")
	if channelType == "" {
		channelType = c.Query("channel_type")
	}

	service := send_way_service.SendWay{
		Name:      c.Query("name"),
		Type:      channelType,
		StartTime: startTime,
		EndTime:   endTime,
		PageNum:   pageToOffset(page, size),
		PageSize:  size,
	}
	ways, err := service.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出渠道失败", nil)
		return
	}

	rows := make([][]string, 0, len(ways))
	for _, item := range ways {
		rows = append(rows, []string{
			item.ID,
			item.Name,
			item.Type,
			"-",
			item.CreatedAt.String(),
			item.UpdatedAt.String(),
		})
	}
	if err := writeCsvAttachment(c, buildExportFilename("sendways"),
		[]string{"ID", "渠道名称", "渠道类型", "状态", "创建时间", "更新时间"},
		rows,
	); err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出渠道失败", nil)
	}
}

func ExportLoginLogCsv(c *gin.Context) {
	appG := app.Gin{C: c}
	page, size := parseExportPageAndSize(c, "page", "page_size")
	startTime, endTime := pickDateRangeQuery(c)
	logs, _, err := models.GetLoginLogs(page, size, startTime, endTime)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出登录日志失败", nil)
		return
	}

	rows := make([][]string, 0, len(logs))
	for _, item := range logs {
		rows = append(rows, []string{
			strconv.Itoa(int(item.ID)),
			strconv.Itoa(item.UserID),
			item.Username,
			item.IP,
			item.UA,
			item.CreatedAt.String(),
		})
	}
	if err := writeCsvAttachment(c, buildExportFilename("loginlogs"),
		[]string{"ID", "用户ID", "用户名", "IP", "UA", "登录时间"},
		rows,
	); err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出登录日志失败", nil)
	}
}

func ExportConsumeLogCsv(c *gin.Context) {
	appG := app.Gin{C: c}
	page, size := parseExportPageAndSize(c, "page", "page_size")
	startTime, endTime := pickDateRangeQuery(c)
	service := consume_log_service.ConsumeLogService{}
	subscriptionID := c.Query("subscription_id")
	subscriptionName := c.Query("subscription_name")
	if subscriptionID == "" && subscriptionName != "" {
		subscriptionID = service.ResolveSubscriptionIDByName(subscriptionName)
	}
	matched := service.ParseStatusValue(c.Query("matched"))
	sendStatus := c.Query("send_status")
	if sendStatus == "" {
		sendStatus = c.Query("status")
	}
	sendStatus = service.ParseStatusValue(sendStatus)

	logs, _, err := service.GetConsumeLogList(
		subscriptionID,
		sendStatus,
		matched,
		startTime,
		endTime,
		page,
		size,
	)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出消费日志失败", nil)
		return
	}

	rows := make([][]string, 0, len(logs))
	for _, raw := range logs {
		item, ok := raw.(consume_log_service.ConsumeLogDTO)
		if !ok {
			continue
		}
		matchedText := "未匹配"
		if item.Matched == 1 {
			matchedText = "已匹配"
		}
		sendText := "未发送"
		switch item.SendStatus {
		case 1:
			sendText = "发送成功"
		case 2:
			sendText = "发送失败"
		}
		rows = append(rows, []string{
			strconv.Itoa(int(item.ID)),
			item.SubscriptionID,
			item.SubscriptionName,
			item.RawMessage,
			matchedText,
			sendText,
			item.SendError,
			item.CreatedOn,
		})
	}
	if err := writeCsvAttachment(c, buildExportFilename("consume-logs"),
		[]string{"ID", "订阅ID", "订阅名称", "原始消息", "匹配状态", "发送状态", "发送错误", "消费时间"},
		rows,
	); err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出消费日志失败", nil)
	}
}

func ExportMQSourceCsv(c *gin.Context) {
	appG := app.Gin{C: c}
	page, size := parseExportPageAndSize(c, "page", "page_size")
	startTime, endTime := pickDateRangeQuery(c)
	status := strings.TrimSpace(c.Query("status"))
	if status == "__untested__" {
		status = "untested"
	}

	service := mq_source_service.MQSourceService{}
	list, _, err := service.GetAll(c.Query("name"), status, c.Query("type"), startTime, endTime, page, size)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出数据源失败", nil)
		return
	}

	rows := make([][]string, 0, len(list))
	for _, item := range list {
		statusText := "未测试"
		switch item.LastTestStatus {
		case "success":
			statusText = "在线"
		case "failed":
			statusText = "离线"
		}
		rows = append(rows, []string{
			item.ID,
			item.Name,
			item.Type,
			item.NamesrvAddr,
			statusText,
			strconv.Itoa(item.BindingCount),
			item.LastTestTime.String(),
			item.CreatedAt.String(),
			item.UpdatedAt.String(),
		})
	}
	if err := writeCsvAttachment(c, buildExportFilename("mq-sources"),
		[]string{"ID", "数据源名称", "类型", "地址", "状态", "绑定订阅数", "最后测试时间", "创建时间", "更新时间"},
		rows,
	); err != nil {
		appG.CResponse(http.StatusInternalServerError, "导出数据源失败", nil)
	}
}
