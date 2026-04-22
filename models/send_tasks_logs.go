package models

import (
	"fmt"
	"ops-message-unified-push/pkg/util"
	"strings"

	//"time"

	"gorm.io/gorm"
)

type SendTasksLogs struct {
	ID       int    `gorm:"primaryKey" json:"id" `
	TaskID   string `json:"task_id" gorm:"type:varchar(12) ;default:'';index:task_id"`
	Type     string `json:"type" gorm:"type:varchar(20) ;default:'task';comment:'类型：task-系统任务，template-接口调用，cron_message-定时消息'"`
	Name     string `json:"name" gorm:"type:varchar(256) ;default:'';comment:'任务或模板名称'"`
	Log      string `json:"log" gorm:"type:text ;"`
	Status   *int   `json:"status" gorm:"type:int ;default:0;"`
	CallerIp string `json:"caller_ip" gorm:"type:varchar(256) ;default:'';"`

	CreatedAt util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime "`
	UpdatedAt util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime ;"`
}

// Add 添加日志记录
func (log *SendTasksLogs) Add() error {
	if err := db.Create(&log).Error; err != nil {
		return err
	}
	return nil
}

// 日志列表的结果
type LogsResult struct {
	ID         int       `json:"id"`
	TaskID     string    `json:"task_id"`
	Type       string    `json:"type"`
	Name       string    `json:"name"`
	Log        string    `json:"log"`
	CreatedOn  util.Time `json:"created_on"`
	ModifiedOn util.Time `json:"modified_on"`
	Status     int       `json:"status"`
	CallerIp   string    `json:"caller_ip"`
}

// GetSendLogs 获取所有日志记录
func GetSendLogs(pageNum int, pageSize int, name string, taskId string, maps map[string]interface{}) ([]LogsResult, error) {
	var logs []LogsResult
	logt := GetSchema(SendTasksLogs{})

	// 简化查询，只查询日志表
	query := db.Table(logt)

	dayVal, ok := maps["day_created_on"]
	if ok {
		delete(maps, "day_created_on")
		query = query.Where(fmt.Sprintf("DATE(%s.created_on) = ?", logt), dayVal)
	}

	// 处理时间范围
	if startTime, ok := maps["start_time"]; ok && startTime != "" {
		delete(maps, "start_time")
		query = query.Where(fmt.Sprintf("%s.created_on >= ?", logt), startTime)
	}
	if endTime, ok := maps["end_time"]; ok && endTime != "" {
		delete(maps, "end_time")
		query = query.Where(fmt.Sprintf("%s.created_on <= ?", logt), endTime)
	}

	query = query.Where(maps)

	// 按名称搜索（搜索日志表的 name 字段）
	if name != "" {
		query = query.Where(fmt.Sprintf("%s.name like ?", logt), fmt.Sprintf("%%%s%%", name))
	}
	query = applyTaskIDFilter(query, logt, taskId)
	query = query.Order("created_on DESC")
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	query.Scan(&logs)

	normalizeCronLogs(&logs)
	fillTaskNamesForLogs(&logs)

	return logs, nil
}

// fillTaskNamesForLogs 为历史日志数据补充任务名称
func fillTaskNamesForLogs(logs *[]LogsResult) {
	if logs == nil || len(*logs) == 0 {
		return
	}

	templateIDsMap := make(map[string]bool)
	cronIdsMap := make(map[string]bool)
	for _, log := range *logs {
		if log.Name != "" || log.TaskID == "" {
			continue
		}
		if log.Type == "cron" || log.Type == "cron_message" {
			cronIdsMap[log.TaskID] = true
			continue
		}
		templateIDsMap[log.TaskID] = true
	}

	if len(templateIDsMap) == 0 && len(cronIdsMap) == 0 {
		return
	}

	templateNameMap := make(map[string]string)
	if len(templateIDsMap) > 0 {
		templateIDs := make([]string, 0, len(templateIDsMap))
		for templateID := range templateIDsMap {
			templateIDs = append(templateIDs, templateID)
		}
		var templates []TemplateResult
		templateT := GetSchema(Template{})
		db.Table(templateT).
			Select("id, name").
			Where("id IN ?", templateIDs).
			Scan(&templates)
		for _, template := range templates {
			templateNameMap[template.ID] = template.Name
		}
	}

	cronNameMap := make(map[string]string)
	if len(cronIdsMap) > 0 {
		cronIds := make([]string, 0, len(cronIdsMap))
		for cronId := range cronIdsMap {
			cronIds = append(cronIds, cronId)
		}
		var cronMsgs []CronMessages
		cront := GetSchema(CronMessages{})
		db.Table(cront).
			Select("id, name").
			Where("id IN ?", cronIds).
			Scan(&cronMsgs)
		for _, cronMsg := range cronMsgs {
			cronNameMap[cronMsg.ID] = cronMsg.Name
		}
	}

	for i := range *logs {
		log := &(*logs)[i]
		if log.Name != "" || log.TaskID == "" {
			continue
		}
		if log.Type == "cron" || log.Type == "cron_message" {
			if cronName, exists := cronNameMap[log.TaskID]; exists {
				log.Name = cronName
			}
			continue
		}
		if templateName, exists := templateNameMap[log.TaskID]; exists {
			log.Name = templateName
		}
	}
}

