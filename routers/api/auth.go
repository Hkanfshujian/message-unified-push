package api

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/e"
	"ops-message-unified-push/pkg/util"
	"ops-message-unified-push/service/auth_service"
	"ops-message-unified-push/service/settings_service"
)

type AttemptInfo struct {
	Count       int
	LastAttempt time.Time
	LockUntil   time.Time
}

var loginAttempts sync.Map

const (
	MaxFailures      = 5
	FailResetTime    = 10 * time.Minute
	LockDurationTime = 30 * time.Minute
)

func init() {
	// 启动一个后台 goroutine 每 30 分钟清理一次过期的记录，防止恶意攻击导致内存泄漏
	go func() {
		for {
			time.Sleep(30 * time.Minute)
			now := time.Now()
			loginAttempts.Range(func(key, value interface{}) bool {
				info := value.(*AttemptInfo)
				// 如果距离上次尝试已经超过了重置时间，且也没有在锁定中，就可以清理掉了
				if now.Sub(info.LastAttempt) > FailResetTime && now.After(info.LockUntil) {
					loginAttempts.Delete(key)
				}
				return true
			})
		}
	}()
}

type auth struct {
	Username string `valid:"Required; MaxSize(50)"`
	Password string `valid:"Required; MaxSize(50)"`
}

type ReqAuth struct {
	Username string `json:"username" validate:"required,max=50" label:"用户名"`
	Password string `json:"passwd" validate:"required,max=50" label:"密码"`
}

type ReqRegister struct {
	Username        string `json:"username" validate:"required,min=3,max=50" label:"用户名"`
	Password        string `json:"passwd" validate:"required,min=6,max=50" label:"密码"`
	ConfirmPassword string `json:"confirm_passwd" validate:"required,min=6,max=50" label:"确认密码"`
}

func GetAuth(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  ReqAuth
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	ip := c.ClientIP()
	now := time.Now()
	var info *AttemptInfo
	if val, ok := loginAttempts.Load(ip); ok {
		info = val.(*AttemptInfo)
		if now.Before(info.LockUntil) {
			appG.CResponse(http.StatusForbidden, fmt.Sprintf("登录失败次数过多，请于%d分钟后再试！", int(info.LockUntil.Sub(now).Minutes())+1), nil)
			return
		}
	}

	authService := auth_service.Auth{Username: req.Username, Password: req.Password}
	isExist, err := authService.Check()
	if err != nil {
		// OIDC 用户特殊处理
		if err.Error() == "OIDC_USER_USE_OIDC_LOGIN" {
			appG.CResponse(http.StatusForbidden, "OIDC_USER_USE_OIDC_LOGIN", nil)
			return
		}
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("校验失败：%s", err), nil)
		return
	}
	if !isExist {
		if info == nil {
			info = &AttemptInfo{}
		}
		// 如果距离上次失败超过设定时间，且没有处于锁定状态，则重置计数
		if now.Sub(info.LastAttempt) > FailResetTime && now.After(info.LockUntil) {
			info.Count = 0
		}
		info.Count++
		info.LastAttempt = now
		if info.Count >= MaxFailures {
			info.LockUntil = now.Add(LockDurationTime)
		}
		loginAttempts.Store(ip, info)

		appG.CResponse(http.StatusUnauthorized, "账号或密码不正确！", nil)
		return
	}

	// 成功登录则清除失败记录
	loginAttempts.Delete(ip)

	// 获取配置的 cookie 过期天数
	expDays := settings_service.GetCookieExpDays()
	token, err := util.GenerateToken(req.Username, req.Password, expDays)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("生成token失败：%s", err), nil)
		return
	}

	// 查询用户ID并记录登录日志
	if u, _ := models.GetUserByUsername(req.Username); u != nil {
		_ = models.AddLoginLog(u.ID, req.Username, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	appG.CResponse(http.StatusOK, "登录成功!", map[string]string{
		"token": token,
	})
}

func RegisterAuth(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  ReqRegister
	)
	if !isRegisterEnabled() {
		appG.CResponse(http.StatusForbidden, "当前环境已关闭注册，请联系管理员", nil)
		return
	}
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}
	if strings.TrimSpace(req.Password) != strings.TrimSpace(req.ConfirmPassword) {
		appG.CResponse(http.StatusBadRequest, "两次密码不一致", nil)
		return
	}
	exist, err := models.IsUsernameExists(strings.TrimSpace(req.Username), 0)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("校验用户名失败：%s", err.Error()), nil)
		return
	}
	if exist {
		appG.CResponse(http.StatusBadRequest, "用户名已存在", nil)
		return
	}
	if err = models.AddUser(strings.TrimSpace(req.Username), req.Password); err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("注册失败：%s", err.Error()), nil)
		return
	}

	// 绑定默认用户组
	if u, _ := models.GetUserByUsername(strings.TrimSpace(req.Username)); u != nil {
		bindDefaultGroupForLocalUser(u.ID)
	}

	expDays := settings_service.GetCookieExpDays()
	token, err := util.GenerateToken(strings.TrimSpace(req.Username), req.Password, expDays)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("注册成功但生成token失败：%s", err.Error()), nil)
		return
	}
	if u, _ := models.GetUserByUsername(strings.TrimSpace(req.Username)); u != nil {
		_ = models.AddLoginLog(u.ID, req.Username, c.ClientIP(), c.GetHeader("User-Agent"))
	}
	appG.CResponse(http.StatusOK, "注册成功", map[string]string{
		"token": token,
	})
}

