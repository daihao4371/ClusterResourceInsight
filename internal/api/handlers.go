package api

import (
	"cluster-resource-insight/internal/models"
	"math"
	"sort"
	"strconv"
	"time"

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

	// 新增系统统计接口
	r.GET("/stats", getSystemStats(resourceCollector))

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

	// Pod搜索与分页接口
	podsGroup := r.Group("/pods")
	{
		podsGroup.GET("/search", searchPods(multiCollector))
		podsGroup.GET("/list", listPods(multiCollector))
		podsGroup.GET("/problems", getProblemsWithPagination(multiCollector)) // 新增问题Pod分页接口
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
		scheduleGroup.GET("/status", getScheduleStatus(scheduleService))                    // 获取调度服务状态
		scheduleGroup.POST("/start", startScheduleService(scheduleService))                 // 启动调度服务
		scheduleGroup.POST("/stop", stopScheduleService(scheduleService))                   // 停止调度服务
		scheduleGroup.GET("/jobs", getScheduleJobs(scheduleService))                        // 获取所有调度任务状态
		scheduleGroup.POST("/jobs/:cluster_id/restart", restartClusterJob(scheduleService)) // 重启指定集群任务
		scheduleGroup.PUT("/settings", updateScheduleSettings(scheduleService))             // 更新全局调度配置
	}

	// 集群管理接口
	clusterService := service.NewClusterService()

	// 集群管理路由组
	clusterGroup := r.Group("/clusters")
	{
		clusterGroup.POST("", createCluster(clusterService))                   // 创建集群
		clusterGroup.GET("", getAllClusters(clusterService))                   // 获取所有集群
		clusterGroup.GET("/:id", getClusterByID(clusterService))               // 根据ID获取集群
		clusterGroup.PUT("/:id", updateCluster(clusterService))                // 更新集群配置
		clusterGroup.DELETE("/:id", deleteCluster(clusterService))             // 删除集群
		clusterGroup.POST("/:id/test", testClusterConnection(clusterService))  // 测试集群连接
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

		// 获取分页参数
		pageStr := c.DefaultQuery("page", "1")
		sizeStr := c.DefaultQuery("size", "50")
		clusterName := c.Query("cluster_name")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			page = 1
		}

		size, err := strconv.Atoi(sizeStr)
		if err != nil || size <= 0 {
			size = 50
		}

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
		total := len(filteredProblems)
		start := (page - 1) * size
		end := start + size

		if start > total {
			start = total
		}
		if end > total {
			end = total
		}

		var pagedProblems []collector.PodResourceInfo
		if start < end {
			pagedProblems = filteredProblems[start:end]
		}

		// 构造分页响应
		pagedResult := &collector.AnalysisResult{
			TotalPods:        result.TotalPods,
			UnreasonablePods: result.UnreasonablePods,
			Top50Problems:    pagedProblems,
			GeneratedAt:      result.GeneratedAt,
			ClustersAnalyzed: result.ClustersAnalyzed,
		}

		logger.Info("Analysis complete: total_pods=%d, unreasonable_pods=%d, page=%d, size=%d, cluster=%s",
			result.TotalPods, result.UnreasonablePods, page, size, clusterName)

		response.OkWithData(gin.H{
			"data": pagedResult,
			"pagination": gin.H{
				"page":        page,
				"size":        size,
				"total":       total,
				"total_pages": (total + size - 1) / size,
				"has_next":    end < total,
				"has_prev":    page > 1,
			},
			"filter": gin.H{
				"cluster_name": clusterName,
			},
		}, c)
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
		"status":  "healthy",
		"service": "cluster-resource-insight",
	}, c)
}

