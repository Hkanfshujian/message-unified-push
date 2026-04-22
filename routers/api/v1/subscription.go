package v1

import (
	"ops-message-unified-push/service/subscription_rule"
	"net/http"
	"strconv"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/e"
	"ops-message-unified-push/service/subscription_service"

	"github.com/gin-gonic/gin"
)

type SubscriptionController struct{}

// GetSubscriptionList 获取订阅列表
func (c *SubscriptionController) GetSubscriptionList(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	name := ctx.Query("name")
	status := ctx.Query("status")
	sourceID := ctx.Query("source_id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	service := subscription_service.SubscriptionService{}
	subscriptions, total, err := service.GetAll(name, status, sourceID, page, pageSize)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_GET_SUBSCRIPTION_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", map[string]interface{}{
		"list":      subscriptions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AddSubscription 新增订阅
func (c *SubscriptionController) AddSubscription(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	var req subscription_service.AddSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, e.GetMsg(e.INVALID_PARAMS), nil)
		return
	}

	service := subscription_service.SubscriptionService{}
	subscription, err := service.Add(req)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_ADD_SUBSCRIPTION_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", subscription)
}

// EditSubscription 编辑订阅
func (c *SubscriptionController) EditSubscription(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	id := ctx.Param("id")
	var req subscription_service.EditSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, e.GetMsg(e.INVALID_PARAMS), nil)
		return
	}

	service := subscription_service.SubscriptionService{}
	err := service.Edit(id, req)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "编辑订阅失败: "+err.Error(), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", nil)
}

// DeleteSubscription 删除订阅
func (c *SubscriptionController) DeleteSubscription(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	id := ctx.Param("id")

	service := subscription_service.SubscriptionService{}
	err := service.Delete(id)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_DELETE_SUBSCRIPTION_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", nil)
}

// StartSubscription 启动订阅
func (c *SubscriptionController) StartSubscription(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	id := ctx.Param("id")

	service := subscription_service.SubscriptionService{}
	err := service.Start(id)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_START_SUBSCRIPTION_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", nil)
}

// StopSubscription 停止订阅
func (c *SubscriptionController) StopSubscription(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	id := ctx.Param("id")

	service := subscription_service.SubscriptionService{}
	err := service.Stop(id)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_STOP_SUBSCRIPTION_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", nil)
}

// GetSubscriptionByID 获取订阅详情
func (c *SubscriptionController) GetSubscriptionByID(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	id := ctx.Param("id")

	service := subscription_service.SubscriptionService{}
	subscription, err := service.GetByID(id)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_GET_SUBSCRIPTION_FAIL), nil)
		return
	}
	appG.CResponse(http.StatusOK, "ok", map[string]interface{}{
		"id":                     subscription.ID,
		"source_id":              subscription.SourceID,
		"name":                   subscription.Name,
		"topic":                  subscription.Topic,
		"tag":                    subscription.Tag,
		"group_name":             subscription.GroupName,
		"validate_regex":         subscription.ValidateRegex,
		"extract_regex":          subscription.ExtractRegex,
		"extract_field":          subscription.ExtractField,
		"extract_rules":          subscription_rule.ParseStoredExtractRules(subscription.ExtractRegex, subscription.ExtractField),
		"template_id":            subscription.TemplateID,
		"consume_mode":           subscription.ConsumeMode,
		"template_content_type":  subscription.ConsumeMode,
		"status":                 subscription.Status,
	})
}

// TestSubscriptionRegex 测试订阅正则
func (c *SubscriptionController) TestSubscriptionRegex(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	var req subscription_service.RegexTestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, e.GetMsg(e.INVALID_PARAMS), nil)
		return
	}

	service := subscription_service.SubscriptionService{}
	result, err := service.TestRegex(req)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, err.Error(), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", result)
}

