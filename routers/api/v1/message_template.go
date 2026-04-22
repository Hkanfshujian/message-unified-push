package v1

import (
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/e"
	"ops-message-unified-push/pkg/util"
	"ops-message-unified-push/service/message_template_service"
	"ops-message-unified-push/service/send_ins_service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetMessageTemplateList 获取消息模板列表
func GetMessageTemplateList(c *gin.Context) {
	appG := app.Gin{C: c}
	text := c.Query("text")
	status := c.Query("status")

	offset, limit := util.GetPageSize(c)
	templateService := message_template_service.TemplateService{
		Text:     text,
		Status:   status,
		PageNum:  offset,
		PageSize: limit,
	}

	templates, err := templateService.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取消息模板失败！", nil)
		return
	}

	count, err := templateService.Count()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取消息模板总数失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取消息模板成功", map[string]interface{}{
		"lists": templates,
		"total": count,
	})
}

// GetMessageTemplate 获取单个消息模板
func GetMessageTemplate(c *gin.Context) {
	appG := app.Gin{C: c}
	id := c.Query("id")

	templateService := message_template_service.TemplateService{
		ID: id,
	}

	exists, err := templateService.ExistByID()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "查询模板失败！", nil)
		return
	}

	if !exists {
		appG.CResponse(http.StatusNotFound, "模板不存在！", nil)
		return
	}

	template, err := templateService.Get()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取模板详情失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取模板详情成功", template)
}

// AddMessageTemplate 添加消息模板
func AddMessageTemplate(c *gin.Context) {
	appG := app.Gin{C: c}

	var req struct {
		ID               string `json:"id"`
		Name             string `json:"name" binding:"required"`
		Description      string `json:"description"`
		TextTemplate     string `json:"text_template"`
		HTMLTemplate     string `json:"html_template"`
		MarkdownTemplate string `json:"markdown_template"`
		Placeholders     string `json:"placeholders"`
		AtMobiles        string `json:"at_mobiles"`
		AtUserIds        string `json:"at_user_ids"`
		IsAtAll          bool   `json:"is_at_all"`
		Status           string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, "参数错误："+err.Error(), nil)
		return
	}

	if req.TextTemplate == "" && req.HTMLTemplate == "" && req.MarkdownTemplate == "" {
		appG.CResponse(http.StatusBadRequest, "至少需要填写一种格式的模板内容", nil)
		return
	}

	if req.Status == "" {
		req.Status = "enabled"
	}

	templateService := message_template_service.TemplateService{
		ID:               strings.TrimSpace(req.ID),
		Name:             req.Name,
		Description:      req.Description,
		TextTemplate:     req.TextTemplate,
		HTMLTemplate:     req.HTMLTemplate,
		MarkdownTemplate: req.MarkdownTemplate,
		Placeholders:     req.Placeholders,
		AtMobiles:        req.AtMobiles,
		AtUserIds:        req.AtUserIds,
		IsAtAll:          req.IsAtAll,
		Status:           req.Status,
	}

	if err := templateService.Add(); err != nil {
		appG.CResponse(http.StatusInternalServerError, "添加模板失败："+err.Error(), nil)
		return
	}

	appG.CResponse(http.StatusOK, "添加模板成功", nil)
}