func GetPublicAuthConfig(c *gin.Context) {
	appG := app.Gin{C: c}
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	appG.CResponse(http.StatusOK, "获取认证公开配置成功", map[string]interface{}{
		"casdoor_enabled":     boolToSettingString(isCasdoorEnabled()),
		"register_enabled":   boolToSettingString(isRegisterEnabled()),
		"casdoor_button_text": getCasdoorButtonText(),
		"casdoor_button_icon": getCasdoorButtonIcon(),
	})
}

func isCasdoorEnabled() bool {
	setting, _ := models.GetSettingByKey("auth_config", "casdoor_enabled")
	if setting.ID <= 0 {
		return false
	}
	val := strings.TrimSpace(setting.Value)
	return strings.EqualFold(val, "true") || val == "1"
}

func isRegisterEnabled() bool {
	setting, _ := models.GetSettingByKey("auth_config", "register_enabled")
	if setting.ID <= 0 {
		return false
	}
	val := strings.TrimSpace(setting.Value)
	return strings.EqualFold(val, "true") || val == "1"
}

func boolToSettingString(flag bool) string {
	if flag {
		return "true"
	}
	return "false"
}

func getCasdoorButtonText() string {
	setting, _ := models.GetSettingByKey("auth_config", "casdoor_button_text")
	if setting.ID <= 0 {
		return "企微登录"
	}
	val := strings.TrimSpace(setting.Value)
	if val == "" {
		return "企微登录"
	}
	return val
}

func getCasdoorButtonIcon() string {
	setting, _ := models.GetSettingByKey("auth_config", "casdoor_button_icon")
	if setting.ID <= 0 {
		return ""
	}
	return strings.TrimSpace(setting.Value)
}

func getAuthSettingValue(key string, defaultValue string) string {
	setting, _ := models.GetSettingByKey("auth_config", key)
	if setting.ID <= 0 {
		return defaultValue
	}
	val := strings.TrimSpace(setting.Value)
	if val == "" {
		return defaultValue
	}
	return val
}

// bindDefaultGroupForLocalUser 为本地新用户绑定默认用户组
func bindDefaultGroupForLocalUser(userID int) {
	groupCode := getAuthSettingValue("local_default_group_code", "")
	if groupCode == "" {
		return
	}
	group, err := models.GetUserGroupByCode(groupCode)
	if err != nil || group == nil {
		return
	}
	_ = models.AssignUserToGroupIfNotExists(userID, group.ID, "system")
}
