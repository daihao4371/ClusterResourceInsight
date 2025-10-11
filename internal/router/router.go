package router

import (
	"cluster-resource-insight/internal/api"
	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/middleware"
	"cluster-resource-insight/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有API路由
func SetupRoutes(r *gin.RouterGroup, resourceCollector *collector.ResourceCollector) {
	// 应用CORS中间件
	r.Use(middleware.CORS())

	// ============== 公开接口(无需认证) ==============
	userService := service.NewUserService()

	// 认证接口
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", api.Login(userService))
		authGroup.POST("/logout", api.Logout())
	}

	// 健康检查接口
	r.GET("/health", api.HealthCheck)

	// ============== 需要认证的接口 ==============
	// 应用JWT认证中间件
	authenticated := r.Group("")
	authenticated.Use(middleware.JWTAuth())
	{
		// 当前用户信息
		authenticated.GET("/auth/me", api.GetCurrentUser(userService))

		// 原有的分析接口
		authenticated.GET("/analysis", api.GetResourceAnalysis(resourceCollector))
		authenticated.GET("/pods", api.GetPodsData(resourceCollector))
		authenticated.GET("/stats", api.GetSystemStats(resourceCollector))

		// 资源统计接口
		multiCollector := collector.NewMultiClusterResourceCollector()
		statisticsGroup := authenticated.Group("/statistics")
		{
			statisticsGroup.GET("/top-memory-request", api.GetTopMemoryRequestPods(multiCollector))
			statisticsGroup.GET("/top-cpu-request", api.GetTopCPURequestPods(multiCollector))
			statisticsGroup.GET("/top-resource-namespaces", api.GetTopResourceNamespaces(multiCollector))
			statisticsGroup.GET("/resource-distribution", api.GetResourceDistribution(multiCollector))
		}

		// 命名空间相关接口
		namespacesGroup := authenticated.Group("/namespaces")
		{
			namespacesGroup.GET("", api.GetAllNamespaces(multiCollector))
			namespacesGroup.GET("/:namespace/pods", api.GetNamespacePods(multiCollector))
			namespacesGroup.GET("/:namespace/tree-data", api.GetNamespaceTreeData(multiCollector))
		}

		// Pod搜索与分页接口
		podsGroup := authenticated.Group("/pods")
		{
			podsGroup.GET("/search", api.SearchPods(multiCollector))
			podsGroup.GET("/list", api.ListPods(multiCollector))
			podsGroup.GET("/problems", api.GetProblemsWithPagination(multiCollector))
			podsGroup.GET("/filter-options", api.GetFilterOptions(multiCollector))
			podsGroup.GET("/:cluster/:namespace/:pod/detail", api.GetPodDetailAnalysis(multiCollector))
			podsGroup.GET("/:cluster/:namespace/:pod/trend", api.GetPodTrendData(multiCollector))
		}

		// 历史数据接口
		historyService := service.NewHistoryService()
		historyGroup := authenticated.Group("/history")
		{
			historyGroup.GET("/query", api.QueryHistoryData(historyService))
			historyGroup.GET("/trends", api.GetTrendData(historyService))
			historyGroup.GET("/system-trends", api.GetSystemTrendData(historyService))
			historyGroup.GET("/statistics", api.GetHistoryStatistics(historyService))
			historyGroup.POST("/collect", api.TriggerDataCollection(multiCollector))
			historyGroup.DELETE("/cleanup", api.CleanupOldData(historyService))
		}

		// 调度管理接口
		scheduleService := service.NewScheduleService()
		scheduleGroup := authenticated.Group("/schedule")
		{
			scheduleGroup.GET("/status", api.GetScheduleStatus(scheduleService))
			scheduleGroup.POST("/start", api.StartScheduleService(scheduleService))
			scheduleGroup.POST("/stop", api.StopScheduleService(scheduleService))
			scheduleGroup.GET("/jobs", api.GetScheduleJobs(scheduleService))
			scheduleGroup.POST("/jobs/:cluster_id/start", api.StartClusterJob(scheduleService))
			scheduleGroup.POST("/jobs/:cluster_id/stop", api.StopClusterJob(scheduleService))
			scheduleGroup.POST("/jobs/:cluster_id/restart", api.RestartClusterJob(scheduleService))
			scheduleGroup.POST("/jobs/:cluster_id/execute", api.ExecuteJobNow(scheduleService))
			scheduleGroup.PUT("/jobs/:cluster_id/settings", api.UpdateJobSettings(scheduleService))
			scheduleGroup.POST("/jobs/refresh", api.RefreshJobsFromClusters(scheduleService))
			scheduleGroup.GET("/settings", api.GetScheduleSettings(scheduleService))
			scheduleGroup.PUT("/settings", api.UpdateScheduleSettings(scheduleService))
		}

		// 活动和告警接口
		activityService := service.NewActivityService()
		activitiesGroup := authenticated.Group("/activities")
		{
			activitiesGroup.GET("/recent", api.GetRecentActivities(activityService))
			activitiesGroup.DELETE("/cleanup", api.CleanupOldActivitiesAndAlerts(activityService))
			optimizationGroup := activitiesGroup.Group("/optimization")
			{
				optimizationGroup.GET("/config", api.GetOptimizationConfig(activityService))
				optimizationGroup.PUT("/config", api.UpdateOptimizationConfig(activityService))
				optimizationGroup.POST("/execute", api.ExecuteOptimization(activityService))
			}
			activitiesGroup.GET("/stats", api.GetActivityStats(activityService))
			activitiesGroup.GET("/database-stats", api.GetDatabaseStats(activityService))
		}

		alertsGroup := authenticated.Group("/alerts")
		{
			alertsGroup.GET("/recent", api.GetRecentAlerts(activityService))
			alertsGroup.PUT("/:id/status", api.UpdateAlertStatus(activityService))
			alertsGroup.PUT("/:id/resolve", api.ResolveAlert(activityService))
			alertsGroup.PUT("/:id/dismiss", api.DismissAlert(activityService))
			alertsGroup.GET("/:id", api.GetAlertDetails(activityService))
			alertsGroup.POST("/deduplicate", api.DeduplicateAlerts(activityService))
		}

		// 集群管理接口
		clusterService := service.NewClusterService()
		clusterGroup := authenticated.Group("/clusters")
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

		// 用户管理接口
		userGroup := authenticated.Group("/users")
		{
			userGroup.POST("", api.CreateUser(userService))
			userGroup.GET("", api.ListUsers(userService))
			userGroup.GET("/:id", api.GetUserByID(userService))
			userGroup.PUT("/:id", api.UpdateUser(userService))
			userGroup.DELETE("/:id", api.DeleteUser(userService))
			userGroup.PUT("/:id/password", api.ChangePassword(userService))
			userGroup.PUT("/:id/reset-password", api.ResetPassword(userService))
		}

		// 角色管理接口
		roleService := service.NewRoleService()
		roleGroup := authenticated.Group("/roles")
		{
			roleGroup.POST("", api.CreateRole(roleService))
			roleGroup.GET("", api.ListRoles(roleService))
			roleGroup.GET("/active", api.GetAllActiveRoles(roleService))
			roleGroup.GET("/:id", api.GetRoleByID(roleService))
			roleGroup.PUT("/:id", api.UpdateRole(roleService))
			roleGroup.DELETE("/:id", api.DeleteRole(roleService))
			roleGroup.GET("/:id/users", api.GetRoleUsers(roleService))
			roleGroup.GET("/:id/permissions", api.GetRolePermissions(roleService))
		}

		// API管理接口
		apiService := service.NewApiService()
		apiGroup := authenticated.Group("/apis")
		{
			apiGroup.POST("", api.CreateApi(apiService))
			apiGroup.GET("", api.ListApis(apiService))
			apiGroup.GET("/groups", api.GetApiGroups(apiService))
			apiGroup.GET("/statistics", api.GetApiStatistics(apiService))
			apiGroup.GET("/:id", api.GetApiByID(apiService))
			apiGroup.PUT("/:id", api.UpdateApi(apiService))
			apiGroup.DELETE("/:id", api.DeleteApi(apiService))
			apiGroup.POST("/batch", api.BatchCreateApis(apiService))
			apiGroup.POST("/sync", api.SyncApis(apiService))
		}

		// 审计日志接口
		auditService := service.NewAuditService()
		auditGroup := authenticated.Group("/audit")
		{
			auditGroup.GET("/logs", api.QueryAuditLogs(auditService))
			auditGroup.GET("/logs/:id", api.GetAuditLogByID(auditService))
			auditGroup.GET("/users/:user_id/logs", api.GetUserAuditLogs(auditService))
			auditGroup.GET("/statistics", api.GetAuditStatistics(auditService))
			auditGroup.GET("/trend", api.GetOperationTrend(auditService))
			auditGroup.DELETE("/cleanup", api.CleanupOldAuditLogs(auditService))
			auditGroup.GET("/export", api.ExportAuditLogs(auditService))
		}
	}
}