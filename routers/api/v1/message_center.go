package v1

import (
	"fmt"
	"net/http"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/service/message_center_service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func currentMessageUserID(c *gin.Context) (int, error) {
	if value, ok := c.Get("userID"); ok {
		switch typed := value.(type) {
		case int:
			return typed, nil
		case uint:
			return int(typed), nil
		}
	}
	username := app.GetCurrentUserName(c)
	if strings.TrimSpace(username) == "" {
		return 0, fmt.Errorf("未登录")
	}
	user, err := models.GetUserByUsername(username)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func messageQueryOptions(c *gin.Context) message_center_service.QueryOptions {
	page, _ := strconv.Atoi(c.DefaultQuery("page_num", c.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", c.DefaultQuery("size", "20")))
	return message_center_service.QueryOptions{
		Category:   c.DefaultQuery("category", "system"),
		Type:       c.Query("type"),
		ReadStatus: c.DefaultQuery("read_status", "all"),
		StartTime:  c.Query("start_time"),
		EndTime:    c.Query("end_time"),
		PageNum:    page,
		PageSize:   pageSize,
	}
}

func GetMessageCenterUnreadCount(c *gin.Context) {
	appG := app.Gin{C: c}
	userID, err := currentMessageUserID(c)
	if err != nil {
		appG.CResponse(http.StatusUnauthorized, "获取当前用户失败", nil)
		return
	}
	service := message_center_service.MessageCenterService{}
	count, version, err := service.GetUnreadCount(userID, c.DefaultQuery("category", "all"))
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取未读数量失败", nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取未读数量成功", map[string]interface{}{"unread_count": count, "display_count": message_center_service.DisplayCount(count), "sync_version": version})
}

func GetMessageCenterMessages(c *gin.Context) {
	appG := app.Gin{C: c}
	userID, err := currentMessageUserID(c)
	if err != nil {
		appG.CResponse(http.StatusUnauthorized, "获取当前用户失败", nil)
		return
	}
	service := message_center_service.MessageCenterService{}
	options := messageQueryOptions(c)
	list, total, version, err := service.ListMessages(userID, options)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取消息列表失败", nil)
		return
	}
	page, pageSize := message_center_service.NormalizePage(options.PageNum, options.PageSize)
	appG.CResponse(http.StatusOK, "获取消息列表成功", map[string]interface{}{"list": list, "lists": list, "total": total, "page_num": page, "page_size": pageSize, "has_more": int64(page*pageSize) < total, "sync_version": version})
}

func MarkMessageCenterMessageRead(c *gin.Context) {
	appG := app.Gin{C: c}
	userID, err := currentMessageUserID(c)
	if err != nil {
		appG.CResponse(http.StatusUnauthorized, "获取当前用户失败", nil)
		return
	}
	var req struct {
		Category  string `json:"category"`
		MessageID string `json:"message_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.MessageID) == "" {
		appG.CResponse(http.StatusBadRequest, "参数错误", nil)
		return
	}
	unread, version, err := (message_center_service.MessageCenterService{}).MarkRead(userID, req.Category, req.MessageID)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "标记已读失败", nil)
		return
	}
	appG.CResponse(http.StatusOK, "标记已读成功", map[string]interface{}{"message_id": req.MessageID, "is_read": true, "unread_count": unread, "sync_version": version})
}

func MarkAllMessageCenterMessagesRead(c *gin.Context) {
	appG := app.Gin{C: c}
	userID, err := currentMessageUserID(c)
	if err != nil {
		appG.CResponse(http.StatusUnauthorized, "获取当前用户失败", nil)
		return
	}
	var req struct {
		Category  string `json:"category"`
		Type      string `json:"type"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}
	_ = c.ShouldBindJSON(&req)
	affected, unread, version, err := (message_center_service.MessageCenterService{}).MarkAllRead(userID, message_center_service.QueryOptions{Category: req.Category, Type: req.Type, StartTime: req.StartTime, EndTime: req.EndTime})
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "全部已读失败", nil)
		return
	}
	appG.CResponse(http.StatusOK, "全部已读成功", map[string]interface{}{"affected_count": affected, "unread_count": unread, "sync_version": version})
}

