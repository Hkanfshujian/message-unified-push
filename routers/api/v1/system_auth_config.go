package v1

import (
	"fmt"
	"net/http"

	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/constant"
	"ops-message-unified-push/pkg/e"
	"ops-message-unified-push/service/settings_service"

	"github.com/gin-gonic/gin"
)

type UpdateSystemAuthConfigReq struct {
	Data map[string]interface{} `json:"data"`
}

func GetSystemAuthConfig(c *gin.Context) {
	appG := app.Gin{C: c}
	settingService := settings_service.UserSettings{}
	settings, err := settingService.GetUserSetting(constant.AuthConfigSectionName)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取认证配置失败：%s", err.Error()), nil)
		return
	}
	// 用默认值填充缺失的字段
	for key, value := range constant.AuthConfigDefaultValueMap {
		if _, exists := settings[key]; !exists {
			settings[key] = value
		}
	}
	appG.CResponse(http.StatusOK, "获取认证配置成功", settings)
}

func UpdateSystemAuthConfig(c *gin.Context) {
	var req UpdateSystemAuthConfigReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	settingService := settings_service.UserSettings{}
	currentSettings, err := settingService.GetUserSetting(constant.AuthConfigSectionName)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("读取当前认证配置失败：%s", err.Error()), nil)
		return
	}
	// 用默认值填充缺失的字段，确保验证通过
	for key, value := range constant.AuthConfigDefaultValueMap {
		if _, exists := currentSettings[key]; !exists {
			currentSettings[key] = value
		}
	}
	for key, value := range req.Data {
		// 将值转换为字符串
		var strValue string
		switch v := value.(type) {
		case bool:
			if v {
				strValue = "true"
			} else {
				strValue = "false"
			}
		case string:
			strValue = v
		default:
			strValue = fmt.Sprintf("%v", v)
		}
		currentSettings[key] = strValue
	}
	diffStr := settingService.ValidateDiffSetting(constant.AuthConfigSectionName, currentSettings)
	if diffStr != "" {
		appG.CResponse(http.StatusBadRequest, diffStr, nil)
		return
	}
	currentUser := app.GetCurrentUserName(c)
	for key, value := range req.Data {
		// 将值转换为字符串
		var strValue string
		switch v := value.(type) {
		case bool:
			if v {
				strValue = "true"
			} else {
				strValue = "false"
			}
		case string:
			strValue = v
		default:
			strValue = fmt.Sprintf("%v", v)
		}
		if err := settingService.EditSettings(constant.AuthConfigSectionName, key, strValue, currentUser); err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("保存认证配置失败：%s", err.Error()), nil)
			return
		}
	}
	appG.CResponse(http.StatusOK, "保存认证配置成功", nil)
}
