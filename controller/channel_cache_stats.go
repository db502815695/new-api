package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetChannelCacheStats(c *gin.Context) {
	days := 7
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && (d == 1 || d == 7 || d == 30) {
			days = d
		}
	}

	results, err := model.GetChannelCacheStats(days)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "获取缓存统计失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}
