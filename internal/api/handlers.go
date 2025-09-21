package api

import (
	"fmt"
	"strconv"
	"time"

	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"
	"cluster-resource-insight/pkg/statistics"
	"cluster-resource-insight/pkg/utils"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 已移动到 internal/router 包中
// 这个文件现在只包含处理器函数

// 创建集群
func CreateCluster(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.CreateClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest("请求参数格式错误: "+err.Error(), c)
			return
		}

		cluster, err := clusterService.CreateCluster(&req)
		if err != nil {
			logger.Error("创建集群失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithDetailed(cluster, "集群创建成功", c)
	}
}

// 获取所有集群
func GetAllClusters(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusters, err := clusterService.GetAllClusters()
		if err != nil {
			logger.Error("获取集群列表失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data":  clusters,
			"count": len(clusters),
		}, c)
	}
}

// GetClusterByID 根据ID获取集群
func GetClusterByID(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的集群ID", c)
			return
		}

		cluster, err := clusterService.GetClusterByID(uint(id))
		if err != nil {
			response.NotFound(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data": cluster,
		}, c)
	}
}

// 更新集群配置
func UpdateCluster(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的集群ID", c)
			return
		}

		var req service.UpdateClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest("请求参数格式错误: "+err.Error(), c)
			return
		}

		cluster, err := clusterService.UpdateCluster(uint(id), &req)
		if err != nil {
			logger.Error("更新集群失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithDetailed(cluster, "集群更新成功", c)
	}
}

// 删除集群
func DeleteCluster(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的集群ID", c)
			return
		}

		err = clusterService.DeleteCluster(uint(id))
		if err != nil {
			logger.Error("删除集群失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("集群删除成功", c)
	}
}

// 测试集群连接
func TestClusterConnection(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的集群ID", c)
			return
		}

		result, err := clusterService.TestClusterConnection(uint(id))
		if err != nil {
			logger.Error("测试集群连接失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data": result,
		}, c)
	}
}

// 测试集群配置（创建前验证）
func TestClusterByConfig(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.CreateClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest("请求参数格式错误: "+err.Error(), c)
			return
		}

		result, err := clusterService.TestClusterConnectionByConfig(&req)
		if err != nil {
			logger.Error("测试集群配置失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data": result,
		}, c)
	}
}

// 批量测试所有集群
func BatchTestAllClusters(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		results, err := clusterService.BatchTestAllClusters()
		if err != nil {
			logger.Error("批量测试集群失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data":  results,
			"count": len(results),
		}, c)
	}
}

func GetResourceAnalysis(resourceCollector *collector.ResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Starting resource analysis...")

		// 使用统一的分页处理器解析分页参数
		paginationHandler := utils.NewHttpPaginationHandler()
		paginationParams := paginationHandler.ParsePaginationParams(c, 50)

		clusterName := c.Query("cluster_name")

		// 使用多集群收集器获取数据
		multiCollector := collector.NewMultiClusterResourceCollector()
		result, err := multiCollector.CollectAllClustersData(c.Request.Context())
		if err != nil {
			logger.Error("Error in resource analysis: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		// 应用集群筛选
		var filteredProblems []collector.PodResourceInfo
		if clusterName != "" {
			for _, pod := range result.Top50Problems {
				if pod.ClusterName == clusterName {
					filteredProblems = append(filteredProblems, pod)
				}
			}
		} else {
			filteredProblems = result.Top50Problems
		}

		// 应用分页
		pagedProblems, paginationResult := paginationHandler.ApplyPaginationToSlice(filteredProblems, paginationParams)

		// 构造分页响应
		pagedResult := &collector.AnalysisResult{
			TotalPods:        result.TotalPods,
			UnreasonablePods: result.UnreasonablePods,
			Top50Problems:    pagedProblems,
			GeneratedAt:      result.GeneratedAt,
			ClustersAnalyzed: result.ClustersAnalyzed,
		}

		logger.Info("Analysis complete: total_pods=%d, unreasonable_pods=%d, page=%d, size=%d, cluster=%s",
			result.TotalPods, result.UnreasonablePods, paginationParams.Page, paginationParams.Size, clusterName)

		// 使用统一的分页响应构建器
		responseData := paginationHandler.BuildPaginationResponse(paginationParams, paginationResult.Total, pagedResult)
		responseData["filter"] = gin.H{
			"cluster_name": clusterName,
		}

		response.OkWithData(responseData, c)
	}
}

func GetPodsData(resourceCollector *collector.ResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取查询参数
		limitStr := c.DefaultQuery("limit", "50")
		onlyProblems := c.DefaultQuery("only_problems", "true")

		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 50
		}

		result, err := resourceCollector.CollectAllPodsData(c.Request.Context())
		if err != nil {
			response.InternalServerError(err.Error(), c)
			return
		}

		// 根据参数返回不同的数据
		if onlyProblems == "true" {
			pods := result.Top50Problems
			if len(pods) > limit {
				pods = pods[:limit]
			}
			response.OkWithData(gin.H{
				"pods":        pods,
				"total_count": len(result.Top50Problems),
				"limit":       limit,
			}, c)
		} else {
			// 返回所有数据，这里需要重新收集
			response.OkWithData(gin.H{
				"total_pods":        result.TotalPods,
				"unreasonable_pods": result.UnreasonablePods,
				"problems":          result.Top50Problems[:utils.MinInt(limit, len(result.Top50Problems))],
			}, c)
		}
	}
}

