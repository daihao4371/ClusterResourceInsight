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

	// 认证接口 - 用户登录登出功能
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", api.Login(userService))       // 用户登录，验证用户名密码并返回JWT token
		authGroup.POST("/logout", api.Logout())                 // 用户登出，清除会话信息
	}

	// 健康检查接口 - 用于监控系统运行状态
	r.GET("/health", api.HealthCheck) // 返回服务健康状态，供负载均衡器和监控系统使用

	// ============== 需要认证的接口 ==============
	// 创建审计服务（用于审计中间件）
	auditService := service.NewAuditService()

	// 应用JWT认证中间件
	authenticated := r.Group("")
	authenticated.Use(middleware.JWTAuth())
	// 应用审计日志中间件（需要在JWTAuth之后，确保能获取到用户信息）
	authenticated.Use(middleware.AuditMiddleware(auditService))
	{
		// 获取当前登录用户信息 - 用于前端展示用户名、角色等
		authenticated.GET("/auth/me", api.GetCurrentUser(userService))

		// 资源分析基础接口 - 提供资源配置分析、Pod数据和系统统计
		authenticated.GET("/analysis", api.GetResourceAnalysis(resourceCollector)) // 分析所有Pod的资源配置合理性
		authenticated.GET("/pods", api.GetPodsData(resourceCollector))             // 获取Pod列表及资源使用数据
		authenticated.GET("/stats", api.GetSystemStats(resourceCollector))         // 获取系统整体统计信息

		// 资源统计接口 - 提供各维度的资源排行和分布分析
		multiCollector := collector.NewMultiClusterResourceCollector()
		statisticsGroup := authenticated.Group("/statistics")
		{
			statisticsGroup.GET("/top-memory-request", api.GetTopMemoryRequestPods(multiCollector))    // 获取内存请求量Top N的Pod
			statisticsGroup.GET("/top-cpu-request", api.GetTopCPURequestPods(multiCollector))          // 获取CPU请求量Top N的Pod
			statisticsGroup.GET("/top-resource-namespaces", api.GetTopResourceNamespaces(multiCollector)) // 获取资源占用Top N的命名空间
			statisticsGroup.GET("/resource-distribution", api.GetResourceDistribution(multiCollector))  // 获取资源分布情况
		}

		// 命名空间相关接口 - 提供命名空间级别的资源查看
		namespacesGroup := authenticated.Group("/namespaces")
		{
			namespacesGroup.GET("", api.GetAllNamespaces(multiCollector))                       // 获取所有命名空间列表
			namespacesGroup.GET("/:namespace/pods", api.GetNamespacePods(multiCollector))       // 获取指定命名空间下的所有Pod
			namespacesGroup.GET("/:namespace/tree-data", api.GetNamespaceTreeData(multiCollector)) // 获取命名空间的树形结构数据
		}

		// Pod搜索与分页接口 - 提供灵活的Pod查询和详情分析
		podsGroup := authenticated.Group("/pods")
		{
			podsGroup.GET("/search", api.SearchPods(multiCollector))                                   // 按条件搜索Pod
			podsGroup.GET("/list", api.ListPods(multiCollector))                                       // 分页获取Pod列表
			podsGroup.GET("/problems", api.GetProblemsWithPagination(multiCollector))                  // 分页获取存在问题的Pod列表
			podsGroup.GET("/filter-options", api.GetFilterOptions(multiCollector))                     // 获取可用的过滤选项（集群、命名空间等）
			podsGroup.GET("/:cluster/:namespace/:pod/detail", api.GetPodDetailAnalysis(multiCollector)) // 获取指定Pod的详细分析数据
			podsGroup.GET("/:cluster/:namespace/:pod/trend", api.GetPodTrendData(multiCollector))      // 获取指定Pod的资源使用趋势
		}

		// 历史数据接口 - 提供历史监控数据的查询和管理
		historyService := service.NewHistoryService()
		historyGroup := authenticated.Group("/history")
		{
			historyGroup.GET("/query", api.QueryHistoryData(historyService))           // 查询历史监控数据，支持多种过滤条件
			historyGroup.GET("/trends", api.GetTrendData(historyService))              // 获取资源使用趋势数据
			historyGroup.GET("/system-trends", api.GetSystemTrendData(historyService)) // 获取系统级别的趋势数据
			historyGroup.GET("/statistics", api.GetHistoryStatistics(historyService))  // 获取历史数据统计信息
			historyGroup.POST("/collect", api.TriggerDataCollection(multiCollector))   // 手动触发数据采集任务
			historyGroup.DELETE("/cleanup", api.CleanupOldData(historyService))        // 清理过期的历史数据，释放存储空间
		}

		// 调度管理接口 - 管理定时数据采集任务的执行
		scheduleService := service.NewScheduleService()
		scheduleGroup := authenticated.Group("/schedule")
		{
			scheduleGroup.GET("/status", api.GetScheduleStatus(scheduleService))                         // 获取调度服务的运行状态
			scheduleGroup.POST("/start", api.StartScheduleService(scheduleService))                      // 启动调度服务，开始自动采集
			scheduleGroup.POST("/stop", api.StopScheduleService(scheduleService))                        // 停止调度服务，暂停自动采集
			scheduleGroup.GET("/jobs", api.GetScheduleJobs(scheduleService))                             // 获取所有集群的调度任务列表
			scheduleGroup.POST("/jobs/:cluster_id/start", api.StartClusterJob(scheduleService))          // 启动指定集群的调度任务
			scheduleGroup.POST("/jobs/:cluster_id/stop", api.StopClusterJob(scheduleService))            // 停止指定集群的调度任务
			scheduleGroup.POST("/jobs/:cluster_id/restart", api.RestartClusterJob(scheduleService))      // 重启指定集群的调度任务
			scheduleGroup.POST("/jobs/:cluster_id/execute", api.ExecuteJobNow(scheduleService))          // 立即执行指定集群的采集任务
			scheduleGroup.PUT("/jobs/:cluster_id/settings", api.UpdateJobSettings(scheduleService))      // 更新指定集群的调度设置（如采集间隔）
			scheduleGroup.POST("/jobs/refresh", api.RefreshJobsFromClusters(scheduleService))            // 从数据库刷新集群列表，同步调度任务
			scheduleGroup.GET("/settings", api.GetScheduleSettings(scheduleService))                     // 获取全局调度设置
			scheduleGroup.PUT("/settings", api.UpdateScheduleSettings(scheduleService))                  // 更新全局调度设置
		}

		// 活动和告警接口 - 提供系统活动和告警的查询与管理
		activityService := service.NewActivityService()
		activitiesGroup := authenticated.Group("/activities")
		{
			activitiesGroup.GET("/recent", api.GetRecentActivities(activityService))                     // 获取最近的系统活动记录
			activitiesGroup.DELETE("/cleanup", api.CleanupOldActivitiesAndAlerts(activityService))      // 清理过期的活动和告警数据
			// 活动优化子组 - 管理活动记录的降噪和优化配置
			optimizationGroup := activitiesGroup.Group("/optimization")
			{
				optimizationGroup.GET("/config", api.GetOptimizationConfig(activityService))             // 获取活动优化配置
				optimizationGroup.PUT("/config", api.UpdateOptimizationConfig(activityService))          // 更新活动优化配置
				optimizationGroup.POST("/execute", api.ExecuteOptimization(activityService))             // 手动执行活动优化任务
			}
			activitiesGroup.GET("/stats", api.GetActivityStats(activityService))                         // 获取活动统计信息
			activitiesGroup.GET("/database-stats", api.GetDatabaseStats(activityService))                // 获取数据库统计信息
		}

		// 告警接口 - 管理系统告警的查看和处理
		alertsGroup := authenticated.Group("/alerts")
		{
			alertsGroup.GET("/recent", api.GetRecentAlerts(activityService))                             // 获取最近的告警列表
			alertsGroup.PUT("/:id/status", api.UpdateAlertStatus(activityService))                       // 更新告警状态
			alertsGroup.PUT("/:id/resolve", api.ResolveAlert(activityService))                           // 将告警标记为已解决
			alertsGroup.PUT("/:id/dismiss", api.DismissAlert(activityService))                           // 忽略告警，不再显示
			alertsGroup.GET("/:id", api.GetAlertDetails(activityService))                                // 获取告警详情
			alertsGroup.POST("/deduplicate", api.DeduplicateAlerts(activityService))                     // 执行告警去重，合并重复告警
		}

		// 集群管理接口 - 管理Kubernetes集群的连接和配置
		clusterService := service.NewClusterService()
		clusterGroup := authenticated.Group("/clusters")
		{
			clusterGroup.POST("", api.CreateCluster(clusterService))                                     // 创建新的集群配置
			clusterGroup.GET("", api.GetAllClusters(clusterService))                                     // 获取所有已配置的集群列表
			clusterGroup.GET("/:id", api.GetClusterByID(clusterService))                                 // 获取指定集群的详细信息
			clusterGroup.PUT("/:id", api.UpdateCluster(clusterService))                                  // 更新集群配置信息
			clusterGroup.DELETE("/:id", api.DeleteCluster(clusterService))                               // 删除集群配置
			clusterGroup.POST("/:id/test", api.TestClusterConnection(clusterService))                    // 测试指定集群的连接状态
			clusterGroup.POST("/test", api.TestClusterByConfig(clusterService))                          // 使用临时配置测试集群连接
			clusterGroup.POST("/batch-test", api.BatchTestAllClusters(clusterService))                   // 批量测试所有集群的连接状态
		}

		// 用户管理接口 - 管理系统用户账号
		userGroup := authenticated.Group("/users")
		{
			userGroup.POST("", api.CreateUser(userService))                                              // 创建新用户
			userGroup.GET("", api.ListUsers(userService))                                                // 获取用户列表
			userGroup.GET("/:id", api.GetUserByID(userService))                                          // 获取指定用户详情
			userGroup.PUT("/:id", api.UpdateUser(userService))                                           // 更新用户信息
			userGroup.DELETE("/:id", api.DeleteUser(userService))                                        // 删除用户
			userGroup.PUT("/:id/password", api.ChangePassword(userService))                              // 用户修改自己的密码
			userGroup.PUT("/:id/reset-password", api.ResetPassword(userService))                         // 管理员重置用户密码
		}

		// 角色管理接口 - 管理用户角色和权限
		roleService := service.NewRoleService()
		roleGroup := authenticated.Group("/roles")
		{
			roleGroup.POST("", api.CreateRole(roleService))                                              // 创建新角色
			roleGroup.GET("", api.ListRoles(roleService))                                                // 获取角色列表
			roleGroup.GET("/active", api.GetAllActiveRoles(roleService))                                 // 获取所有活跃角色（用于下拉选择）
			roleGroup.GET("/:id", api.GetRoleByID(roleService))                                          // 获取角色详情
			roleGroup.PUT("/:id", api.UpdateRole(roleService))                                           // 更新角色信息和权限
			roleGroup.DELETE("/:id", api.DeleteRole(roleService))                                        // 删除角色
			roleGroup.GET("/:id/users", api.GetRoleUsers(roleService))                                   // 获取拥有该角色的用户列表
			roleGroup.GET("/:id/permissions", api.GetRolePermissions(roleService))                       // 获取角色的权限列表
		}

		// API管理接口 - 管理系统API端点的元数据
		apiService := service.NewApiService()
		apiGroup := authenticated.Group("/apis")
		{
			apiGroup.POST("", api.CreateApi(apiService))                                                 // 创建新的API端点记录
			apiGroup.GET("", api.ListApis(apiService))                                                   // 获取API端点列表
			apiGroup.GET("/groups", api.GetApiGroups(apiService))                                        // 获取API分组列表
			apiGroup.GET("/statistics", api.GetApiStatistics(apiService))                                // 获取API统计信息
			apiGroup.GET("/:id", api.GetApiByID(apiService))                                             // 获取指定API详情
			apiGroup.PUT("/:id", api.UpdateApi(apiService))                                              // 更新API信息
			apiGroup.DELETE("/:id", api.DeleteApi(apiService))                                           // 删除API记录
			apiGroup.POST("/batch", api.BatchCreateApis(apiService))                                     // 批量创建API记录
			apiGroup.POST("/sync", api.SyncApis(apiService))                                             // 同步系统路由到API表
		}

		// 审计日志接口 - 记录和查询用户操作日志
		auditGroup := authenticated.Group("/audit")
		{
			auditGroup.GET("/logs", api.QueryAuditLogs(auditService))                                    // 查询审计日志，支持多种过滤条件
			auditGroup.GET("/logs/:id", api.GetAuditLogByID(auditService))                               // 获取指定审计日志详情
			auditGroup.GET("/users/:user_id/logs", api.GetUserAuditLogs(auditService))                   // 获取指定用户的操作日志
			auditGroup.GET("/statistics", api.GetAuditStatistics(auditService))                          // 获取审计日志统计信息
			auditGroup.GET("/trend", api.GetOperationTrend(auditService))                                // 获取操作趋势分析
			auditGroup.DELETE("/cleanup", api.CleanupOldAuditLogs(auditService))                         // 清理过期的审计日志
			auditGroup.GET("/export", api.ExportAuditLogs(auditService))                                 // 导出审计日志为CSV等格式
		}
	}
}