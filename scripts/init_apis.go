package main

import (
	"flag"
	"fmt"
	"log"

	"cluster-resource-insight/internal/config"
	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/models"
)

// 系统所有API定义
func getAllSystemAPIs() []models.ApiEndpoint {
	return []models.ApiEndpoint{
		// 认证相关API
		{Path: "/api/auth/login", Method: "POST", Group: "auth", Description: "用户登录", IsPublic: true, Status: "active"},
		{Path: "/api/auth/logout", Method: "POST", Group: "auth", Description: "用户登出", IsPublic: true, Status: "active"},
		{Path: "/api/auth/me", Method: "GET", Group: "auth", Description: "获取当前用户信息", Permission: "user:view", Status: "active"},

		// 健康检查
		{Path: "/api/health", Method: "GET", Group: "system", Description: "健康检查", IsPublic: true, Status: "active"},

		// 资源分析API
		{Path: "/api/analysis", Method: "GET", Group: "resource", Description: "获取资源分析数据", Permission: "resource:view", Status: "active"},
		{Path: "/api/pods", Method: "GET", Group: "resource", Description: "获取Pod数据", Permission: "resource:view", Status: "active"},
		{Path: "/api/stats", Method: "GET", Group: "resource", Description: "获取系统统计", Permission: "resource:view", Status: "active"},

		// 资源统计API
		{Path: "/api/statistics/top-memory-request", Method: "GET", Group: "statistics", Description: "获取内存请求Top Pod", Permission: "resource:view", Status: "active"},
		{Path: "/api/statistics/top-cpu-request", Method: "GET", Group: "statistics", Description: "获取CPU请求Top Pod", Permission: "resource:view", Status: "active"},
		{Path: "/api/statistics/top-resource-namespaces", Method: "GET", Group: "statistics", Description: "获取资源Top命名空间", Permission: "resource:view", Status: "active"},
		{Path: "/api/statistics/resource-distribution", Method: "GET", Group: "statistics", Description: "获取资源分布", Permission: "resource:view", Status: "active"},

		// 命名空间API
		{Path: "/api/namespaces", Method: "GET", Group: "namespace", Description: "获取所有命名空间", Permission: "resource:view", Status: "active"},
		{Path: "/api/namespaces/:namespace/pods", Method: "GET", Group: "namespace", Description: "获取命名空间下的Pod", Permission: "resource:view", Status: "active"},
		{Path: "/api/namespaces/:namespace/tree-data", Method: "GET", Group: "namespace", Description: "获取命名空间树形数据", Permission: "resource:view", Status: "active"},

		// Pod API
		{Path: "/api/pods/search", Method: "GET", Group: "pod", Description: "搜索Pod", Permission: "resource:view", Status: "active"},
		{Path: "/api/pods/list", Method: "GET", Group: "pod", Description: "获取Pod列表", Permission: "resource:view", Status: "active"},
		{Path: "/api/pods/problems", Method: "GET", Group: "pod", Description: "获取问题Pod列表", Permission: "resource:view", Status: "active"},
		{Path: "/api/pods/filter-options", Method: "GET", Group: "pod", Description: "获取过滤选项", Permission: "resource:view", Status: "active"},
		{Path: "/api/pods/:cluster/:namespace/:pod/detail", Method: "GET", Group: "pod", Description: "获取Pod详情分析", Permission: "resource:view", Status: "active"},
		{Path: "/api/pods/:cluster/:namespace/:pod/trend", Method: "GET", Group: "pod", Description: "获取Pod趋势数据", Permission: "resource:view", Status: "active"},

		// 历史数据API
		{Path: "/api/history/query", Method: "GET", Group: "history", Description: "查询历史数据", Permission: "history:view", Status: "active"},
		{Path: "/api/history/trends", Method: "GET", Group: "history", Description: "获取趋势数据", Permission: "history:view", Status: "active"},
		{Path: "/api/history/system-trends", Method: "GET", Group: "history", Description: "获取系统趋势数据", Permission: "history:view", Status: "active"},
		{Path: "/api/history/statistics", Method: "GET", Group: "history", Description: "获取历史统计", Permission: "history:view", Status: "active"},
		{Path: "/api/history/collect", Method: "POST", Group: "history", Description: "触发数据采集", Permission: "history:manage", Status: "active"},
		{Path: "/api/history/cleanup", Method: "DELETE", Group: "history", Description: "清理历史数据", Permission: "history:manage", Status: "active"},

		// 调度管理API
		{Path: "/api/schedule/status", Method: "GET", Group: "schedule", Description: "获取调度状态", Permission: "schedule:view", Status: "active"},
		{Path: "/api/schedule/start", Method: "POST", Group: "schedule", Description: "启动调度服务", Permission: "schedule:manage", Status: "active"},
		{Path: "/api/schedule/stop", Method: "POST", Group: "schedule", Description: "停止调度服务", Permission: "schedule:manage", Status: "active"},
		{Path: "/api/schedule/jobs", Method: "GET", Group: "schedule", Description: "获取调度任务列表", Permission: "schedule:view", Status: "active"},
		{Path: "/api/schedule/jobs/:cluster_id/start", Method: "POST", Group: "schedule", Description: "启动集群任务", Permission: "schedule:manage", Status: "active"},
		{Path: "/api/schedule/jobs/:cluster_id/stop", Method: "POST", Group: "schedule", Description: "停止集群任务", Permission: "schedule:manage", Status: "active"},
		{Path: "/api/schedule/jobs/:cluster_id/restart", Method: "POST", Group: "schedule", Description: "重启集群任务", Permission: "schedule:manage", Status: "active"},
		{Path: "/api/schedule/jobs/:cluster_id/execute", Method: "POST", Group: "schedule", Description: "立即执行任务", Permission: "schedule:manage", Status: "active"},
		{Path: "/api/schedule/jobs/:cluster_id/settings", Method: "PUT", Group: "schedule", Description: "更新任务设置", Permission: "schedule:manage", Status: "active"},
		{Path: "/api/schedule/jobs/refresh", Method: "POST", Group: "schedule", Description: "刷新集群任务", Permission: "schedule:manage", Status: "active"},
		{Path: "/api/schedule/settings", Method: "GET", Group: "schedule", Description: "获取调度设置", Permission: "schedule:view", Status: "active"},
		{Path: "/api/schedule/settings", Method: "PUT", Group: "schedule", Description: "更新调度设置", Permission: "schedule:manage", Status: "active"},

		// 活动和告警API
		{Path: "/api/activities/recent", Method: "GET", Group: "activity", Description: "获取最近活动", Permission: "activity:view", Status: "active"},
		{Path: "/api/activities/cleanup", Method: "DELETE", Group: "activity", Description: "清理旧活动和告警", Permission: "activity:manage", Status: "active"},
		{Path: "/api/activities/optimization/config", Method: "GET", Group: "activity", Description: "获取优化配置", Permission: "activity:view", Status: "active"},
		{Path: "/api/activities/optimization/config", Method: "PUT", Group: "activity", Description: "更新优化配置", Permission: "activity:manage", Status: "active"},
		{Path: "/api/activities/optimization/execute", Method: "POST", Group: "activity", Description: "执行优化", Permission: "activity:manage", Status: "active"},
		{Path: "/api/activities/stats", Method: "GET", Group: "activity", Description: "获取活动统计", Permission: "activity:view", Status: "active"},
		{Path: "/api/activities/database-stats", Method: "GET", Group: "activity", Description: "获取数据库统计", Permission: "activity:view", Status: "active"},

		// 告警API
		{Path: "/api/alerts/recent", Method: "GET", Group: "alert", Description: "获取最近告警", Permission: "alert:view", Status: "active"},
		{Path: "/api/alerts/:id/status", Method: "PUT", Group: "alert", Description: "更新告警状态", Permission: "alert:manage", Status: "active"},
		{Path: "/api/alerts/:id/resolve", Method: "PUT", Group: "alert", Description: "解决告警", Permission: "alert:manage", Status: "active"},
		{Path: "/api/alerts/:id/dismiss", Method: "PUT", Group: "alert", Description: "忽略告警", Permission: "alert:manage", Status: "active"},
		{Path: "/api/alerts/:id", Method: "GET", Group: "alert", Description: "获取告警详情", Permission: "alert:view", Status: "active"},
		{Path: "/api/alerts/deduplicate", Method: "POST", Group: "alert", Description: "去重告警", Permission: "alert:manage", Status: "active"},

		// 集群管理API
		{Path: "/api/clusters", Method: "POST", Group: "cluster", Description: "创建集群", Permission: "cluster:create", Status: "active"},
		{Path: "/api/clusters", Method: "GET", Group: "cluster", Description: "获取所有集群", Permission: "cluster:view", Status: "active"},
		{Path: "/api/clusters/:id", Method: "GET", Group: "cluster", Description: "获取集群详情", Permission: "cluster:view", Status: "active"},
		{Path: "/api/clusters/:id", Method: "PUT", Group: "cluster", Description: "更新集群", Permission: "cluster:update", Status: "active"},
		{Path: "/api/clusters/:id", Method: "DELETE", Group: "cluster", Description: "删除集群", Permission: "cluster:delete", Status: "active"},
		{Path: "/api/clusters/:id/test", Method: "POST", Group: "cluster", Description: "测试集群连接", Permission: "cluster:test", Status: "active"},
		{Path: "/api/clusters/test", Method: "POST", Group: "cluster", Description: "测试集群配置连接", Permission: "cluster:test", Status: "active"},
		{Path: "/api/clusters/batch-test", Method: "POST", Group: "cluster", Description: "批量测试集群连接", Permission: "cluster:test", Status: "active"},

		// 用户管理API
		{Path: "/api/users", Method: "POST", Group: "user", Description: "创建用户", Permission: "user:create", Status: "active"},
		{Path: "/api/users", Method: "GET", Group: "user", Description: "获取用户列表", Permission: "user:view", Status: "active"},
		{Path: "/api/users/:id", Method: "GET", Group: "user", Description: "获取用户详情", Permission: "user:view", Status: "active"},
		{Path: "/api/users/:id", Method: "PUT", Group: "user", Description: "更新用户", Permission: "user:update", Status: "active"},
		{Path: "/api/users/:id", Method: "DELETE", Group: "user", Description: "删除用户", Permission: "user:delete", Status: "active"},
		{Path: "/api/users/:id/password", Method: "PUT", Group: "user", Description: "修改密码", Permission: "user:update", Status: "active"},
		{Path: "/api/users/:id/reset-password", Method: "PUT", Group: "user", Description: "重置密码", Permission: "user:update", Status: "active"},

		// 角色管理API
		{Path: "/api/roles", Method: "POST", Group: "role", Description: "创建角色", Permission: "role:create", Status: "active"},
		{Path: "/api/roles", Method: "GET", Group: "role", Description: "获取角色列表", Permission: "role:view", Status: "active"},
		{Path: "/api/roles/active", Method: "GET", Group: "role", Description: "获取所有活跃角色", Permission: "role:view", Status: "active"},
		{Path: "/api/roles/:id", Method: "GET", Group: "role", Description: "获取角色详情", Permission: "role:view", Status: "active"},
		{Path: "/api/roles/:id", Method: "PUT", Group: "role", Description: "更新角色", Permission: "role:update", Status: "active"},
		{Path: "/api/roles/:id", Method: "DELETE", Group: "role", Description: "删除角色", Permission: "role:delete", Status: "active"},
		{Path: "/api/roles/:id/users", Method: "GET", Group: "role", Description: "获取角色用户", Permission: "role:view", Status: "active"},
		{Path: "/api/roles/:id/permissions", Method: "GET", Group: "role", Description: "获取角色权限", Permission: "role:view", Status: "active"},

		// API管理
		{Path: "/api/apis", Method: "POST", Group: "api", Description: "创建API", Permission: "api:create", Status: "active"},
		{Path: "/api/apis", Method: "GET", Group: "api", Description: "获取API列表", Permission: "api:view", Status: "active"},
		{Path: "/api/apis/groups", Method: "GET", Group: "api", Description: "获取API分组", Permission: "api:view", Status: "active"},
		{Path: "/api/apis/statistics", Method: "GET", Group: "api", Description: "获取API统计", Permission: "api:view", Status: "active"},
		{Path: "/api/apis/:id", Method: "GET", Group: "api", Description: "获取API详情", Permission: "api:view", Status: "active"},
		{Path: "/api/apis/:id", Method: "PUT", Group: "api", Description: "更新API", Permission: "api:update", Status: "active"},
		{Path: "/api/apis/:id", Method: "DELETE", Group: "api", Description: "删除API", Permission: "api:delete", Status: "active"},
		{Path: "/api/apis/batch", Method: "POST", Group: "api", Description: "批量创建API", Permission: "api:create", Status: "active"},
		{Path: "/api/apis/sync", Method: "POST", Group: "api", Description: "同步API", Permission: "api:manage", Status: "active"},

		// 审计日志API
		{Path: "/api/audit/logs", Method: "GET", Group: "audit", Description: "查询审计日志", Permission: "audit:view", Status: "active"},
		{Path: "/api/audit/logs/:id", Method: "GET", Group: "audit", Description: "获取审计日志详情", Permission: "audit:view", Status: "active"},
		{Path: "/api/audit/users/:user_id/logs", Method: "GET", Group: "audit", Description: "获取用户审计日志", Permission: "audit:view", Status: "active"},
		{Path: "/api/audit/statistics", Method: "GET", Group: "audit", Description: "获取审计统计", Permission: "audit:view", Status: "active"},
		{Path: "/api/audit/trend", Method: "GET", Group: "audit", Description: "获取操作趋势", Permission: "audit:view", Status: "active"},
		{Path: "/api/audit/cleanup", Method: "DELETE", Group: "audit", Description: "清理旧审计日志", Permission: "audit:manage", Status: "active"},
		{Path: "/api/audit/export", Method: "GET", Group: "audit", Description: "导出审计日志", Permission: "audit:export", Status: "active"},
	}
}

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.toml", "配置文件路径")
	flag.Parse()

	log.Println("开始初始化API数据...")

	// 加载配置
	appConfig, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化数据库
	dbConfig := &database.DatabaseConfig{
		Host:     appConfig.Database.Host,
		Port:     appConfig.Database.Port,
		Username: appConfig.Database.Username,
		Password: appConfig.Database.Password,
		DBName:   appConfig.Database.Database,
		Charset:  appConfig.Database.Charset,
	}

	if err := database.InitDatabase(dbConfig); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.CloseDatabase()

	// 获取数据库实例
	db := database.GetDB()

	// 获取所有系统API
	apis := getAllSystemAPIs()

	// 统计
	var created, updated, skipped int

	// 批量插入或更新
	for _, api := range apis {
		var existingAPI models.ApiEndpoint
		result := db.Where("path = ? AND method = ?", api.Path, api.Method).First(&existingAPI)

		if result.Error != nil {
			// API不存在,创建新的
			if err := db.Create(&api).Error; err != nil {
				log.Printf("创建API失败 [%s %s]: %v", api.Method, api.Path, err)
				continue
			}
			created++
			log.Printf("✓ 创建API: %s %s - %s", api.Method, api.Path, api.Description)
		} else {
			// API已存在,更新信息
			updates := map[string]interface{}{
				"group":       api.Group,
				"description": api.Description,
				"permission":  api.Permission,
				"is_public":   api.IsPublic,
				"status":      api.Status,
			}
			if err := db.Model(&existingAPI).Updates(updates).Error; err != nil {
				log.Printf("更新API失败 [%s %s]: %v", api.Method, api.Path, err)
				continue
			}
			updated++
			log.Printf("✓ 更新API: %s %s - %s", api.Method, api.Path, api.Description)
		}
	}

	// 输出统计信息
	fmt.Println("\n==================== 初始化完成 ====================")
	fmt.Printf("总计: %d 个API\n", len(apis))
	fmt.Printf("新创建: %d 个\n", created)
	fmt.Printf("已更新: %d 个\n", updated)
	fmt.Printf("跳过: %d 个\n", skipped)
	fmt.Println("===================================================")
}