func HealthCheck(c *gin.Context) {
	response.OkWithData(gin.H{
		"status":  "healthy",
		"service": "cluster-resource-insight",
	}, c)
}

// 系统统计接口 - 提供Dashboard页面所需的统计数据
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

// 资源统计相关handlers
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

func GetNamespacesSummary(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		summaries, err := multiCollector.GetNamespacesSummary(c.Request.Context())
		if err != nil {
			logger.Error("获取命名空间汇总失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data":  summaries,
			"count": len(summaries),
		}, c)
	}
}

// 命名空间相关handlers
func GetAllNamespaces(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		summaries, err := multiCollector.GetNamespacesSummary(c.Request.Context())
		if err != nil {
			logger.Error("获取命名空间列表失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		// 提取命名空间名称列表
		namespaces := make([]string, 0)
		namespaceMap := make(map[string]bool)

		for _, summary := range summaries {
			if !namespaceMap[summary.NamespaceName] {
				namespaces = append(namespaces, summary.NamespaceName)
				namespaceMap[summary.NamespaceName] = true
			}
		}

		response.OkWithData(gin.H{
			"data":  namespaces,
			"count": len(namespaces),
		}, c)
	}
}

func GetNamespacePods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		if namespace == "" {
			response.BadRequest("命名空间参数不能为空", c)
			return
		}

		pods, err := multiCollector.GetNamespacePods(c.Request.Context(), namespace)
		if err != nil {
			logger.Error("获取命名空间Pod失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data":      pods,
			"count":     len(pods),
			"namespace": namespace,
		}, c)
	}
}

func GetNamespaceTreeData(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		if namespace == "" {
			response.BadRequest("命名空间参数不能为空", c)
			return
		}

		treeData, err := multiCollector.GetNamespaceTreeData(c.Request.Context(), namespace)
		if err != nil {
			logger.Error("获取命名空间树状数据失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data": treeData,
		}, c)
	}
}

// Pod搜索与分页handlers
func SearchPods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req collector.PodSearchRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			response.BadRequest("请求参数格式错误: "+err.Error(), c)
			return
		}

		// 设置默认值
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.Size <= 0 {
			req.Size = 10
		}

		podsResponse, err := multiCollector.SearchPods(c.Request.Context(), req)
		if err != nil {
			logger.Error("搜索Pod失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(podsResponse, c)
	}
}

