package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/constant"
	"ops-message-unified-push/pkg/e"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetProfileTheme(c *gin.Context) {
	appG := app.Gin{C: c}
	currentUser := app.GetCurrentUserName(c)
	if strings.TrimSpace(currentUser) == "" {
		appG.CResponse(http.StatusUnauthorized, "未登录", nil)
		return
	}
	user, err := models.GetUserByUsername(currentUser)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("查询用户失败：%s", err.Error()), nil)
		return
	}
	defaultTheme := "blue"
	defaultThemeSetting, _ := models.GetSettingByKey(constant.SiteSettingSectionName, "theme_color")
	if defaultThemeSetting.ID > 0 && strings.TrimSpace(defaultThemeSetting.Value) != "" {
		defaultTheme = strings.TrimSpace(defaultThemeSetting.Value)
	}
	resp := map[string]interface{}{
		"theme_color": defaultTheme,
		"theme_mode":  "system",
		"sidebar_bg":  "#0b3c51",
	}
	pref, err := models.GetUserPreferenceByUserID(user.ID)
	if err == nil && pref != nil {
		if strings.TrimSpace(pref.ThemeColor) != "" {
			resp["theme_color"] = strings.TrimSpace(pref.ThemeColor)
		}
		if strings.TrimSpace(pref.ThemeMode) != "" {
			resp["theme_mode"] = strings.TrimSpace(pref.ThemeMode)
		}
		if strings.TrimSpace(pref.SidebarBg) != "" {
			resp["sidebar_bg"] = strings.TrimSpace(pref.SidebarBg)
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("查询个人主题失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取个人主题成功", resp)
}

type UpdateProfileThemeReq struct {
	ThemeColor string `json:"theme_color" validate:"required,max=30" label:"主题色"`
	ThemeMode  string `json:"theme_mode" validate:"required,oneof=light dark system" label:"主题模式"`
	SidebarBg  string `json:"sidebar_bg" validate:"omitempty,max=30" label:"侧边栏背景色"`
}

func UpdateProfileTheme(c *gin.Context) {
	var req UpdateProfileThemeReq
	appG := app.Gin{C: c}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	currentUser := app.GetCurrentUserName(c)
	if strings.TrimSpace(currentUser) == "" {
		appG.CResponse(http.StatusUnauthorized, "未登录", nil)
		return
	}
	user, err := models.GetUserByUsername(currentUser)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("查询用户失败：%s", err.Error()), nil)
		return
	}
	sidebarBg := strings.TrimSpace(req.SidebarBg)
	if sidebarBg == "" {
		sidebarBg = "#0b3c51"
	}
	if err = models.UpsertUserPreference(user.ID, strings.TrimSpace(req.ThemeColor), strings.TrimSpace(req.ThemeMode), sidebarBg, currentUser); err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("保存个人主题失败：%s", err.Error()), nil)
		return
	}
	appG.CResponse(http.StatusOK, "保存个人主题成功", nil)
}
