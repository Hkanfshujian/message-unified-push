package v1

import "github.com/gin-gonic/gin"

func pickDateRangeQuery(c *gin.Context) (string, string) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	if startTime == "" {
		startTime = c.Query("start_date")
	}
	if endTime == "" {
		endTime = c.Query("end_date")
	}
	return startTime, endTime
}
