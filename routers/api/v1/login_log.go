package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/e"
)

// GetRecentLoginLogs 登录日志列表（支持分页和时间范围过滤）
func GetRecentLoginLogs(c *gin.Context) {
	appG := app.Gin{C: c}

	// 获取时间范围参数
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 使用分页列表查询
	logs, total, err := models.GetLoginLogs(page, pageSize, startTime, endTime)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, e.GetMsg(e.ERROR), nil)
		return
	}
	appG.CResponse(http.StatusOK, "success", gin.H{
		"lists":     logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}


