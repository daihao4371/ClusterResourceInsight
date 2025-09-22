package api

import (
	"fmt"
	"strconv"

	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"

	"github.com/gin-gonic/gin"
)

// GetRecentActivities 获取最近活动 - 获取系统最近的活动记录列表
func GetRecentActivities(activityService *service.ActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 10
		}

		activities, err := activityService.GetRecentActivities(limit)
		if err != nil {
			logger.Error("获取最近活动失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data":  activities,
			"count": len(activities),
			"limit": limit,
		}, c)
	}
}

// GetRecentAlerts 获取最近告警 - 获取系统最近的告警信息列表
func GetRecentAlerts(activityService *service.ActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 10
		}

		alerts, err := activityService.GetRecentAlerts(limit)
		if err != nil {
			logger.Error("获取最近告警失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data":  alerts,
			"count": len(alerts),
			"limit": limit,
		}, c)
	}
}

// UpdateAlertStatus 更新告警状态 - 修改指定告警的状态信息
func UpdateAlertStatus(activityService *service.ActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取告警ID
		alertIDStr := c.Param("id")
		alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的告警ID", c)
			return
		}

		// 解析请求体
		var req struct {
			Status string `json:"status" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest("请求参数错误: "+err.Error(), c)
			return
		}

		// 更新告警状态
		err = activityService.UpdateAlertStatus(uint(alertID), req.Status)
		if err != nil {
			logger.Error("更新告警状态失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("告警状态更新成功", c)
	}
}

// ResolveAlert 解决告警 - 将指定告警标记为已解决状态
func ResolveAlert(activityService *service.ActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取告警ID
		alertIDStr := c.Param("id")
		alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的告警ID", c)
			return
		}

		// 解决告警
		err = activityService.ResolveAlert(uint(alertID))
		if err != nil {
			logger.Error("解决告警失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("告警已标记为已解决", c)
	}
}

// DismissAlert 忽略告警 - 将指定告警标记为已忽略状态
func DismissAlert(activityService *service.ActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取告警ID
		alertIDStr := c.Param("id")
		alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的告警ID", c)
			return
		}

		// 忽略告警
		err = activityService.DismissAlert(uint(alertID))
		if err != nil {
			logger.Error("忽略告警失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("告警已忽略", c)
	}
}

// GetAlertDetails 获取告警详情 - 查询指定ID的告警详细信息
func GetAlertDetails(activityService *service.ActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取告警ID
		alertIDStr := c.Param("id")
		alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的告警ID", c)
			return
		}

		// 获取告警详情
		alert, err := activityService.GetAlertByID(uint(alertID))
		if err != nil {
			logger.Error("获取告警详情失败: %v", err)
			if err.Error() == fmt.Sprintf("未找到ID为 %d 的告警记录", alertID) {
				response.NotFound("告警记录不存在", c)
			} else {
				response.InternalServerError(err.Error(), c)
			}
			return
		}

		response.OkWithData(alert, c)
	}
}

// CleanupOldActivitiesAndAlerts 清理旧的活动和告警数据 - 根据保留天数清理过期的活动和告警记录
func CleanupOldActivitiesAndAlerts(activityService *service.ActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		retentionDaysStr := c.DefaultQuery("retention_days", "0") // 默认清理所有数据
		retentionDays, err := strconv.Atoi(retentionDaysStr)
		if err != nil {
			retentionDays = 0
		}

		err = activityService.CleanupOldActivities(c.Request.Context(), retentionDays)
		if err != nil {
			logger.Error("清理旧数据失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("旧数据清理完成", c)
	}
}
