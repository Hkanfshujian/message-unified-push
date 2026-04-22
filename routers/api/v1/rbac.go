package v1

import (
	"fmt"
	"net/http"

	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/service/rbac_service"

	"github.com/gin-gonic/gin"
)

func GetCurrentUserPermissions(c *gin.Context) {
	appG := app.Gin{C: c}
	currentUser := app.GetCurrentUserName(c)
	if currentUser == "" {
		appG.CResponse(http.StatusUnauthorized, "获取当前用户失败", nil)
		return
	}

	service := rbac_service.UserAuthzService{UserName: currentUser}
	data, err := service.GetCurrentUserPermissions()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取当前用户权限失败！错误原因：%s", err.Error()), nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取当前用户权限成功！", data)
}
