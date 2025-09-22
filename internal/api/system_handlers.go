package api

import (
	"strconv"
	"time"

	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"
	"cluster-resource-insight/pkg/statistics"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查 - 检查服务运行状态
func HealthCheck(c *gin.Context) {
	response.OkWithData(gin.H{
		"status":  "healthy",
		"service": "cluster-resource-insight",
	}, c)
}

// GetSystemStats 获取系统统计数据 - 提供Dashboard页面所需的统计数据
func GetSystemStats(resourceCollector *collector.ResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("获取系统统计数据...")

		// 使用多集群收集器获取完整的统计数据
		multiCollector := collector.NewMultiClusterResourceCollector()

		// 并行获取集群和分析数据
		clusterService := service.NewClusterService()
		clusters, err := clusterService.GetAllClusters()
		if err != nil {
			logger.Error("获取集群列表失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		// 获取资源分析数据
		analysisResult, err := multiCollector.CollectAllClustersData(c.Request.Context())
		if err != nil {
			logger.Error("获取资源分析数据失败: %v", err)
			// 如果分析数据获取失败，仍然返回基础统计信息
			analysisResult = &collector.AnalysisResult{
				TotalPods:        0,
				UnreasonablePods: 0,
				Top50Problems:    []collector.PodResourceInfo{},
				GeneratedAt:      time.Now(),
				ClustersAnalyzed: 0,
			}
		}

		// 使用统计构建器构建响应数据
		statsBuilder := statistics.NewSystemStatsBuilder()
		stats := statsBuilder.BuildSystemStats(clusters, analysisResult)

		logger.Info("系统统计数据获取完成: clusters=%d, online=%d, pods=%d, problems=%d",
			len(clusters), stats["online_clusters"], analysisResult.TotalPods, analysisResult.UnreasonablePods)

		response.OkWithData(stats, c)
	}
}

// GetTopMemoryRequestPods 获取Top内存请求Pod列表 - 按内存请求量排序的Pod统计
func GetTopMemoryRequestPods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "50")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 50
		}

		pods, err := multiCollector.GetTopMemoryRequestPods(c.Request.Context(), limit)
		if err != nil {
			logger.Error("获取Top内存请求Pod失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data":  pods,
			"count": len(pods),
			"limit": limit,
		}, c)
	}
}

// GetTopCPURequestPods 获取Top CPU请求Pod列表 - 按CPU请求量排序的Pod统计
func GetTopCPURequestPods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "50")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 50
		}

		pods, err := multiCollector.GetTopCPURequestPods(c.Request.Context(), limit)
		if err != nil {
			logger.Error("获取Top CPU请求Pod失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data":  pods,
			"count": len(pods),
			"limit": limit,
		}, c)
	}
}
