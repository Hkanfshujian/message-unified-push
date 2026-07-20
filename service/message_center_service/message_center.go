package message_center_service

import (
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/util"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	CategoryAll    = "all"
	CategorySystem = "system"
	CategoryPush   = "push"
)

type TargetScopeInput struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
}

type QueryOptions struct {
	Category   string
	Type       string
	ReadStatus string
	StartTime  string
	EndTime    string
	PageNum    int
	PageSize   int
}

type MessageListItem struct {
	ID             string `json:"id"`
	Category       string `json:"category"`
	Type           string `json:"type"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	Content        string `json:"content"`
	Time           string `json:"time"`
	IsRead         bool   `json:"is_read"`
	IsPinned       bool   `json:"is_pinned"`
	ThumbnailURL   string `json:"thumbnail_url"`
	ThumbnailRatio string `json:"thumbnail_ratio"`
	SourceSubject  string `json:"source_subject"`
	TargetURL      string `json:"target_url"`
}

type DeliveryEventItem struct {
	EventID     string `json:"event_id"`
	Category    string `json:"category"`
	MessageID   string `json:"message_id"`
	EventType   string `json:"event_type"`
	SyncVersion int64  `json:"sync_version"`
	OccurredAt  string `json:"occurred_at"`
}

type AdminMessagePayload struct {
	ID                 string             `json:"id"`
	Type               string             `json:"type"`
	Title              string             `json:"title"`
	Summary            string             `json:"summary"`
	Content            string             `json:"content"`
	Status             string             `json:"status"`
	PublishTime        string             `json:"publish_time"`
	EffectiveStartTime string             `json:"effective_start_time"`
	EffectiveEndTime   string             `json:"effective_end_time"`
	IsPinned           bool               `json:"is_pinned"`
	TargetScopes       []TargetScopeInput `json:"target_scopes"`
}

type MessageCenterService struct{}

func NormalizeCategory(category string) string {
	category = strings.TrimSpace(category)
	if category == CategorySystem || category == CategoryPush || category == CategoryAll {
		return category
	}
	return CategoryAll
}

func DisplayCount(count int64) string {
	if count <= 0 {
		return "0"
	}
	if count > 99 {
		return "99+"
	}
	return strconv.FormatInt(count, 10)
}

func NormalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func (s MessageCenterService) VerifyTargetScopeAdapters() map[string]interface{} {
	return map[string]interface{}{
		"user":       "models.Auth.id",
		"role":       "models.RbacRole.id",
		"group":      "models.RbacUserGroup.id",
		"department": "external-compatible target_id string",
		"position":   "external-compatible target_id string",
	}
}

func (s MessageCenterService) GetUnreadCount(userID int, category string) (int64, int64, error) {
	if err := s.ensureVisibleSystemRecipientStates(userID); err != nil {
		return 0, 0, err
	}
	query := models.GetDB().Model(&models.MessageRecipientState{}).Where("user_id = ? AND is_read = ? AND is_deleted = ?", userID, false, false)
	category = NormalizeCategory(category)
	if category != CategoryAll {
		query = query.Where("message_category = ?", category)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, 0, err
	}
	version, err := s.GetLatestSyncVersion(userID)
	return count, version, err
}

func (s MessageCenterService) GetLatestSyncVersion(userID int) (int64, error) {
	var version int64
	err := models.GetDB().Model(&models.MessageDeliveryEvent{}).Where("user_id = ?", userID).Select("COALESCE(MAX(sync_version), 0)").Scan(&version).Error
	return version, err
}

func (s MessageCenterService) ListMessages(userID int, options QueryOptions) ([]MessageListItem, int64, int64, error) {
	category := NormalizeCategory(options.Category)
	if category == CategoryAll {
		category = CategorySystem
	}
	if category == CategoryPush {
		return s.listPushMessages(userID, options)
	}
	return s.listSystemMessages(userID, options)
}

func (s MessageCenterService) MarkRead(userID int, category string, messageID string) (int64, int64, error) {
	return s.updateState(userID, NormalizeCategory(category), messageID, "read")
}

func (s MessageCenterService) DeleteForUser(userID int, category string, messageID string) (int64, int64, error) {
	return s.updateState(userID, NormalizeCategory(category), messageID, "deleted")
}

func (s MessageCenterService) MarkAllRead(userID int, options QueryOptions) (int64, int64, int64, error) {
	category := NormalizeCategory(options.Category)
	query := models.GetDB().Model(&models.MessageRecipientState{}).Where("user_id = ? AND is_read = ? AND is_deleted = ?", userID, false, false)
	if category != CategoryAll {
		query = query.Where("message_category = ?", category)
	}
	var states []models.MessageRecipientState
	if err := query.Find(&states).Error; err != nil {
		return 0, 0, 0, err
	}
	version := int64(0)
	for _, state := range states {
		v, err := s.nextSyncVersion(userID)
		if err != nil {
			return 0, 0, 0, err
		}
		version = v
		now := util.TimeNow()
		if err := models.GetDB().Model(&models.MessageRecipientState{}).Where("id = ?", state.ID).Updates(map[string]interface{}{"is_read": true, "read_at": now, "sync_version": v}).Error; err != nil {
			return 0, 0, 0, err
		}
		_ = s.recordEvent(userID, state.MessageCategory, state.MessageID, "read", v)
	}
	unread, latest, err := s.GetUnreadCount(userID, category)
	if latest > version {
		version = latest
	}
	return int64(len(states)), unread, version, err
}

func (s MessageCenterService) Sync(userID int, afterVersion int64) ([]DeliveryEventItem, int64, int64, error) {
	var events []models.MessageDeliveryEvent
	err := models.GetDB().Where("user_id = ? AND sync_version > ?", userID, afterVersion).Order("sync_version ASC").Limit(100).Find(&events).Error
	if err != nil {
		return nil, 0, 0, err
	}
	result := make([]DeliveryEventItem, 0, len(events))
	latest := afterVersion
	for _, event := range events {
		if event.SyncVersion > latest {
			latest = event.SyncVersion
		}
		result = append(result, DeliveryEventItem{EventID: strconv.FormatUint(uint64(event.ID), 10), Category: event.MessageCategory, MessageID: event.MessageID, EventType: event.EventType, SyncVersion: event.SyncVersion, OccurredAt: event.OccurredAt.String()})
	}
	unread, currentLatest, err := s.GetUnreadCount(userID, CategoryAll)
	if currentLatest > latest {
		latest = currentLatest
	}
	return result, unread, latest, err
}

func (s MessageCenterService) CreateSystemMessage(payload AdminMessagePayload, operator string) error {
	messageID := payload.ID
	if strings.TrimSpace(messageID) == "" {
		var err error
		messageID, err = util.GenerateRandomString(12)
		if err != nil {
			return err
		}
	}
	status := normalizeAdminStatus(payload.Status)
	message := models.SystemMessage{ID: messageID, Type: defaultString(payload.Type, "announcement"), Title: strings.TrimSpace(payload.Title), Summary: strings.TrimSpace(payload.Summary), Content: payload.Content, PublishTime: parseTimeOrNow(payload.PublishTime), EffectiveStartTime: parseTimeOrNow(payload.EffectiveStartTime), EffectiveEndTime: parseTimeOrFuture(payload.EffectiveEndTime), IsPinned: payload.IsPinned, Status: status, CreatedBy: operator, ModifiedBy: operator}
	// The message and its scope rows form one write boundary.
	return models.WithTransaction(func(tx *gorm.DB) error {
		if err := tx.Create(&message).Error; err != nil {
			return err
		}
		if err := s.replaceScopesWithTx(tx, CategorySystem, message.ID, payload.TargetScopes); err != nil {
			return err
		}
		if status != "published" {
			return nil
		}
		return s.refreshRecipientsWithTx(tx, message.ID, payload.TargetScopes, "created")
	})
}

func (s MessageCenterService) UpdateSystemMessage(payload AdminMessagePayload, operator string) error {
	message, err := models.GetSystemMessageByID(payload.ID)
	if err != nil {
		return err
	}
	if message.Status == "published" {
		return fmt.Errorf("已发布通知仅支持查看")
	}
	status := normalizeAdminStatus(payload.Status)
	data := map[string]interface{}{"type": defaultString(payload.Type, message.Type), "title": strings.TrimSpace(payload.Title), "summary": strings.TrimSpace(payload.Summary), "content": payload.Content, "status": status, "publish_time": parseTimeOrNow(payload.PublishTime), "effective_start_time": parseTimeOrNow(payload.EffectiveStartTime), "effective_end_time": parseTimeOrFuture(payload.EffectiveEndTime), "is_pinned": payload.IsPinned, "modified_by": operator}
	return models.WithTransaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.SystemMessage{}).Where("id = ?", payload.ID).Updates(data).Error; err != nil {
			return err
		}
		if err := s.replaceScopesWithTx(tx, CategorySystem, payload.ID, payload.TargetScopes); err != nil {
			return err
		}
		if status != "published" {
			return nil
		}
		return s.refreshRecipientsWithTx(tx, payload.ID, payload.TargetScopes, "scope_changed")
	})
}

func (s MessageCenterService) DeleteSystemMessage(id string, operator string) error {
	return s.DeleteSystemMessages([]string{id}, operator)
}

func (s MessageCenterService) DeleteSystemMessages(ids []string, _ string) error {
	cleanIDs := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" && !seen[id] {
			cleanIDs = append(cleanIDs, id)
			seen[id] = true
		}
	}
	if len(cleanIDs) == 0 {
		return fmt.Errorf("请选择要删除的通知")
	}
	return models.WithTransaction(func(tx *gorm.DB) error {
		var states []models.MessageRecipientState
		if err := tx.Where("message_category = ? AND message_id IN ?", CategorySystem, cleanIDs).Find(&states).Error; err != nil {
			return err
		}
		for _, state := range states {
			version, err := s.nextSyncVersionWithTx(tx, state.UserID)
			if err != nil {
				return err
			}
			if err := tx.Create(&models.MessageDeliveryEvent{UserID: state.UserID, MessageCategory: CategorySystem, MessageID: state.MessageID, EventType: "deleted", SyncVersion: version}).Error; err != nil {
				return err
			}
		}
		if err := tx.Unscoped().Where("message_category = ? AND message_id IN ?", CategorySystem, cleanIDs).Delete(&models.MessageRecipientState{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("message_category = ? AND message_id IN ?", CategorySystem, cleanIDs).Delete(&models.MessageTargetScope{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Where("id IN ?", cleanIDs).Delete(&models.SystemMessage{}).Error
	})
}

func (s MessageCenterService) ListAdminMessages(keyword, messageType, status, startTime, endTime string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	page, pageSize = NormalizePage(page, pageSize)
	query := models.GetDB().Model(&models.SystemMessage{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR summary LIKE ? OR content LIKE ?", like, like, like)
	}
	if messageType != "" {
		query = query.Where("type = ?", messageType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startTime != "" {
		query = query.Where("publish_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("publish_time <= ?", endTime)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var messages []models.SystemMessage
	if err := query.Order("is_pinned DESC, publish_time DESC, created_on DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, 0, err
	}
	list := make([]map[string]interface{}, 0, len(messages))
	for _, message := range messages {
		var scopes []models.MessageTargetScope
		_ = models.GetDB().Where("message_category = ? AND message_id = ?", CategorySystem, message.ID).Find(&scopes).Error
		list = append(list, map[string]interface{}{"id": message.ID, "type": message.Type, "title": message.Title, "summary": message.Summary, "content": message.Content, "publish_time": message.PublishTime.String(), "effective_start_time": message.EffectiveStartTime.String(), "effective_end_time": message.EffectiveEndTime.String(), "is_pinned": message.IsPinned, "status": message.Status, "target_scopes": scopes, "created_on": message.CreatedAt.String(), "modified_on": message.UpdatedAt.String()})
	}
	return list, total, nil
}

func normalizeAdminStatus(status string) string {
	if strings.TrimSpace(status) == "draft" {
		return "draft"
	}
	return "published"
}

func (s MessageCenterService) ProjectRecentBusinessPushes(limit int) error {
	if limit <= 0 {
		limit = 50
	}
	var logs []models.SendTasksLogs
	if err := models.GetDB().Order("created_on DESC").Limit(limit).Find(&logs).Error; err != nil {
		return err
	}
	for _, log := range logs {
		id := fmt.Sprintf("log-%d", log.ID)
		if _, err := models.GetBusinessPushMessageByID(id); err == nil {
			continue
		}
		typeName := "template_send"
		if log.Type == "cron_message" || log.Type == "cron" {
			typeName = "cron_sent"
		}
		title := defaultString(log.Name, "业务推送")
		message := models.BusinessPushMessage{ID: id, Type: typeName, Title: title, Summary: strings.TrimSpace(firstLine(log.Log)), Content: log.Log, SourceSubject: title, SourceType: log.Type, SourceID: log.TaskID, PushTime: log.CreatedAt, TargetURL: "/logs/task", ThumbnailRatio: "16:9"}
		_ = models.AddBusinessPushMessage(&message)
	}
	return nil
}

func (s MessageCenterService) listSystemMessages(userID int, options QueryOptions) ([]MessageListItem, int64, int64, error) {
	if err := s.ensureVisibleSystemRecipientStates(userID); err != nil {
		return nil, 0, 0, err
	}
	page, pageSize := NormalizePage(options.PageNum, options.PageSize)
	now := util.TimeNow()
	query := models.GetDB().Table(models.GetSchema(models.MessageRecipientState{})+" AS rs").Select("sm.*, rs.is_read").Joins("JOIN "+models.GetSchema(models.SystemMessage{})+" AS sm ON sm.id = rs.message_id").Where("rs.user_id = ? AND rs.message_category = ? AND rs.is_deleted = ? AND sm.status <> ? AND sm.effective_start_time <= ? AND sm.effective_end_time >= ?", userID, CategorySystem, false, "deleted", now, now)
	query = applyCommonFilters(query, options, "sm", "rs")
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}
	type row struct {
		models.SystemMessage
		IsRead bool `json:"is_read"`
	}
	var rows []row
	if err := query.Order("sm.is_pinned DESC, sm.publish_time DESC, sm.created_on DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, 0, 0, err
	}
	list := make([]MessageListItem, 0, len(rows))
	for _, item := range rows {
		list = append(list, MessageListItem{ID: item.ID, Category: CategorySystem, Type: item.Type, Title: item.Title, Summary: item.Summary, Content: item.Content, Time: item.PublishTime.String(), IsRead: item.IsRead, IsPinned: item.IsPinned})
	}
	version, err := s.GetLatestSyncVersion(userID)
	return list, total, version, err
}

func (s MessageCenterService) listPushMessages(userID int, options QueryOptions) ([]MessageListItem, int64, int64, error) {
	_ = s.ProjectRecentBusinessPushes(80)
	page, pageSize := NormalizePage(options.PageNum, options.PageSize)
	query := models.GetDB().Table(models.GetSchema(models.MessageRecipientState{})+" AS rs").Select("bp.*, rs.is_read").Joins("JOIN "+models.GetSchema(models.BusinessPushMessage{})+" AS bp ON bp.id = rs.message_id").Where("rs.user_id = ? AND rs.message_category = ? AND rs.is_deleted = ?", userID, CategoryPush, false)
	query = applyCommonFilters(query, options, "bp", "rs")
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}
	type row struct {
		models.BusinessPushMessage
		IsRead bool `json:"is_read"`
	}
	var rows []row
	if err := query.Order("bp.push_time DESC, bp.created_on DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, 0, 0, err
	}
	list := make([]MessageListItem, 0, len(rows))
	for _, item := range rows {
		list = append(list, MessageListItem{ID: item.ID, Category: CategoryPush, Type: item.Type, Title: item.Title, Summary: item.Summary, Content: item.Content, Time: item.PushTime.String(), IsRead: item.IsRead, ThumbnailURL: item.ThumbnailURL, ThumbnailRatio: item.ThumbnailRatio, SourceSubject: item.SourceSubject, TargetURL: item.TargetURL})
	}
	version, err := s.GetLatestSyncVersion(userID)
	return list, total, version, err
}

func applyCommonFilters(query *gorm.DB, options QueryOptions, alias string, stateAlias string) *gorm.DB {
	if options.Type != "" {
		query = query.Where(alias+".type = ?", options.Type)
	}
	if options.ReadStatus == "read" {
		query = query.Where(stateAlias+".is_read = ?", true)
	}
	if options.ReadStatus == "unread" {
		query = query.Where(stateAlias+".is_read = ?", false)
	}
	if options.StartTime != "" {
		query = query.Where(alias+".created_on >= ?", options.StartTime)
	}
	if options.EndTime != "" {
		query = query.Where(alias+".created_on <= ?", options.EndTime)
	}
	return query
}

func (s MessageCenterService) updateState(userID int, category string, messageID string, action string) (int64, int64, error) {
	if category == CategorySystem {
		if err := s.ensureVisibleSystemRecipientState(userID, messageID); err != nil {
			return 0, 0, err
		}
	}
	version, err := s.nextSyncVersion(userID)
	if err != nil {
		return 0, 0, err
	}
	data := map[string]interface{}{"sync_version": version}
	now := util.TimeNow()
	eventType := "read"
	if action == "deleted" {
		data["is_deleted"] = true
		data["deleted_at"] = now
		eventType = "deleted"
	} else {
		data["is_read"] = true
		data["read_at"] = now
	}
	err = models.GetDB().Model(&models.MessageRecipientState{}).Where("user_id = ? AND message_category = ? AND message_id = ?", userID, category, messageID).Updates(data).Error
	if err != nil {
		return 0, 0, err
	}
	_ = s.recordEvent(userID, category, messageID, eventType, version)
	unread, latest, err := s.GetUnreadCount(userID, CategoryAll)
	if latest > version {
		version = latest
	}
	return unread, version, err
}

func (s MessageCenterService) ensureVisibleSystemRecipientStates(userID int) error {
	scopes, err := s.currentUserTargetScopes(userID)
	if err != nil {
		return err
	}
	now := util.TimeNow()
	var messages []models.SystemMessage
	query := models.GetDB().Model(&models.SystemMessage{}).Joins("JOIN "+models.GetSchema(models.MessageTargetScope{})+" AS mts ON mts.message_id = "+models.GetSchema(models.SystemMessage{})+".id AND mts.message_category = ?", CategorySystem).Where(models.GetSchema(models.SystemMessage{})+".status = ? AND "+models.GetSchema(models.SystemMessage{})+".effective_start_time <= ? AND "+models.GetSchema(models.SystemMessage{})+".effective_end_time >= ?", "published", now, now)
	query = query.Where(scopeVisibilityCondition(scopes), scopeVisibilityArgs(scopes)...)
	if err := query.Group(models.GetSchema(models.SystemMessage{}) + ".id").Find(&messages).Error; err != nil {
		return err
	}
	return models.WithTransaction(func(tx *gorm.DB) error {
		for _, message := range messages {
			if err := s.ensureSystemRecipientStateWithTx(tx, userID, message.ID, false); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s MessageCenterService) ensureVisibleSystemRecipientState(userID int, messageID string) error {
	scopes, err := s.currentUserTargetScopes(userID)
	if err != nil {
		return err
	}
	now := util.TimeNow()
	var count int64
	query := models.GetDB().Model(&models.SystemMessage{}).Joins("JOIN "+models.GetSchema(models.MessageTargetScope{})+" AS mts ON mts.message_id = "+models.GetSchema(models.SystemMessage{})+".id AND mts.message_category = ?", CategorySystem).Where(models.GetSchema(models.SystemMessage{})+".id = ? AND "+models.GetSchema(models.SystemMessage{})+".status = ? AND "+models.GetSchema(models.SystemMessage{})+".effective_start_time <= ? AND "+models.GetSchema(models.SystemMessage{})+".effective_end_time >= ?", messageID, "published", now, now)
	query = query.Where(scopeVisibilityCondition(scopes), scopeVisibilityArgs(scopes)...)
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("消息不存在或当前用户不可见")
	}
	return models.WithTransaction(func(tx *gorm.DB) error {
		return s.ensureSystemRecipientStateWithTx(tx, userID, messageID, true)
	})
}

func (s MessageCenterService) ensureSystemRecipientStateWithTx(tx *gorm.DB, userID int, messageID string, recordCreatedEvent bool) error {
	var state models.MessageRecipientState
	err := tx.Where("user_id = ? AND message_category = ? AND message_id = ?", userID, CategorySystem, messageID).First(&state).Error
	if err == nil {
		if state.IsDeleted {
			version, versionErr := s.nextSyncVersionWithTx(tx, userID)
			if versionErr != nil {
				return versionErr
			}
			return tx.Model(&models.MessageRecipientState{}).Where("id = ?", state.ID).Updates(map[string]interface{}{"is_deleted": false, "sync_version": version}).Error
		}
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	version, versionErr := s.nextSyncVersionWithTx(tx, userID)
	if versionErr != nil {
		return versionErr
	}
	if err := tx.Create(&models.MessageRecipientState{UserID: userID, MessageCategory: CategorySystem, MessageID: messageID, SyncVersion: version}).Error; err != nil {
		return err
	}
	if recordCreatedEvent {
		return tx.Create(&models.MessageDeliveryEvent{UserID: userID, MessageCategory: CategorySystem, MessageID: messageID, EventType: "created", SyncVersion: version}).Error
	}
	return nil
}

func (s MessageCenterService) currentUserTargetScopes(userID int) ([]TargetScopeInput, error) {
	scopes := []TargetScopeInput{{TargetType: "all", TargetID: "all"}, {TargetType: "user", TargetID: strconv.Itoa(userID)}}
	roleIDs, err := models.GetRoleIDsByUserID(userID)
	if err != nil {
		return nil, err
	}
	for _, roleID := range roleIDs {
		scopes = append(scopes, TargetScopeInput{TargetType: "role", TargetID: strconv.FormatUint(uint64(roleID), 10)})
	}
	groupIDs, err := models.GetGroupIDsByUserID(userID)
	if err != nil {
		return nil, err
	}
	for _, groupID := range groupIDs {
		id := strconv.FormatUint(uint64(groupID), 10)
		scopes = append(scopes, TargetScopeInput{TargetType: "group", TargetID: id}, TargetScopeInput{TargetType: "department", TargetID: id}, TargetScopeInput{TargetType: "position", TargetID: id})
	}
	return scopes, nil
}

func scopeVisibilityCondition(scopes []TargetScopeInput) string {
	conditions := make([]string, 0, len(scopes))
	for range scopes {
		conditions = append(conditions, "(mts.target_type = ? AND mts.target_id = ?)")
	}
	if len(conditions) == 0 {
		return "1 = 0"
	}
	return "(" + strings.Join(conditions, " OR ") + ")"
}

func scopeVisibilityArgs(scopes []TargetScopeInput) []interface{} {
	args := make([]interface{}, 0, len(scopes)*2)
	for _, scope := range scopes {
		args = append(args, scope.TargetType, scope.TargetID)
	}
	return args
}

func (s MessageCenterService) nextSyncVersion(userID int) (int64, error) {
	return s.nextSyncVersionWithTx(models.GetDB(), userID)
}

func (s MessageCenterService) nextSyncVersionWithTx(tx *gorm.DB, userID int) (int64, error) {
	var latest int64
	err := tx.Model(&models.MessageDeliveryEvent{}).Where("user_id = ?", userID).Select("COALESCE(MAX(sync_version), 0)").Scan(&latest).Error
	return latest + 1, err
}

func (s MessageCenterService) recordEvent(userID int, category string, messageID string, eventType string, version int64) error {
	return models.AddMessageDeliveryEvent(&models.MessageDeliveryEvent{UserID: userID, MessageCategory: category, MessageID: messageID, EventType: eventType, SyncVersion: version})
}

func (s MessageCenterService) replaceScopesWithTx(tx *gorm.DB, category string, messageID string, inputs []TargetScopeInput) error {
	if err := tx.Where("message_category = ? AND message_id = ?", category, messageID).Delete(&models.MessageTargetScope{}).Error; err != nil {
		return err
	}
	if len(inputs) == 0 {
		inputs = []TargetScopeInput{{TargetType: "all", TargetID: "all"}}
	}
	scopes := make([]models.MessageTargetScope, 0, len(inputs))
	seen := map[string]bool{}
	for _, input := range inputs {
		typeName := strings.TrimSpace(input.TargetType)
		targetID := strings.TrimSpace(input.TargetID)
		if typeName == "" || targetID == "" {
			continue
		}
		key := typeName + ":" + targetID
		if seen[key] {
			continue
		}
		seen[key] = true
		scopes = append(scopes, models.MessageTargetScope{MessageCategory: category, MessageID: messageID, TargetType: typeName, TargetID: targetID})
	}
	if len(scopes) == 0 {
		return nil
	}
	return tx.Create(&scopes).Error
}

func (s MessageCenterService) refreshRecipientsWithTx(tx *gorm.DB, messageID string, scopes []TargetScopeInput, eventType string) error {
	userIDs, err := resolveTargetUsers(tx, scopes)
	if err != nil {
		return err
	}
	for _, userID := range userIDs {
		var state models.MessageRecipientState
		err := tx.Where("user_id = ? AND message_category = ? AND message_id = ?", userID, CategorySystem, messageID).First(&state).Error
		version, versionErr := s.nextSyncVersionWithTx(tx, userID)
		if versionErr != nil {
			return versionErr
		}
		if err == nil {
			if updateErr := tx.Model(&models.MessageRecipientState{}).Where("id = ?", state.ID).Updates(map[string]interface{}{"is_deleted": false, "sync_version": version}).Error; updateErr != nil {
				return updateErr
			}
		} else {
			if createErr := tx.Create(&models.MessageRecipientState{UserID: userID, MessageCategory: CategorySystem, MessageID: messageID, SyncVersion: version}).Error; createErr != nil {
				return createErr
			}
		}
		if eventErr := tx.Create(&models.MessageDeliveryEvent{UserID: userID, MessageCategory: CategorySystem, MessageID: messageID, EventType: eventType, SyncVersion: version}).Error; eventErr != nil {
			return eventErr
		}
	}
	return nil
}

func resolveTargetUsers(tx *gorm.DB, scopes []TargetScopeInput) ([]int, error) {
	ids := map[int]bool{}
	if len(scopes) == 0 {
		var users []models.Auth
		if err := tx.Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			ids[user.ID] = true
		}
	}
	for _, scope := range scopes {
		targetID := strings.TrimSpace(scope.TargetID)
		switch strings.TrimSpace(scope.TargetType) {
		case "all":
			var users []models.Auth
			if err := tx.Find(&users).Error; err != nil {
				return nil, err
			}
			for _, user := range users {
				ids[user.ID] = true
			}
		case "user":
			if id, err := strconv.Atoi(targetID); err == nil {
				ids[id] = true
			}
		case "role":
			var rels []models.RbacUserRole
			if err := tx.Where("role_id = ?", targetID).Find(&rels).Error; err != nil {
				return nil, err
			}
			for _, rel := range rels {
				ids[rel.UserID] = true
			}
		case "group", "department", "position":
			var rels []models.RbacUserGroupMember
			if err := tx.Where("group_id = ?", targetID).Find(&rels).Error; err != nil {
				return nil, err
			}
			for _, rel := range rels {
				ids[rel.UserID] = true
			}
		}
	}
	result := make([]int, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	return result, nil
}

func parseTimeOrNow(value string) util.Time {
	if t, ok := parseTime(value); ok {
		return util.Time(t)
	}
	return util.TimeNow()
}

func parseTimeOrFuture(value string) util.Time {
	if t, ok := parseTime(value); ok {
		return util.Time(t)
	}
	return util.Time(util.GetNowTime().AddDate(1, 0, 0))
}

func parseTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	formats := []string{"2006-01-02 15:04:05", "2006-01-02T15:04", time.RFC3339, "2006-01-02"}
	for _, format := range formats {
		if t, err := time.ParseInLocation(format, value, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "业务推送已完成"
	}
	if idx := strings.Index(value, "\n"); idx >= 0 {
		return value[:idx]
	}
	if len(value) > 200 {
		return value[:200]
	}
	return value
}
