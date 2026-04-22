package v1

import (
	"net/http"
	"strconv"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/e"
	"ops-message-unified-push/service/mq_source_service"

	"github.com/gin-gonic/gin"
)

type MQSourceController struct{}

// GetMQSourceList 获取消息队列数据源列表
func (c *MQSourceController) GetMQSourceList(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	name := ctx.Query("name")
	status := ctx.Query("status")
	mqType := ctx.Query("type")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	service := mq_source_service.MQSourceService{}
	sources, total, err := service.GetAll(name, status, mqType, page, pageSize)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_GET_SOURCE_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", map[string]interface{}{
		"list":      sources,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AddMQSource 新增消息队列数据源
func (c *MQSourceController) AddMQSource(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	var req mq_source_service.AddMQSourceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, e.GetMsg(e.INVALID_PARAMS), nil)
		return
	}

	service := mq_source_service.MQSourceService{}
	source, err := service.Add(req)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_ADD_SOURCE_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", source)
}

// EditMQSource 编辑消息队列数据源
func (c *MQSourceController) EditMQSource(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	id := ctx.Param("id")
	var req mq_source_service.EditMQSourceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, e.GetMsg(e.INVALID_PARAMS), nil)
		return
	}

	service := mq_source_service.MQSourceService{}
	err := service.Edit(id, req)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_EDIT_SOURCE_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", nil)
}

// DeleteMQSource 删除消息队列数据源
func (c *MQSourceController) DeleteMQSource(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	id := ctx.Param("id")

	service := mq_source_service.MQSourceService{}
	err := service.Delete(id)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_DELETE_SOURCE_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", nil)
}

// TestMQSource 测试消息队列连接
func (c *MQSourceController) TestMQSource(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	id := ctx.Param("id")

	service := mq_source_service.MQSourceService{}
	err := service.TestConnectionByID(id)
	if err != nil {
		appG.CResponse(http.StatusOK, "ok", map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	appG.CResponse(http.StatusOK, "ok", map[string]interface{}{
		"success": true,
		"message": "连接测试成功",
	})
}

// TestMQSourceConfig 测试消息队列连接配置（无需保存）
func (c *MQSourceController) TestMQSourceConfig(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	var req mq_source_service.TestConnectionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, e.GetMsg(e.INVALID_PARAMS), nil)
		return
	}

	service := mq_source_service.MQSourceService{}
	err := service.TestConnectionDirect(req.Type, req.NamesrvAddr, req.AccessKey, req.SecretKey)
	if err != nil {
		appG.CResponse(http.StatusOK, "ok", map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	appG.CResponse(http.StatusOK, "ok", map[string]interface{}{
		"success": true,
		"message": "连接测试成功",
	})
}

// GetMQSourceByID 获取单个数据源详情
func (c *MQSourceController) GetMQSourceByID(ctx *gin.Context) {
	var (
		appG = app.Gin{C: ctx}
	)

	id := ctx.Param("id")

	service := mq_source_service.MQSourceService{}
	source, err := service.GetByID(id)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR_GET_SOURCE_FAIL), nil)
		return
	}

	appG.CResponse(http.StatusOK, "ok", source)
}