// GetSendLogsTotal 获取所有日志总数
func GetSendLogsTotal(name string, taskId string, maps map[string]interface{}) (int64, error) {
	var total int64
	logt := GetSchema(SendTasksLogs{})

	// 简化查询，只查询日志表
	query := db.Table(logt)

	dayVal, ok := maps["day_created_on"]
	if ok {
		delete(maps, "day_created_on")
		query = query.Where(fmt.Sprintf("DATE(%s.created_on) = ?", logt), dayVal)
	}

	// 处理时间范围
	if startTime, ok := maps["start_time"]; ok && startTime != "" {
		delete(maps, "start_time")
		query = query.Where(fmt.Sprintf("%s.created_on >= ?", logt), startTime)
	}
	if endTime, ok := maps["end_time"]; ok && endTime != "" {
		delete(maps, "end_time")
		query = query.Where(fmt.Sprintf("%s.created_on <= ?", logt), endTime)
	}

	query = query.Where(maps)

	// 按名称搜索（搜索日志表的 name 字段）
	if name != "" {
		query = query.Where(fmt.Sprintf("%s.name like ?", logt), fmt.Sprintf("%%%s%%", name))
	}
	query = applyTaskIDFilter(query, logt, taskId)
	query.Count(&total)
	return total, nil
}

func applyTaskIDFilter(query *gorm.DB, logTable string, taskID string) *gorm.DB {
	if taskID == "" {
		return query
	}
	cronMsg, err := GetCronMsgByID(taskID)
	if err == nil && cronMsg.ID != "" && cronMsg.TemplateID != "" {
		return query.Where(
			fmt.Sprintf("(%s.task_id = ? OR (%s.task_id = ? AND (%s.caller_ip LIKE ? OR %s.name = ?)))", logTable, logTable, logTable, logTable),
			taskID,
			cronMsg.TemplateID,
			"[CronTemplate]%",
			cronMsg.Name,
		)
	}
	return query.Where(fmt.Sprintf("%s.task_id = ?", logTable), taskID)
}

func normalizeCronLogs(logs *[]LogsResult) {
	if logs == nil || len(*logs) == 0 {
		return
	}
	templateIDsMap := make(map[string]bool)
	for i := range *logs {
		log := &(*logs)[i]
		if log.Type == "template" && strings.HasPrefix(log.CallerIp, "[CronTemplate]") {
			log.Type = "cron_message"
			templateIDsMap[log.TaskID] = true
			title := extractTitleFromLog(log.Log)
			if title != "" {
				log.Name = title
			}
		}
	}
	if len(templateIDsMap) == 0 {
		// 兼容历史数据：将旧的 cron 类型归一到 cron_message
		for i := range *logs {
			log := &(*logs)[i]
			if log.Type == "cron" {
				log.Type = "cron_message"
			}
		}
		return
	}
	templateIDs := make([]string, 0, len(templateIDsMap))
	for templateID := range templateIDsMap {
		templateIDs = append(templateIDs, templateID)
	}
	var cronMsgs []CronMessages
	cront := GetSchema(CronMessages{})
	db.Table(cront).
		Select("id, name, task_id").
		Where("task_id IN ?", templateIDs).
		Scan(&cronMsgs)
	templateNameMap := make(map[string]string)
	for _, cronMsg := range cronMsgs {
		if _, exists := templateNameMap[cronMsg.TemplateID]; !exists {
			templateNameMap[cronMsg.TemplateID] = cronMsg.Name
		}
	}
	for i := range *logs {
		log := &(*logs)[i]
		if log.Type != "cron_message" {
			continue
		}
		if log.Name == "" {
			if cronName, exists := templateNameMap[log.TaskID]; exists {
				log.Name = cronName
			}
		}
	}
}