// EditMessageTemplate 编辑消息模板
func EditMessageTemplate(c *gin.Context) {
	appG := app.Gin{C: c}

	var req struct {
		ID               string `json:"id" binding:"required"`
		Name             string `json:"name" binding:"required"`
		Description      string `json:"description"`
		TextTemplate     string `json:"text_template"`
		HTMLTemplate     string `json:"html_template"`
		MarkdownTemplate string `json:"markdown_template"`
		Placeholders     string `json:"placeholders"`
		AtMobiles        string `json:"at_mobiles"`
		AtUserIds        string `json:"at_user_ids"`
		IsAtAll          bool   `json:"is_at_all"`
		Status           string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, "参数错误："+err.Error(), nil)
		return
	}

	if req.TextTemplate == "" && req.HTMLTemplate == "" && req.MarkdownTemplate == "" {
		appG.CResponse(http.StatusBadRequest, "至少需要填写一种格式的模板内容", nil)
		return
	}

	templateService := message_template_service.TemplateService{
		ID: req.ID,
	}

	exists, err := templateService.ExistByID()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "查询模板失败！", nil)
		return
	}

	if !exists {
		appG.CResponse(http.StatusNotFound, "模板不存在！", nil)
		return
	}

	templateService.Name = req.Name
	templateService.Description = req.Description
	templateService.TextTemplate = req.TextTemplate
	templateService.HTMLTemplate = req.HTMLTemplate
	templateService.MarkdownTemplate = req.MarkdownTemplate
	templateService.Placeholders = req.Placeholders
	templateService.AtMobiles = req.AtMobiles
	templateService.AtUserIds = req.AtUserIds
	templateService.IsAtAll = req.IsAtAll
	templateService.Status = req.Status

	if err := templateService.Update(); err != nil {
		appG.CResponse(http.StatusInternalServerError, "更新模板失败："+err.Error(), nil)
		return
	}

	appG.CResponse(http.StatusOK, "更新模板成功", nil)
}

// DeleteMessageTemplate 删除消息模板
func DeleteMessageTemplate(c *gin.Context) {
	appG := app.Gin{C: c}

	var req struct {
		ID string `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, "参数错误："+err.Error(), nil)
		return
	}

	templateService := message_template_service.TemplateService{
		ID: req.ID,
	}

	exists, err := templateService.ExistByID()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "查询模板失败！", nil)
		return
	}

	if !exists {
		appG.CResponse(http.StatusNotFound, "模板不存在！", nil)
		return
	}

	if err := templateService.Delete(); err != nil {
		appG.CResponse(http.StatusInternalServerError, "删除模板失败："+err.Error(), nil)
		return
	}

	appG.CResponse(http.StatusOK, "删除模板成功", nil)
}

// PreviewMessageTemplate 预览消息模板
func PreviewMessageTemplate(c *gin.Context) {
	appG := app.Gin{C: c}

	var req struct {
		ID     string            `json:"id" binding:"required"`
		Params map[string]string `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, "参数错误："+err.Error(), nil)
		return
	}

	templateService := message_template_service.TemplateService{
		ID: req.ID,
	}

	exists, err := templateService.ExistByID()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "查询模板失败！", nil)
		return
	}

	if !exists {
		appG.CResponse(http.StatusNotFound, "模板不存在！", nil)
		return
	}

	preview, err := templateService.PreviewTemplate(req.Params)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "预览模板失败："+err.Error(), nil)
		return
	}

	appG.CResponse(http.StatusOK, "预览模板成功", preview)
}

// GetTemplateWithIns 获取模板及其关联的实例
func GetTemplateWithIns(c *gin.Context) {
	appG := app.Gin{C: c}
	id := c.Query("id")

	if id == "" {
		appG.CResponse(http.StatusBadRequest, "模板ID为空！", nil)
		return
	}

	// 获取模板信息
	template, err := models.GetTemplateByID(id)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "获取模板信息失败！", nil)
		return
	}

	// 获取关联的实例列表
	insList, err := models.GetTemplateInsList(id)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "获取实例列表失败！", nil)
		return
	}

	result := map[string]interface{}{
		"template": template,
		"ins_list": insList,
	}

	appG.CResponse(http.StatusOK, "获取模板信息成功", result)
}