func ListPods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用统一的分页处理器解析分页参数
		paginationHandler := utils.NewHttpPaginationHandler()
		paginationParams := paginationHandler.ParsePaginationParams(c, 10)

		req := collector.PodSearchRequest{
			Page: paginationParams.Page,
			Size: paginationParams.Size,
		}

		podsResponse, err := multiCollector.SearchPods(c.Request.Context(), req)
		if err != nil {
			logger.Error("获取Pod列表失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(podsResponse, c)
	}
}

// 历史数据相关handlers
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

// 调度服务管理handlers
func GetScheduleStatus(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := scheduleService.GetStatus()
		response.OkWithData(gin.H{
			"data": status,
		}, c)
	}
}

func StartScheduleService(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := scheduleService.Start(c.Request.Context())
		if err != nil {
			logger.Error("启动调度服务失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("调度服务启动成功", c)
	}
}

func StopScheduleService(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := scheduleService.Stop()
		if err != nil {
			logger.Error("停止调度服务失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("调度服务停止成功", c)
	}
}

func GetScheduleJobs(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobs := scheduleService.GetAllJobs()
		response.OkWithData(gin.H{
			"data":  jobs,
			"count": len(jobs),
		}, c)
	}
}

func RestartClusterJob(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterIDStr := c.Param("cluster_id")
		clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的集群ID", c)
			return
		}

		err = scheduleService.RestartJob(uint(clusterID))
		if err != nil {
			logger.Error("重启集群任务失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("集群任务重启成功", c)
	}
}

func UpdateScheduleSettings(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings service.GlobalScheduleSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			response.BadRequest("请求参数格式错误: "+err.Error(), c)
			return
		}

		err := scheduleService.UpdateSettings(&settings)
		if err != nil {
			logger.Error("更新调度设置失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithDetailed(settings, "调度设置更新成功", c)
	}
}

// 问题Pod分页接口
func GetProblemsWithPagination(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用统一的分页处理器解析分页参数
		paginationHandler := utils.NewHttpPaginationHandler()
		paginationParams := paginationHandler.ParsePaginationParams(c, 10)

		clusterName := c.Query("cluster_name")
		sortBy := c.DefaultQuery("sort_by", "total_waste")

		// 获取所有问题Pod数据
		result, err := multiCollector.CollectAllClustersData(c.Request.Context())
		if err != nil {
			logger.Error("获取问题Pod数据失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		// 应用集群筛选
		var filteredProblems []collector.PodResourceInfo
		if clusterName != "" {
			for _, pod := range result.Top50Problems {
				if pod.ClusterName == clusterName {
					filteredProblems = append(filteredProblems, pod)
				}
			}
		} else {
			filteredProblems = result.Top50Problems
		}

		// 应用排序
		podSorter := utils.NewPodSorter()
		podSorter.SortProblems(filteredProblems, sortBy)

		// 使用统一的分页处理器应用分页
		pagedProblems, paginationResult := paginationHandler.ApplyPaginationToSlice(filteredProblems, paginationParams)

		logger.Info("问题Pod分页查询完成: total=%d, page=%d, size=%d, cluster=%s, sort=%s",
			paginationResult.Total, paginationParams.Page, paginationParams.Size, clusterName, sortBy)

		// 使用统一的分页响应构建器
		responseData := paginationHandler.BuildPaginationResponse(paginationParams, paginationResult.Total, pagedProblems)
		responseData["cluster_name"] = clusterName
		responseData["sort_by"] = sortBy

		response.OkWithData(responseData, c)
	}
}

// getSystemTrendData 获取系统级趋势数据 - 为Dashboard提供聚合的趋势图表数据
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

// 活动和告警相关handlers
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

// updateAlertStatus 更新告警状态
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

// resolveAlert 解决告警
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

// dismissAlert 忽略告警
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

// getAlertDetails 获取告警详情
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

// generateSampleActivities 生成示例活动数据
func GenerateSampleActivities(activityService *service.ActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := activityService.GenerateRealtimeActivities()
		if err != nil {
			logger.Error("生成示例活动数据失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("示例活动数据生成成功", c)
	}
}

// cleanupOldActivitiesAndAlerts 清理旧的活动和告警数据
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