func extractTitleFromLog(logText string) string {
	startFlag := "发送标题《"
	startIdx := strings.Index(logText, startFlag)
	if startIdx < 0 {
		return ""
	}
	titleStart := startIdx + len(startFlag)
	endIdx := strings.Index(logText[titleStart:], "》")
	if endIdx < 0 {
		return ""
	}
	return strings.TrimSpace(logText[titleStart : titleStart+endIdx])
}

// GetSendLogsTotal 获取所有日志总数
func DeleteOutDateLogs(keepNum int) (int, error) {
	var affectedRows int

	// 优化方案：使用GORM的Offset和Limit找到临界ID，兼容多种数据库
	// 1. 获取第 keepNum 条记录的ID作为临界值
	var threshold SendTasksLogs
	result := db.Model(&SendTasksLogs{}).
		Select("id").
		Order("created_on DESC").
		Offset(keepNum - 1).
		Limit(1).
		First(&threshold)

	// 如果记录总数不足keepNum条，则不需要删除
	if result.Error != nil {
		return 0, nil
	}

	// 2. 删除ID小于临界值的记录
	deleteResult := db.Where("id < ?", threshold.ID).Delete(&SendTasksLogs{})
	if deleteResult.Error != nil {
		return affectedRows, deleteResult.Error
	}

	affectedRows = int(deleteResult.RowsAffected)
	return affectedRows, nil
}

type StatisticData struct {
	TodaySuccNum    int `json:"today_succ_num"`
	TodayFailedNum  int `json:"today_failed_num"`
	TodayTotalNum   int `json:"today_total_num"`
	MessageTotalNum int `json:"message_total_num"`

	LatestSendData []LatestSendData `json:"latest_send_data" gorm:"many2many:latest_send_data;"`
	WayCateData    []WayCateData    `json:"way_cate_data" gorm:"many2many:way_cate_data;"`
}

// BasicStatisticData 基础统计数据
type BasicStatisticData struct {
	TodaySuccNum    int `json:"today_succ_num"`
	TodayFailedNum  int `json:"today_failed_num"`
	TodayTotalNum   int `json:"today_total_num"`
	MessageTotalNum int `json:"message_total_num"`
}

// TrendStatisticData 趋势统计数据
type TrendStatisticData struct {
	LatestSendData []LatestSendData `json:"latest_send_data"`
}

// ChannelStatisticData 渠道统计数据
type ChannelStatisticData struct {
	WayCateData []WayCateData `json:"way_cate_data"`
}

type LatestSendData struct {
	Day          string `json:"day"`
	Num          int    `json:"num"`
	SuccNum      int    `json:"succ_num"`
	DaySuccNum   int    `json:"day_succ_num"`
	DayFailedNum int    `json:"day_failed_num"`
}

type WayCateData struct {
	WayName  string `json:"way_name"`
	CountNum int    `json:"count_num"`
}

