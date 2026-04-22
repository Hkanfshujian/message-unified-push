package consume_log_service

import (
	"encoding/json"
	"ops-message-unified-push/models"
	"strconv"
	"strings"
)

type ConsumeLogService struct {
}

type ConsumeLogDTO struct {
	ID               uint   `json:"id"`
	SubscriptionID   string `json:"subscription_id"`
	SubscriptionName string `json:"subscription_name"`
	RawMessage       string `json:"raw_message"`
	Matched          int    `json:"matched"`
	ExtractedValues  string `json:"extracted_values"`
	SendStatus       int    `json:"send_status"`
	SendError        string `json:"send_error"`
	CreatedOn        string `json:"created_on"`
}

// GetConsumeLogList 获取消费日志列表
func (s *ConsumeLogService) GetConsumeLogList(subscriptionID, status, matched, startTime, endTime string, page, pageSize int) ([]interface{}, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// 兼容前端 send_status 参数传入 status 字段
	sendStatus := status
	logs, err := models.GetConsumeLogs(page, pageSize, subscriptionID, matched, sendStatus, startTime, endTime)
	if err != nil {
		return nil, 0, err
	}
	total, err := models.GetConsumeLogsTotal(subscriptionID, matched, sendStatus, startTime, endTime)
	if err != nil {
		return nil, 0, err
	}

	// 批量补充订阅名称
	subNames := map[string]string{}
	for _, log := range logs {
		if log.SubscriptionID == "" {
			continue
		}
		if _, ok := subNames[log.SubscriptionID]; ok {
			continue
		}
		sub, subErr := models.GetSubscriptionByID(log.SubscriptionID)
		if subErr == nil {
			subNames[log.SubscriptionID] = sub.Name
		}
	}

	result := make([]interface{}, 0, len(logs))
	for _, log := range logs {
		extracted := encodeExtractedValues(log.ExtractedValues)
		result = append(result, ConsumeLogDTO{
			ID:               log.ID,
			SubscriptionID:   log.SubscriptionID,
			SubscriptionName: subNames[log.SubscriptionID],
			RawMessage:       log.RawMessage,
			Matched:          log.Matched,
			ExtractedValues:  extracted,
			SendStatus:       log.SendStatus,
			SendError:        log.SendError,
			CreatedOn:        log.ConsumeTime.String(),
		})
	}
	return result, total, nil
}

// GetConsumeLogByID 根据 ID 获取消费日志
func (s *ConsumeLogService) GetConsumeLogByID(id uint) (interface{}, error) {
	log, err := models.GetConsumeLogByID(id)
	if err != nil {
		return nil, err
	}

	subName := ""
	if log.SubscriptionID != "" {
		if sub, subErr := models.GetSubscriptionByID(log.SubscriptionID); subErr == nil {
			subName = sub.Name
		}
	}

	extracted := encodeExtractedValues(log.ExtractedValues)

	return ConsumeLogDTO{
		ID:               log.ID,
		SubscriptionID:   log.SubscriptionID,
		SubscriptionName: subName,
		RawMessage:       log.RawMessage,
		Matched:          log.Matched,
		ExtractedValues:  extracted,
		SendStatus:       log.SendStatus,
		SendError:        log.SendError,
		CreatedOn:        log.ConsumeTime.String(),
	}, nil
}

// CleanOldLogs 清理旧日志
func (s *ConsumeLogService) CleanOldLogs(days int) (int64, error) {
	// TODO: Phase 2 实现
	return 0, nil
}

// GetConsumeStats 获取订阅消费统计
func (s *ConsumeLogService) GetConsumeStats(subscriptionID string) (map[string]interface{}, error) {
	totalConsume, err := models.GetConsumeLogsTotal(subscriptionID, "", "", "", "")
	if err != nil {
		return nil, err
	}
	totalMatched, err := models.GetConsumeLogsTotal(subscriptionID, "1", "", "", "")
	if err != nil {
		return nil, err
	}
	totalSent, err := models.GetConsumeLogsTotal(subscriptionID, "", "1", "", "")
	if err != nil {
		return nil, err
	}
	totalFailed, err := models.GetConsumeLogsTotal(subscriptionID, "", "2", "", "")
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_consume": totalConsume,
		"total_matched": totalMatched,
		"total_sent":    totalSent,
		"total_failed":  totalFailed,
	}, nil
}

// ResolveSubscriptionIDByName 支持按订阅名称搜索，取首个精确/模糊匹配的订阅ID
func (s *ConsumeLogService) ResolveSubscriptionIDByName(subscriptionName string) string {
	name := strings.TrimSpace(subscriptionName)
	if name == "" {
		return ""
	}
	// 尝试精确命中 ID（用户可能直接输入了订阅ID）
	if strings.HasPrefix(strings.ToUpper(name), "SUB") {
		return name
	}

	// 取第一页最大100条进行名称匹配（管理端筛选场景）
	subs, err := models.GetSubscriptions(1, 100, name, "", "", "")
	if err != nil || len(subs) == 0 {
		return ""
	}
	// 优先精确名称匹配
	for _, sub := range subs {
		if sub.Name == name {
			return sub.ID
		}
	}
	return subs[0].ID
}

// ParseStatusValue 对字符串状态值做容错转换
func (s *ConsumeLogService) ParseStatusValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "all" {
		return ""
	}
	if _, err := strconv.Atoi(v); err == nil {
		return v
	}
	return ""
}

func encodeExtractedValues(v map[string]string) string {
	if len(v) == 0 {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
