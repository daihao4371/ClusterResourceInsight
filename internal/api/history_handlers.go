package api

import (
	"strconv"

	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"
	"cluster-resource-insight/pkg/utils"

	"github.com/gin-gonic/gin"
)

// QueryHistoryData 查询历史数据 - 支持分页和多条件查询的历史数据检索
func QueryHistoryData(historyService *service.HistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用统一的分页处理器解析分页参数
		paginationHandler := utils.NewHttpPaginationHandler()
		paginationParams := paginationHandler.ParsePaginationParams(c, 20)

		var req service.HistoryQueryRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			response.BadRequest("请求参数格式错误: "+err.Error(), c)
			return
		}

		// 使用统一分页参数
		req.Page = paginationParams.Page
		req.Size = paginationParams.Size

		historyResponse, err := historyService.QueryHistory(req)
		if err != nil {
			logger.Error("查询历史数据失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(historyResponse, c)
	}
}

// GetTrendData 获取趋势数据 - 获取指定条件下的资源使用趋势图表数据
func GetTrendData(historyService *service.HistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterIDStr := c.Query("cluster_id")
		namespace := c.Query("namespace")
		podName := c.Query("pod_name")
		hoursStr := c.DefaultQuery("hours", "24")

		var clusterID uint
		if clusterIDStr != "" {
			id, err := strconv.ParseUint(clusterIDStr, 10, 32)
			if err != nil {
				response.BadRequest("集群ID格式错误", c)
				return
			}
			clusterID = uint(id)
		}

		hours, err := strconv.Atoi(hoursStr)
		if err != nil || hours <= 0 {
			hours = 24
		}

		data, err := historyService.GetTrendData(clusterID, namespace, podName, hours)
		if err != nil {
			logger.Error("获取趋势数据失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data":       data,
			"cluster_id": clusterID,
			"namespace":  namespace,
			"pod_name":   podName,
			"hours":      hours,
			"count":      len(data),
		}, c)
	}
}

// GetHistoryStatistics 获取历史数据统计 - 获取历史数据的汇总统计信息
func GetHistoryStatistics(historyService *service.HistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := historyService.GetStatistics()
		if err != nil {
			logger.Error("获取历史数据统计失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data": stats,
		}, c)
	}
}

// TriggerDataCollection 触发数据收集 - 手动触发多集群数据收集任务
func TriggerDataCollection(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		persistenceStr := c.DefaultQuery("persistence", "true")
		enablePersistence := persistenceStr == "true"

		result, err := multiCollector.CollectAllClustersDataWithPersistence(c.Request.Context(), enablePersistence)
		if err != nil {
			logger.Error("触发数据收集失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithDetailed(result, "数据收集完成", c)
	}
}

// CleanupOldData 清理过期数据 - 清理指定保留天数之前的历史数据
func CleanupOldData(historyService *service.HistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		retentionDaysStr := c.DefaultQuery("retention_days", "30")
		retentionDays, err := strconv.Atoi(retentionDaysStr)
		if err != nil || retentionDays <= 0 {
			retentionDays = 30
		}

		err = historyService.CleanupOldData(c.Request.Context(), retentionDays)
		if err != nil {
			logger.Error("清理过期数据失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"message":        "数据清理完成",
			"retention_days": retentionDays,
		}, c)
	}
}

// GetSystemTrendData 获取系统级趋势数据 - 为Dashboard提供聚合的趋势图表数据
func GetSystemTrendData(historyService *service.HistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		hoursStr := c.DefaultQuery("hours", "24")

		hours, err := strconv.Atoi(hoursStr)
		if err != nil || hours <= 0 {
			hours = 24 // 默认24小时
		}

		// 限制查询范围，避免性能问题
		if hours > 168 { // 最大7天
			hours = 168
		}

		data, err := historyService.GetSystemTrendData(hours)
		if err != nil {
			logger.Error("获取系统趋势数据失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		logger.Info("系统趋势数据获取完成: hours=%d, data_points=%d", hours, len(data))

		response.OkWithData(gin.H{
			"data":  data,
			"hours": hours,
			"count": len(data),
		}, c)
	}
}