// GetStatisticData 获取统计数据
func GetStatisticData() (StatisticData, error) {
	var statistic StatisticData
	var latestData []LatestSendData
	var wayCateData []WayCateData
	logt := GetSchema(SendTasksLogs{})
	inst := GetSchema(SendTasksIns{})
	wayst := GetSchema(SendWays{})
	currDay := util.GetNowTimeStr()[:10]

	// 今日统计数据
	query := db.
		Table(logt).
		Select(`
	COUNT(*) AS today_total_num,
	SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS today_succ_num,
	SUM(CASE WHEN status != 1 or status is null THEN 1 ELSE 0 END) AS today_failed_num`).
		Where("DATE(created_on) = ?", currDay)

	query.Take(&statistic)

	// 	全部消息统计数据
	totalQuery := db.Table(logt).Select(`COUNT(*) AS message_total_num`)
	totalQuery.Take(&statistic)

	// 最近30天数据
	days := 30
	now := util.GetNowTime()
	past := now.AddDate(0, 0, -days)
	pastDate := past.Format("2006-01-02")
	next := now.AddDate(0, 0, 1)
	nextDate := next.Format("2006-01-02")
	queryData := db.
		Table(logt).
		Select(`
	CAST(DATE(created_on) AS CHAR) AS day,
	SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS day_succ_num,
	SUM(CASE WHEN status != 1 or status is null THEN 1 ELSE 0 END) AS day_failed_num,
	COUNT(*) AS num`).
		Where(fmt.Sprintf(" created_on >= '%s' and created_on <= '%s' ", pastDate, nextDate)).
		Group("day").
		Order("day")

	queryData.Scan(&latestData)

	// 消息实例分类数目
	db.
		Table(inst).
		Select(fmt.Sprintf("%s.name as way_name, count(%s.id) as count_num", wayst, wayst)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", wayst, inst, wayst)).
		Group(fmt.Sprintf("%s.id", wayst)).
		Scan(&wayCateData)

	statistic.LatestSendData = latestData
	statistic.WayCateData = wayCateData
	return statistic, nil
}

// GetBasicStatisticData 获取基础统计数据
func GetBasicStatisticData() (BasicStatisticData, error) {
	var statistic BasicStatisticData
	// 改为使用 send_stats 统计表，避免日志清理导致统计数据减少
	statsTable := GetSchema(SendStats{})
	currDay := util.GetNowTimeStr()[:10]

	// 今日统计数据
	db.Table(statsTable).
		Select(`
	COALESCE(SUM(num), 0) AS today_total_num,
	COALESCE(SUM(CASE WHEN status = 'success' THEN num ELSE 0 END), 0) AS today_succ_num,
	COALESCE(SUM(CASE WHEN status = 'failed' THEN num ELSE 0 END), 0) AS today_failed_num`).
		Where("day = ?", currDay).
		Scan(&statistic)

	// 全部消息统计数据
	db.Table(statsTable).
		Select(`COALESCE(SUM(num), 0) AS message_total_num`).
		Scan(&statistic)

	return statistic, nil
}

// GetTrendStatisticData 获取趋势统计数据（使用 send_stats 表）
func GetTrendStatisticData(days int) (TrendStatisticData, error) {
	var statistic TrendStatisticData
	var latestData []LatestSendData
	statsTable := GetSchema(SendStats{})

	// 默认30天，如果传入参数则使用参数值
	if days <= 0 {
		days = 30
	}

	now := util.GetNowTime()
	past := now.AddDate(0, 0, -days)
	pastDate := past.Format("2006-01-02")

	queryData := db.
		Table(statsTable).
		Select(`
			day,
			SUM(CASE WHEN status = 'success' THEN num ELSE 0 END) AS day_succ_num,
			SUM(CASE WHEN status = 'failed' THEN num ELSE 0 END) AS day_failed_num,
			SUM(num) AS num
		`).
		Where("day >= ?", pastDate).
		Group("day").
		Order("day")

	queryData.Scan(&latestData)
	statistic.LatestSendData = latestData
	return statistic, nil
}

// GetChannelStatisticData 获取渠道统计数据（基于 send_stats 表统计任务，再查询渠道）
func GetChannelStatisticData() (ChannelStatisticData, error) {
	var statistic ChannelStatisticData
	var wayCateData []WayCateData
	insTable := GetSchema(SendTasksIns{})
	waysTable := GetSchema(SendWays{})

	// 统计所有任务和模板中关联的实例的渠道数据
	var wayStats []struct {
		WayName  string
		CountNum int
	}

	err := db.
		Table(insTable).
		Select(fmt.Sprintf("%s.name as way_name, COUNT(*) as count_num", waysTable)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", waysTable, insTable, waysTable)).
		Group(fmt.Sprintf("%s.name", waysTable)).
		Scan(&wayStats).Error

	if err != nil {
		return statistic, err
	}

	// 转换为返回格式
	for _, stat := range wayStats {
		wayCateData = append(wayCateData, WayCateData{
			WayName:  stat.WayName,
			CountNum: stat.CountNum,
		})
	}

	// 确保返回空数组而不是 nil
	if len(wayCateData) == 0 {
		wayCateData = []WayCateData{}
	}

	statistic.WayCateData = wayCateData
	return statistic, nil
}