// 系统统计接口 - 提供Dashboard页面所需的统计数据
func getSystemStats(resourceCollector *collector.ResourceCollector) gin.HandlerFunc {
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

		// 统计在线集群数量
		onlineClusters := 0
		for _, cluster := range clusters {
			if cluster.Status == "online" {
				onlineClusters++
			}
		}

		// 计算资源效率 - 基于所有Pod的实际使用率
		resourceEfficiency := calculateResourceEfficiency(analysisResult)

		// 计算集群状态分布
		clusterStatusDistribution := calculateClusterStatusDistribution(clusters)

		// 构建系统统计响应
		stats := gin.H{
			"total_clusters":              len(clusters),
			"online_clusters":             onlineClusters,
			"total_pods":                  analysisResult.TotalPods,
			"problem_pods":                analysisResult.UnreasonablePods,
			"resource_efficiency":         resourceEfficiency,
			"cluster_status_distribution": clusterStatusDistribution,
			"last_update":                 time.Now().Format(time.RFC3339),
		}

		logger.Info("系统统计数据获取完成: clusters=%d, online=%d, pods=%d, problems=%d, efficiency=%.1f%%",
			len(clusters), onlineClusters, analysisResult.TotalPods, analysisResult.UnreasonablePods, resourceEfficiency)

		response.OkWithData(stats, c)
	}
}

// 计算综合资源效率
func calculateResourceEfficiency(analysisResult *collector.AnalysisResult) float64 {
	if analysisResult == nil || analysisResult.TotalPods == 0 {
		return 0.0
	}

	// 获取所有Pod数据（包括问题Pod和正常Pod）
	allPods := analysisResult.Top50Problems

	// 如果没有足够的数据，返回基于问题Pod比例的简单计算
	if len(allPods) == 0 {
		if analysisResult.TotalPods == 0 {
			return 0.0
		}
		return float64(analysisResult.TotalPods-analysisResult.UnreasonablePods) / float64(analysisResult.TotalPods) * 100
	}

	var totalCPUEfficiency, totalMemoryEfficiency float64
	var validPods int

	// 计算所有Pod的平均资源使用效率
	for _, pod := range allPods {
		// CPU效率：实际使用率（如果有请求的话）
		if pod.CPURequest > 0 && pod.CPUReqPct > 0 {
			// CPUReqPct已经是使用率百分比，限制在合理范围内
			cpuEfficiency := math.Min(pod.CPUReqPct, 100.0)
			totalCPUEfficiency += cpuEfficiency
		}

		// 内存效率：实际使用率（如果有请求的话）
		if pod.MemoryRequest > 0 && pod.MemoryReqPct > 0 {
			// MemoryReqPct已经是使用率百分比，限制在合理范围内
			memoryEfficiency := math.Min(pod.MemoryReqPct, 100.0)
			totalMemoryEfficiency += memoryEfficiency
		}

		validPods++
	}

	// 如果没有有效的Pod数据，使用简单算法
	if validPods == 0 {
		if analysisResult.TotalPods == 0 {
			return 0.0
		}
		return float64(analysisResult.TotalPods-analysisResult.UnreasonablePods) / float64(analysisResult.TotalPods) * 100
	}

	// 计算平均效率
	avgCPUEfficiency := totalCPUEfficiency / float64(validPods)
	avgMemoryEfficiency := totalMemoryEfficiency / float64(validPods)

	// 综合资源效率 = (CPU效率 + 内存效率) / 2
	overallEfficiency := (avgCPUEfficiency + avgMemoryEfficiency) / 2.0

	// 确保返回值在合理范围内
	return math.Max(0.0, math.Min(100.0, overallEfficiency))
}

