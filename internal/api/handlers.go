package api

import (
	"log"
	"net/http"
	"strconv"

	"cluster-resource-insight/internal/collector"
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
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求参数格式错误: " + err.Error(),
			})
			return
		}

		cluster, err := clusterService.CreateCluster(&req)
		if err != nil {
			log.Printf("创建集群失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "集群创建成功",
			"data":    cluster,
		})
	}
}

// 获取所有集群
func getAllClusters(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusters, err := clusterService.GetAllClusters()
		if err != nil {
			log.Printf("获取集群列表失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  clusters,
			"count": len(clusters),
		})
	}
}

// 根据ID获取集群
func getClusterByID(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的集群ID",
			})
			return
		}

		cluster, err := clusterService.GetClusterByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": cluster,
		})
	}
}

// 更新集群配置
func updateCluster(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的集群ID",
			})
			return
		}

		var req service.UpdateClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求参数格式错误: " + err.Error(),
			})
			return
		}

		cluster, err := clusterService.UpdateCluster(uint(id), &req)
		if err != nil {
			log.Printf("更新集群失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "集群更新成功",
			"data":    cluster,
		})
	}
}

// 删除集群
func deleteCluster(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的集群ID",
			})
			return
		}

		err = clusterService.DeleteCluster(uint(id))
		if err != nil {
			log.Printf("删除集群失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "集群删除成功",
		})
	}
}

// 测试集群连接
func testClusterConnection(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的集群ID",
			})
			return
		}

		result, err := clusterService.TestClusterConnection(uint(id))
		if err != nil {
			log.Printf("测试集群连接失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": result,
		})
	}
}

// 测试集群配置（创建前验证）
func testClusterByConfig(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.CreateClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求参数格式错误: " + err.Error(),
			})
			return
		}

		result, err := clusterService.TestClusterConnectionByConfig(&req)
		if err != nil {
			log.Printf("测试集群配置失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": result,
		})
	}
}

// 批量测试所有集群
func batchTestAllClusters(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		results, err := clusterService.BatchTestAllClusters()
		if err != nil {
			log.Printf("批量测试集群失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  results,
			"count": len(results),
		})
	}
}

func getResourceAnalysis(resourceCollector *collector.ResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Starting resource analysis...")
		result, err := resourceCollector.CollectAllPodsData(c.Request.Context())
		if err != nil {
			log.Printf("Error in resource analysis: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		log.Printf("Analysis complete: total_pods=%d, unreasonable_pods=%d", result.TotalPods, result.UnreasonablePods)
		c.JSON(http.StatusOK, result)
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// 根据参数返回不同的数据
		if onlyProblems == "true" {
			pods := result.Top50Problems
			if len(pods) > limit {
				pods = pods[:limit]
			}
			c.JSON(http.StatusOK, gin.H{
				"pods":        pods,
				"total_count": len(result.Top50Problems),
				"limit":       limit,
			})
		} else {
			// 返回所有数据，这里需要重新收集
			c.JSON(http.StatusOK, gin.H{
				"total_pods":        result.TotalPods,
				"unreasonable_pods": result.UnreasonablePods,
				"problems":          result.Top50Problems[:min(limit, len(result.Top50Problems))],
			})
		}
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "cluster-resource-insight",
	})
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
			log.Printf("获取Top内存请求Pod失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"data":  pods,
			"count": len(pods),
			"limit": limit,
		})
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
			log.Printf("获取Top CPU请求Pod失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"data":  pods,
			"count": len(pods),
			"limit": limit,
		})
	}
}

func getNamespacesSummary(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		summaries, err := multiCollector.GetNamespacesSummary(c.Request.Context())
		if err != nil {
			log.Printf("获取命名空间汇总失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"data":  summaries,
			"count": len(summaries),
		})
	}
}

