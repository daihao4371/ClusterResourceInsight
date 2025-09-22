package router

import (
	"cluster-resource-insight/internal/api"
	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有API路由 - 将路由配置从handlers.go中分离出来
func SetupRoutes(r *gin.RouterGroup, resourceCollector *collector.ResourceCollector) {
	// 直接复制原handlers.go中的路由配置逻辑，但调用api包中的处理器函数
	
	// 原有的分析接口
	r.GET("/analysis", api.GetResourceAnalysis(resourceCollector))
	r.GET("/pods", api.GetPodsData(resourceCollector))
	r.GET("/health", api.HealthCheck)

	// 新增系统统计接口
	r.GET("/stats", api.GetSystemStats(resourceCollector))

	// 新增的资源统计接口
	multiCollector := collector.NewMultiClusterResourceCollector()
	statisticsGroup := r.Group("/statistics")
	{
		statisticsGroup.GET("/top-memory-request", api.GetTopMemoryRequestPods(multiCollector))
		statisticsGroup.GET("/top-cpu-request", api.GetTopCPURequestPods(multiCollector))
		statisticsGroup.GET("/namespace-summary", api.GetNamespacesSummary(multiCollector))
	}

	// 新增的命名空间相关接口
	namespacesGroup := r.Group("/namespaces")
	{
		namespacesGroup.GET("", api.GetAllNamespaces(multiCollector))
		namespacesGroup.GET("/:namespace/pods", api.GetNamespacePods(multiCollector))
		namespacesGroup.GET("/:namespace/tree-data", api.GetNamespaceTreeData(multiCollector))
	}

	// Pod搜索与分页接口
	podsGroup := r.Group("/pods")
	{
		podsGroup.GET("/search", api.SearchPods(multiCollector))
		podsGroup.GET("/list", api.ListPods(multiCollector))
		podsGroup.GET("/problems", api.GetProblemsWithPagination(multiCollector))
		podsGroup.GET("/filter-options", api.GetFilterOptions(multiCollector)) // 新增筛选选项接口
	}

	// 新增的历史数据接口
	historyService := service.NewHistoryService()
	historyGroup := r.Group("/history")
	{
		historyGroup.GET("/query", api.QueryHistoryData(historyService))
		historyGroup.GET("/trends", api.GetTrendData(historyService))
		historyGroup.GET("/system-trends", api.GetSystemTrendData(historyService))
		historyGroup.GET("/statistics", api.GetHistoryStatistics(historyService))
		historyGroup.POST("/collect", api.TriggerDataCollection(multiCollector))
		historyGroup.DELETE("/cleanup", api.CleanupOldData(historyService))
	}

	// 新增的调度管理接口
	scheduleService := service.NewScheduleService()
	scheduleGroup := r.Group("/schedule")
	{
		scheduleGroup.GET("/status", api.GetScheduleStatus(scheduleService))
		scheduleGroup.POST("/start", api.StartScheduleService(scheduleService))
		scheduleGroup.POST("/stop", api.StopScheduleService(scheduleService))
		scheduleGroup.GET("/jobs", api.GetScheduleJobs(scheduleService))
		scheduleGroup.POST("/jobs/:cluster_id/restart", api.RestartClusterJob(scheduleService))
		scheduleGroup.PUT("/settings", api.UpdateScheduleSettings(scheduleService))
	}

	// 新增活动和告警接口
	activityService := service.NewActivityService()
	activitiesGroup := r.Group("/activities")
	{
		activitiesGroup.GET("/recent", api.GetRecentActivities(activityService))
		activitiesGroup.DELETE("/cleanup", api.CleanupOldActivitiesAndAlerts(activityService))
	}

	alertsGroup := r.Group("/alerts")
	{
		alertsGroup.GET("/recent", api.GetRecentAlerts(activityService))
		alertsGroup.PUT("/:id/status", api.UpdateAlertStatus(activityService))
		alertsGroup.PUT("/:id/resolve", api.ResolveAlert(activityService))
		alertsGroup.PUT("/:id/dismiss", api.DismissAlert(activityService))
		alertsGroup.GET("/:id", api.GetAlertDetails(activityService))
	}

	// 集群管理接口
	clusterService := service.NewClusterService()
	clusterGroup := r.Group("/clusters")
	{
		clusterGroup.POST("", api.CreateCluster(clusterService))
		clusterGroup.GET("", api.GetAllClusters(clusterService))
		clusterGroup.GET("/:id", api.GetClusterByID(clusterService))
		clusterGroup.PUT("/:id", api.UpdateCluster(clusterService))
		clusterGroup.DELETE("/:id", api.DeleteCluster(clusterService))
		clusterGroup.POST("/:id/test", api.TestClusterConnection(clusterService))
		clusterGroup.POST("/test", api.TestClusterByConfig(clusterService))
		clusterGroup.POST("/batch-test", api.BatchTestAllClusters(clusterService))
	}
}