func GetTemplateRelations(c *gin.Context) {
	appG := app.Gin{C: c}
	id := c.Query("id")

	if id == "" {
		appG.CResponse(http.StatusBadRequest, "模板ID为空！", nil)
		return
	}

	relations := make([]map[string]interface{}, 0)

	cronMsgs, err := models.GetCronMsgsByTemplateID(id)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取关联定时消息失败！", nil)
		return
	}
	for _, msg := range cronMsgs {
		relations = append(relations, map[string]interface{}{
			"type": "定时消息",
			"id":   msg.ID,
			"name": msg.Name,
		})
	}

	appG.CResponse(http.StatusOK, "获取模板关联信息成功", map[string]interface{}{
		"relations": relations,
	})
}

// TemplateInsReq 模板实例请求结构
type TemplateInsReq struct {
	ID          string `json:"id" validate:"required,len=12" label:"实例id"`
	TemplateID  string `json:"template_id" validate:"required" label:"模板id"`
	WayID       string `json:"way_id" validate:"required,len=12" label:"渠道id"`
	ContentType string `json:"content_type" validate:"required,max=100" label:"实例内容类型"`
	Config      string `json:"config" validate:"" label:"任务配置"`
	Extra       string `json:"extra" validate:"" label:"任务额外信息"`
	WayType     string `json:"way_type" validate:"required,max=100" label:"渠道类型"`
}

// AddTemplateIns 添加模板关联的实例
func AddTemplateIns(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  TemplateInsReq
	)

	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	sendTaskInsService := send_ins_service.SendTaskInsService{}
	err := sendTaskInsService.AddOne(models.SendTasksIns{
		UUIDModel:   models.UUIDModel{ID: req.ID},
		TemplateID:  req.TemplateID,
		WayID:       req.WayID,
		WayType:     req.WayType,
		ContentType: req.ContentType,
		Config:      req.Config,
		Extra:       req.Extra,
	})
	if err != "" {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("添加实例失败！错误原因：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "添加实例成功！", nil)
}

type TemplateInsDeleteReq struct {
	ID string `json:"id" validate:"required,len=12" label:"实例id"`
}

// DeleteTemplateIns 删除模板实例
func DeleteTemplateIns(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  TemplateInsDeleteReq
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	sendTaskInsService := send_ins_service.SendTaskInsService{
		ID: req.ID,
	}
	if err := sendTaskInsService.Delete(); err != nil {
		appG.CResponse(http.StatusBadRequest, "删除实例失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "删除实例成功！", nil)
}

type TemplateInsEnableReq struct {
	ID     string `json:"id" validate:"required,len=12" label:"实例id"`
	Enable int    `json:"enable" validate:"oneof=0 1" label:"是否启用"`
}

type TemplateInsConfigReq struct {
	ID      string `json:"id" validate:"required,len=12" label:"实例id"`
	Config  string `json:"config" validate:"required" label:"实例配置"`
	WayType string `json:"way_type" validate:"required,max=100" label:"渠道类型"`
}

// UpdateTemplateInsEnable 更新模板实例启用状态
func UpdateTemplateInsEnable(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  TemplateInsEnableReq
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	sendTaskInsService := send_ins_service.SendTaskInsService{
		ID: req.ID,
	}
	data := map[string]interface{}{
		"enable": req.Enable,
	}
	if err := sendTaskInsService.Update(data); err != nil {
		appG.CResponse(http.StatusBadRequest, "更新实例状态失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "更新实例状态成功！", nil)
}

// UpdateTemplateInsConfig 更新模板实例配置
func UpdateTemplateInsConfig(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  TemplateInsConfigReq
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	sendTaskInsService := send_ins_service.SendTaskInsService{
		ID: req.ID,
	}
	if errStr, _ := sendTaskInsService.ValidateDiffIns(models.SendTasksIns{
		WayType: req.WayType,
		Config:  req.Config,
	}); errStr != "" {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("更新实例配置失败！错误原因：%s", errStr), nil)
		return
	}

	data := map[string]interface{}{
		"config": req.Config,
	}
	if err := sendTaskInsService.Update(data); err != nil {
		appG.CResponse(http.StatusBadRequest, "更新实例配置失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "更新实例配置成功！", nil)
}
