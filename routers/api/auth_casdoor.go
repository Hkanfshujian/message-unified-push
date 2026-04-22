package api

import (
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/setting"
	"ops-message-unified-push/service/casdoor_service"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CasdoorLogin 发起 Casdoor 登录
func CasdoorLogin(c *gin.Context) {
	// 创建服务
	service, err := casdoor_service.NewCasdoorService()
	if err != nil {
		logrus.Errorf("[Casdoor] 初始化失败: %v", err)
		c.Redirect(http.StatusFound, buildCasdoorErrorRedirect(err.Error()))
		return
	}

	// 生成 state 和 nonce
	stateCache := casdoor_service.GetStateCache()
	state, nonce, err := stateCache.GenerateState(c.ClientIP())
	if err != nil {
		logrus.Errorf("[Casdoor] 生成 state 失败: %v", err)
		c.Redirect(http.StatusFound, buildCasdoorErrorRedirect("生成登录状态失败"))
		return
	}

	// 构建授权 URL 并重定向
	authURL := service.BuildAuthURL(state, nonce)
	logrus.Infof("[Casdoor] 重定向到授权页: %s", authURL)
	c.Redirect(http.StatusFound, authURL)
}

// CasdoorCallback 处理 Casdoor 回调
func CasdoorCallback(c *gin.Context) {
	// 异常恢复
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("[Casdoor] Callback panic: %v", r)
			c.Redirect(http.StatusFound, buildCasdoorErrorRedirect("服务器内部错误"))
		}
	}()

	state := c.Query("state")
	code := c.Query("code")
	errMsg := c.Query("error")
	errDesc := c.Query("error_description")

	// 检查 Casdoor 返回的错误
	if errMsg != "" {
		logrus.Errorf("[Casdoor] 授权失败: %s - %s", errMsg, errDesc)
		c.Redirect(http.StatusFound, buildCasdoorErrorRedirect(errDesc))
		return
	}

	// 验证参数
	if state == "" || code == "" {
		logrus.Error("[Casdoor] 回调参数缺失")
		c.Redirect(http.StatusFound, buildCasdoorErrorRedirect("回调参数缺失"))
		return
	}

	// 验证并消费 state（防重放）
	stateCache := casdoor_service.GetStateCache()
	stateInfo, valid := stateCache.ConsumeState(state)
	if !valid {
		logrus.Error("[Casdoor] State 无效或已使用")
		c.Redirect(http.StatusFound, buildCasdoorErrorRedirect("登录状态无效或已过期，请重新登录"))
		return
	}
	logrus.Infof("[Casdoor] State 验证通过, ClientIP: %s", stateInfo.ClientIP)

	// 创建服务
	service, err := casdoor_service.NewCasdoorService()
	if err != nil {
		logrus.Errorf("[Casdoor] 初始化失败: %v", err)
		c.Redirect(http.StatusFound, buildCasdoorErrorRedirect(err.Error()))
		return
	}

	// 处理回调
	localToken, user, err := service.HandleCallback(c.Request.Context(), code)
	if err != nil {
		logrus.Errorf("[Casdoor] 回调处理失败: %v", err)
		c.Redirect(http.StatusFound, buildCasdoorErrorRedirect(err.Error()))
		return
	}

	// 记录登录日志
	_ = models.AddLoginLog(user.ID, user.Username, c.ClientIP(), c.GetHeader("User-Agent"))

	// 构建前端重定向 URL
	redirectURL := buildCasdoorSuccessRedirect(localToken)
	logrus.Infof("[Casdoor] 登录成功，重定向到: %s", redirectURL)
	c.Redirect(http.StatusFound, redirectURL)
}

// CasdoorLogout 发起 Casdoor 统一登出
func CasdoorLogout(c *gin.Context) {
	redirectURI := c.Query("redirect_uri")

	// 从 JWT 获取用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		// 兼容历史链路：若中间件尚未注入 userID，则尝试通过用户名回查
		if currentUser, ok := c.Get("currentUserName"); ok {
			if userName, ok := currentUser.(string); ok && strings.TrimSpace(userName) != "" {
				if user, err := models.GetUserByUsername(userName); err == nil && user != nil {
					userID = user.ID
					exists = true
				}
			}
		}
		if !exists {
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"msg":  "未登录",
				"data": gin.H{"logout_url": ""},
			})
			return
		}
	}

	// 创建服务
	service, err := casdoor_service.NewCasdoorService()
	if err != nil {
		logrus.Warnf("[Casdoor] 登出初始化失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "本地登出",
			"data": gin.H{"logout_url": ""},
		})
		return
	}

	// 根据 userID 类型进行转换
	var uid uint
	switch v := userID.(type) {
	case int:
		uid = uint(v)
	case uint:
		uid = v
	case float64:
		uid = uint(v)
	default:
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "用户ID类型错误",
			"data": gin.H{"logout_url": ""},
		})
		return
	}

	// 构建登出 URL
	logoutURL := service.BuildLogoutURL(uid, redirectURI)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取登出地址成功",
		"data": gin.H{"logout_url": logoutURL},
	})
}

// CasdoorLogoutCallback 登出回调
func CasdoorLogoutCallback(c *gin.Context) {
	pathPrefix := getURLPrefix()
	target := pathPrefix + "/#/login?casdoor_logout=1"
	c.Redirect(http.StatusFound, target)
}

// CasdoorStatus 获取 Casdoor 配置状态（公开接口）
func CasdoorStatus(c *gin.Context) {
	config, err := casdoor_service.LoadCasdoorConfig()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"enabled":     false,
				"button_text": "",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"enabled":     config.Enabled,
			"button_text": config.ButtonText,
		},
	})
}

// buildCasdoorErrorRedirect 构建错误重定向 URL
func buildCasdoorErrorRedirect(errMsg string) string {
	pathPrefix := getURLPrefix()
	params := url.Values{}
	params.Set("casdoor_error", "1")
	params.Set("casdoor_error_msg", errMsg)
	return pathPrefix + "/#/login?" + params.Encode()
}

// buildCasdoorSuccessRedirect 构建成功重定向 URL
func buildCasdoorSuccessRedirect(token string) string {
	pathPrefix := getURLPrefix()
	params := url.Values{}
	params.Set("casdoor_token", token)
	params.Set("casdoor_login", "1")
	return pathPrefix + "/#/login?" + params.Encode()
}

// getURLPrefix 获取 URL 前缀
func getURLPrefix() string {
	pathPrefix := strings.TrimSpace(setting.ServerSetting.UrlPrefix)
	if pathPrefix != "" && !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/" + pathPrefix
	}
	return pathPrefix
}

// GetCasdoorStateCount 获取当前 state 缓存数量（调试用）
func GetCasdoorStateCount(c *gin.Context) {
	count := casdoor_service.GetStateCache().Count()
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"count": count},
		"msg":  fmt.Sprintf("当前缓存 %d 个登录状态", count),
	})
}