// 计算集群状态分布 - 为前端图表提供数据
func calculateClusterStatusDistribution(clusters []models.ClusterConfig) []gin.H {
	// 统计各种状态的集群数量
	statusCount := make(map[string]int)

	for _, cluster := range clusters {
		statusCount[cluster.Status]++
	}

	// 构建前端期望的数据格式
	var distribution []gin.H

	// 在线集群
	if count, exists := statusCount["online"]; exists {
		distribution = append(distribution, gin.H{
			"name":  "在线",
			"value": count,
			"color": "#22c55e", // 绿色
		})
	}

	// 离线集群
	if count, exists := statusCount["offline"]; exists {
		distribution = append(distribution, gin.H{
			"name":  "离线",
			"value": count,
			"color": "#ef4444", // 红色
		})
	}

	// 错误状态集群
	if count, exists := statusCount["error"]; exists {
		distribution = append(distribution, gin.H{
			"name":  "错误",
			"value": count,
			"color": "#f59e0b", // 橙色
		})
	}

	// 未知状态集群
	if count, exists := statusCount["unknown"]; exists {
		distribution = append(distribution, gin.H{
			"name":  "未知",
			"value": count,
			"color": "#6b7280", // 灰色
		})
	}

	// 如果没有任何状态数据，返回空分布
	if len(distribution) == 0 {
		distribution = append(distribution, gin.H{
			"name":  "在线",
			"value": 0,
			"color": "#22c55e",
		})
	}

	return distribution
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

// 问题Pod分页接口
func getProblemsWithPagination(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取分页参数
		pageStr := c.DefaultQuery("page", "1")
		sizeStr := c.DefaultQuery("size", "10")
		clusterName := c.Query("cluster_name")
		sortBy := c.DefaultQuery("sort_by", "total_waste")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			page = 1
		}

		size, err := strconv.Atoi(sizeStr)
		if err != nil || size <= 0 || size > 100 { // 限制最大每页数量
			size = 10
		}

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
		sortProblems(filteredProblems, sortBy)

		// 应用分页
		total := len(filteredProblems)
		totalPages := (total + size - 1) / size
		start := (page - 1) * size
		end := start + size

		if start > total {
			start = total
		}
		if end > total {
			end = total
		}

		var pagedProblems []collector.PodResourceInfo
		if start < end {
			pagedProblems = filteredProblems[start:end]
		}

		logger.Info("问题Pod分页查询完成: total=%d, page=%d, size=%d, cluster=%s, sort=%s",
			total, page, size, clusterName, sortBy)

		response.OkWithData(gin.H{
			"pods":         pagedProblems,
			"total":        total,
			"page":         page,
			"size":         size,
			"total_pages":  totalPages,
			"has_next":     end < total,
			"has_prev":     page > 1,
			"cluster_name": clusterName,
			"sort_by":      sortBy,
		}, c)
	}
}

// 对问题Pod进行排序
func sortProblems(problems []collector.PodResourceInfo, sortBy string) {
	switch sortBy {
	case "cpu_waste":
		sort.Slice(problems, func(i, j int) bool {
			cpuWasteI := calculateCPUWaste(problems[i])
			cpuWasteJ := calculateCPUWaste(problems[j])
			return cpuWasteI > cpuWasteJ
		})
	case "memory_waste":
		sort.Slice(problems, func(i, j int) bool {
			memoryWasteI := calculateMemoryWaste(problems[i])
			memoryWasteJ := calculateMemoryWaste(problems[j])
			return memoryWasteI > memoryWasteJ
		})
	case "total_waste":
		fallthrough
	default:
		sort.Slice(problems, func(i, j int) bool {
			totalWasteI := calculateTotalWaste(problems[i])
			totalWasteJ := calculateTotalWaste(problems[j])
			return totalWasteI > totalWasteJ
		})
	}
}

// 计算CPU浪费程度
func calculateCPUWaste(pod collector.PodResourceInfo) float64 {
	if pod.CPUReqPct <= 0 {
		return 0
	}
	return math.Max(0, 100-pod.CPUReqPct)
}

// 计算内存浪费程度
func calculateMemoryWaste(pod collector.PodResourceInfo) float64 {
	if pod.MemoryReqPct <= 0 {
		return 0
	}
	return math.Max(0, 100-pod.MemoryReqPct)
}

// 计算总浪费程度
func calculateTotalWaste(pod collector.PodResourceInfo) float64 {
	cpuWaste := calculateCPUWaste(pod)
	memoryWaste := calculateMemoryWaste(pod)
	return (cpuWaste + memoryWaste) / 2
}
