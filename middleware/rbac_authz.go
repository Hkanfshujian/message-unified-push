package middleware

import (
	"net/http"
	"strings"

	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/e"
	"ops-message-unified-push/service/rbac_service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	authzModeMonitor = "monitor"
	authzModeEnforce = "enforce"
	authzModeDisable = "disable"
)

func RequirePermission(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		checkPermission(c, []string{code}, false)
	}
}

func RequireAnyPermission(codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		checkPermission(c, codes, true)
	}
}

func checkPermission(c *gin.Context, requiredCodes []string, matchAny bool) {
	currentUser, _ := c.Get("currentUserName")
	userName, _ := currentUser.(string)
	if strings.TrimSpace(userName) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR_AUTH_NO_TOKEN,
			"msg":  e.GetMsg(e.ERROR_AUTH_NO_TOKEN),
			"data": nil,
		})
		c.Abort()
		return
	}

	mode := getRbacAuthzMode()
	if mode == authzModeDisable {
		c.Next()
		return
	}

	service := rbac_service.UserAuthzService{UserName: userName}
	var hasPermission bool
	var err error
	if matchAny {
		hasPermission, err = service.HasAnyPermission(requiredCodes)
	} else {
		hasPermission, err = service.HasPermission(requiredCodes[0])
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"prefix":   "[RBAC]",
			"user":     userName,
			"path":     c.FullPath(),
			"method":   c.Request.Method,
			"mode":     mode,
			"required": requiredCodes,
			"error":    err.Error(),
		}).Error("权限校验执行失败")
		c.JSON(http.StatusForbidden, gin.H{
			"code": e.ERROR_AUTH_FORBIDDEN,
			"msg":  e.GetMsg(e.ERROR_AUTH_FORBIDDEN),
			"data": nil,
		})
		c.Abort()
		return
	}

	if hasPermission {
		c.Next()
		return
	}

	logrus.WithFields(logrus.Fields{
		"prefix":   "[RBAC]",
		"user":     userName,
		"path":     c.FullPath(),
		"method":   c.Request.Method,
		"mode":     mode,
		"required": requiredCodes,
		"clientIP": c.ClientIP(),
	}).Warn("权限校验不通过")

	if mode == authzModeMonitor {
		c.Next()
		return
	}

	c.JSON(http.StatusForbidden, gin.H{
		"code": e.ERROR_AUTH_FORBIDDEN,
		"msg":  e.GetMsg(e.ERROR_AUTH_FORBIDDEN),
		"data": nil,
	})
	c.Abort()
}

func getRbacAuthzMode() string {
	setting, err := models.GetSettingByKey("rbac", "authz_mode")
	if err != nil {
		return authzModeMonitor
	}
	mode := strings.ToLower(strings.TrimSpace(setting.Value))
	switch mode {
	case authzModeEnforce, authzModeMonitor, authzModeDisable:
		return mode
	default:
		return authzModeMonitor
	}
}
