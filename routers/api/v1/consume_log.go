package v1

import (
	"net/http"
	"strconv"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/e"
	"ops-message-unified-push/service/consume_log_service"

	"github.com/gin-gonic/gin"
)

type ConsumeLogController struct{}

// GetConsumeLogList 获取消费日志列表
func (c *ConsumeLogController) GetConsumeLogList(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	subscriptionID := ctx.Query("subscription_id")
	subscriptionName := ctx.Query("subscription_name")
	matched := ctx.Query("matched")
	sendStatus := ctx.Query("send_status")
	if sendStatus == "" {
		sendStatus = ctx.Query("status")
	}
	startTime := ctx.Query("start_time")
	endTime := ctx.Query("end_time")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	service := consume_log_service.ConsumeLogService{}
	if subscriptionID == "" && subscriptionName != "" {
		subscriptionID = service.ResolveSubscriptionIDByName(subscriptionName)
	}
	logs, total, err := service.GetConsumeLogList(
		subscriptionID,
		service.ParseStatusValue(sendStatus),
		service.ParseStatusValue(matched),
		startTime,
		endTime,
		page,
		pageSize,
	)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_GET_CONSUME_LOG_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", map[string]interface{}{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetConsumeLogByID 获取消费日志详情
func (c *ConsumeLogController) GetConsumeLogByID(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, e.GetMsg(e.INVALID_PARAMS), nil)
		return
	}

	service := consume_log_service.ConsumeLogService{}
	log, err := service.GetConsumeLogByID(uint(id))
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_GET_CONSUME_LOG_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", log)
}

// GetConsumeStats 获取消费统计
func (c *ConsumeLogController) GetConsumeStats(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	subscriptionID := ctx.Query("subscription_id")

	service := consume_log_service.ConsumeLogService{}
	stats, err := service.GetConsumeStats(subscriptionID)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_GET_CONSUME_LOG_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", stats)
}

