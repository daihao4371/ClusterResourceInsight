package api

import (
	"strconv"

	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.RouterGroup, resourceCollector *collector.ResourceCollector) {
	// 原有的分析接口
	r.GET("/analysis", getResourceAnalysis(resourceCollector))
	r.GET("/pods", getPodsData(resourceCollector))
	r.GET("/health", healthCheck)

	// 新增的资源统计接口
	multiCollector := collector.NewMultiClusterResourceCollector()
	statisticsGroup := r.Group("/statistics")
	{
		statisticsGroup.GET("/top-memory-request", getTopMemoryRequestPods(multiCollector))
		statisticsGroup.GET("/top-cpu-request", getTopCPURequestPods(multiCollector))
		statisticsGroup.GET("/namespace-summary", getNamespacesSummary(multiCollector))
	}

	// 新增的命名空间相关接口
	namespacesGroup := r.Group("/namespaces")
	{
		namespacesGroup.GET("", getAllNamespaces(multiCollector))
		namespacesGroup.GET("/:namespace/pods", getNamespacePods(multiCollector))
		namespacesGroup.GET("/:namespace/tree-data", getNamespaceTreeData(multiCollector))
	}

	// 新增的Pod搜索与分页接口
	podsGroup := r.Group("/pods")
	{
		podsGroup.GET("/search", searchPods(multiCollector))
		podsGroup.GET("/list", listPods(multiCollector))
	}

	// 新增的历史数据接口
	historyService := service.NewHistoryService()
	historyGroup := r.Group("/history")
	{
		historyGroup.GET("/query", queryHistoryData(historyService))
		historyGroup.GET("/trends", getTrendData(historyService))
		historyGroup.GET("/statistics", getHistoryStatistics(historyService))
		historyGroup.POST("/collect", triggerDataCollection(multiCollector))
		historyGroup.DELETE("/cleanup", cleanupOldData(historyService))
	}

	// 新增的调度管理接口
	scheduleService := service.NewScheduleService()
	scheduleGroup := r.Group("/schedule")
	{
		scheduleGroup.GET("/status", getScheduleStatus(scheduleService))      // 获取调度服务状态
		scheduleGroup.POST("/start", startScheduleService(scheduleService))   // 启动调度服务
		scheduleGroup.POST("/stop", stopScheduleService(scheduleService))    // 停止调度服务
		scheduleGroup.GET("/jobs", getScheduleJobs(scheduleService))         // 获取所有调度任务状态
		scheduleGroup.POST("/jobs/:cluster_id/restart", restartClusterJob(scheduleService)) // 重启指定集群任务
		scheduleGroup.PUT("/settings", updateScheduleSettings(scheduleService)) // 更新全局调度配置
	}

	// 集群管理接口
	clusterService := service.NewClusterService()
	
	// 集群管理路由组
	clusterGroup := r.Group("/clusters")
	{
		clusterGroup.POST("", createCluster(clusterService))                    // 创建集群
		clusterGroup.GET("", getAllClusters(clusterService))                   // 获取所有集群
		clusterGroup.GET("/:id", getClusterByID(clusterService))               // 根据ID获取集群
		clusterGroup.PUT("/:id", updateCluster(clusterService))                // 更新集群配置
		clusterGroup.DELETE("/:id", deleteCluster(clusterService))             // 删除集群
		clusterGroup.POST("/:id/test", testClusterConnection(clusterService))   // 测试集群连接
		clusterGroup.POST("/test", testClusterByConfig(clusterService))        // 测试集群配置（创建前验证）
		clusterGroup.POST("/batch-test", batchTestAllClusters(clusterService)) // 批量测试所有集群
	}
}

// 创建集群
func createCluster(clusterService *service.ClusterService) gin.HandlerFunc {
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
func getAllClusters(clusterService *service.ClusterService) gin.HandlerFunc {
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

// 根据ID获取集群
func getClusterByID(clusterService *service.ClusterService) gin.HandlerFunc {
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
func updateCluster(clusterService *service.ClusterService) gin.HandlerFunc {
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
func deleteCluster(clusterService *service.ClusterService) gin.HandlerFunc {
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
func testClusterConnection(clusterService *service.ClusterService) gin.HandlerFunc {
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
func testClusterByConfig(clusterService *service.ClusterService) gin.HandlerFunc {
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
func batchTestAllClusters(clusterService *service.ClusterService) gin.HandlerFunc {
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

func getResourceAnalysis(resourceCollector *collector.ResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Starting resource analysis...")
		result, err := resourceCollector.CollectAllPodsData(c.Request.Context())
		if err != nil {
			logger.Error("Error in resource analysis: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		logger.Info("Analysis complete: total_pods=%d, unreasonable_pods=%d", result.TotalPods, result.UnreasonablePods)
		response.OkWithData(result, c)
	}
}

func getPodsData(resourceCollector *collector.ResourceCollector) gin.HandlerFunc {
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
				"problems":          result.Top50Problems[:min(limit, len(result.Top50Problems))],
			}, c)
		}
	}
}

func healthCheck(c *gin.Context) {
	response.OkWithData(gin.H{
		"status": "healthy",
		"service": "cluster-resource-insight",
	}, c)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 资源统计相关handlers
func getTopMemoryRequestPods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
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

func getTopCPURequestPods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
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

func getNamespacesSummary(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
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
func getAllNamespaces(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
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

func getNamespacePods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
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

func getNamespaceTreeData(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
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
func searchPods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
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

func listPods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		sizeStr := c.DefaultQuery("size", "10")
		
		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			page = 1
		}
		
		size, err := strconv.Atoi(sizeStr)
		if err != nil || size <= 0 {
			size = 10
		}
		
		req := collector.PodSearchRequest{
			Page: page,
			Size: size,
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
func queryHistoryData(historyService *service.HistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.HistoryQueryRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			response.BadRequest("请求参数格式错误: "+err.Error(), c)
			return
		}

		// 设置默认值
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.Size <= 0 {
			req.Size = 20
		}

		historyResponse, err := historyService.QueryHistory(req)
		if err != nil {
			logger.Error("查询历史数据失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(historyResponse, c)
	}
}

func getTrendData(historyService *service.HistoryService) gin.HandlerFunc {
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

func getHistoryStatistics(historyService *service.HistoryService) gin.HandlerFunc {
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

func triggerDataCollection(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
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

func cleanupOldData(historyService *service.HistoryService) gin.HandlerFunc {
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
func getScheduleStatus(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := scheduleService.GetStatus()
		response.OkWithData(gin.H{
			"data": status,
		}, c)
	}
}

func startScheduleService(scheduleService *service.ScheduleService) gin.HandlerFunc {
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

func stopScheduleService(scheduleService *service.ScheduleService) gin.HandlerFunc {
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

func getScheduleJobs(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobs := scheduleService.GetAllJobs()
		response.OkWithData(gin.H{
			"data":  jobs,
			"count": len(jobs),
		}, c)
	}
}

func restartClusterJob(scheduleService *service.ScheduleService) gin.HandlerFunc {
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

func updateScheduleSettings(scheduleService *service.ScheduleService) gin.HandlerFunc {
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