// 命名空间相关handlers
func getAllNamespaces(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		summaries, err := multiCollector.GetNamespacesSummary(c.Request.Context())
		if err != nil {
			log.Printf("获取命名空间列表失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
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
		
		c.JSON(http.StatusOK, gin.H{
			"data":  namespaces,
			"count": len(namespaces),
		})
	}
}

func getNamespacePods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		if namespace == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "命名空间参数不能为空",
			})
			return
		}
		
		pods, err := multiCollector.GetNamespacePods(c.Request.Context(), namespace)
		if err != nil {
			log.Printf("获取命名空间Pod失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"data":      pods,
			"count":     len(pods),
			"namespace": namespace,
		})
	}
}

func getNamespaceTreeData(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		if namespace == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "命名空间参数不能为空",
			})
			return
		}
		
		treeData, err := multiCollector.GetNamespaceTreeData(c.Request.Context(), namespace)
		if err != nil {
			log.Printf("获取命名空间树状数据失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"data": treeData,
		})
	}
}

// Pod搜索与分页handlers
func searchPods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req collector.PodSearchRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求参数格式错误: " + err.Error(),
			})
			return
		}
		
		// 设置默认值
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.Size <= 0 {
			req.Size = 10
		}
		
		response, err := multiCollector.SearchPods(c.Request.Context(), req)
		if err != nil {
			log.Printf("搜索Pod失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, response)
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
		
		response, err := multiCollector.SearchPods(c.Request.Context(), req)
		if err != nil {
			log.Printf("获取Pod列表失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, response)
	}
}

// 历史数据相关handlers
func queryHistoryData(historyService *service.HistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.HistoryQueryRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求参数格式错误: " + err.Error(),
			})
			return
		}

		// 设置默认值
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.Size <= 0 {
			req.Size = 20
		}

		response, err := historyService.QueryHistory(req)
		if err != nil {
			log.Printf("查询历史数据失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, response)
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
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "集群ID格式错误",
				})
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
			log.Printf("获取趋势数据失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":       data,
			"cluster_id": clusterID,
			"namespace":  namespace,
			"pod_name":   podName,
			"hours":      hours,
			"count":      len(data),
		})
	}
}

func getHistoryStatistics(historyService *service.HistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := historyService.GetStatistics()
		if err != nil {
			log.Printf("获取历史数据统计失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": stats,
		})
	}
}

func triggerDataCollection(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		persistenceStr := c.DefaultQuery("persistence", "true")
		enablePersistence := persistenceStr == "true"

		result, err := multiCollector.CollectAllClustersDataWithPersistence(c.Request.Context(), enablePersistence)
		if err != nil {
			log.Printf("触发数据收集失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "数据收集完成",
			"persistence": enablePersistence,
			"result":      result,
		})
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
			log.Printf("清理过期数据失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "数据清理完成",
			"retention_days": retentionDays,
		})
	}
}

// 调度服务管理handlers
func getScheduleStatus(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := scheduleService.GetStatus()
		c.JSON(http.StatusOK, gin.H{
			"data": status,
		})
	}
}

func startScheduleService(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := scheduleService.Start(c.Request.Context())
		if err != nil {
			log.Printf("启动调度服务失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "调度服务启动成功",
		})
	}
}

func stopScheduleService(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := scheduleService.Stop()
		if err != nil {
			log.Printf("停止调度服务失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "调度服务停止成功",
		})
	}
}

func getScheduleJobs(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobs := scheduleService.GetAllJobs()
		c.JSON(http.StatusOK, gin.H{
			"data":  jobs,
			"count": len(jobs),
		})
	}
}

func restartClusterJob(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterIDStr := c.Param("cluster_id")
		clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的集群ID",
			})
			return
		}

		err = scheduleService.RestartJob(uint(clusterID))
		if err != nil {
			log.Printf("重启集群任务失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "集群任务重启成功",
		})
	}
}

func updateScheduleSettings(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings service.GlobalScheduleSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求参数格式错误: " + err.Error(),
			})
			return
		}

		err := scheduleService.UpdateSettings(&settings)
		if err != nil {
			log.Printf("更新调度设置失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "调度设置更新成功",
			"data":    settings,
		})
	}
}