func DeleteMessageCenterMessage(c *gin.Context) {
	appG := app.Gin{C: c}
	userID, err := currentMessageUserID(c)
	if err != nil {
		appG.CResponse(http.StatusUnauthorized, "获取当前用户失败", nil)
		return
	}
	var req struct {
		Category  string `json:"category"`
		MessageID string `json:"message_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.MessageID) == "" {
		appG.CResponse(http.StatusBadRequest, "参数错误", nil)
		return
	}
	unread, version, err := (message_center_service.MessageCenterService{}).DeleteForUser(userID, req.Category, req.MessageID)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "删除消息失败", nil)
		return
	}
	appG.CResponse(http.StatusOK, "删除消息成功", map[string]interface{}{"message_id": req.MessageID, "is_deleted": true, "unread_count": unread, "sync_version": version})
}

func SyncMessageCenter(c *gin.Context) {
	appG := app.Gin{C: c}
	userID, err := currentMessageUserID(c)
	if err != nil {
		appG.CResponse(http.StatusUnauthorized, "获取当前用户失败", nil)
		return
	}
	afterVersion, _ := strconv.ParseInt(c.DefaultQuery("after_version", "0"), 10, 64)
	events, unread, latest, err := (message_center_service.MessageCenterService{}).Sync(userID, afterVersion)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "同步消息失败", nil)
		return
	}
	appG.CResponse(http.StatusOK, "同步消息成功", map[string]interface{}{"events": events, "unread_count": unread, "latest_version": latest})
}

func GetSystemMessages(c *gin.Context) {
	appG := app.Gin{C: c}
	page, _ := strconv.Atoi(c.DefaultQuery("page_num", c.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", c.DefaultQuery("size", "20")))
	list, total, err := (message_center_service.MessageCenterService{}).ListAdminMessages(c.Query("keyword"), c.Query("type"), c.Query("status"), c.Query("start_time"), c.Query("end_time"), page, pageSize)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取系统通知失败", nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取系统通知成功", map[string]interface{}{"list": list, "lists": list, "total": total})
}

func AddSystemMessage(c *gin.Context) {
	appG := app.Gin{C: c}
	var req message_center_service.AdminMessagePayload
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Title) == "" {
		appG.CResponse(http.StatusBadRequest, "参数错误", nil)
		return
	}
	if err := (message_center_service.MessageCenterService{}).CreateSystemMessage(req, app.GetCurrentUserName(c)); err != nil {
		appG.CResponse(http.StatusInternalServerError, "新增系统通知失败："+err.Error(), nil)
		return
	}
	appG.CResponse(http.StatusOK, "新增系统通知成功", nil)
}

func EditSystemMessage(c *gin.Context) {
	appG := app.Gin{C: c}
	var req message_center_service.AdminMessagePayload
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.ID) == "" || strings.TrimSpace(req.Title) == "" {
		appG.CResponse(http.StatusBadRequest, "参数错误", nil)
		return
	}
	if err := (message_center_service.MessageCenterService{}).UpdateSystemMessage(req, app.GetCurrentUserName(c)); err != nil {
		appG.CResponse(http.StatusInternalServerError, "编辑系统通知失败："+err.Error(), nil)
		return
	}
	appG.CResponse(http.StatusOK, "编辑系统通知成功", nil)
}

func DeleteSystemMessage(c *gin.Context) {
	appG := app.Gin{C: c}
	var req struct {
		ID  string   `json:"id"`
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, "参数错误", nil)
		return
	}
	ids := req.IDs
	if strings.TrimSpace(req.ID) != "" {
		ids = append(ids, req.ID)
	}
	if len(ids) == 0 {
		appG.CResponse(http.StatusBadRequest, "请选择要删除的通知", nil)
		return
	}
	if err := (message_center_service.MessageCenterService{}).DeleteSystemMessages(ids, app.GetCurrentUserName(c)); err != nil {
		appG.CResponse(http.StatusInternalServerError, "删除系统通知失败："+err.Error(), nil)
		return
	}
	appG.CResponse(http.StatusOK, "删除系统通知成功", nil)
}
