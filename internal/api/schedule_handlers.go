package api

import (
	"fmt"
	"strconv"

	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"

	"github.com/gin-gonic/gin"
)

// GetScheduleStatus 获取调度服务状态 - 查询当前调度服务的运行状态和配置信息
func GetScheduleStatus(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := scheduleService.GetStatus()
		response.OkWithData(gin.H{
			"data": status,
		}, c)
	}
}

// StartScheduleService 启动调度服务 - 启动后台定时任务调度系统
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

// StopScheduleService 停止调度服务 - 停止后台定时任务调度系统
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

// GetScheduleJobs 获取调度任务列表 - 查看所有已配置的定时任务信息
func GetScheduleJobs(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobs := scheduleService.GetAllJobs()
		response.OkWithData(gin.H{
			"data":  jobs,
			"count": len(jobs),
		}, c)
	}
}

// RestartClusterJob 重启集群任务 - 重启指定集群的定时收集任务
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

// GetScheduleSettings 获取调度设置 - 获取全局调度服务的配置参数
func GetScheduleSettings(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := scheduleService.GetStatus()
		settings := status["global_settings"]
		response.OkWithData(gin.H{
			"data": settings,
		}, c)
	}
}

// StartClusterJob 启动集群任务 - 启动指定集群的定时收集任务
func StartClusterJob(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterIDStr := c.Param("cluster_id")
		clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的集群ID", c)
			return
		}

		// 重启任务实际上会先停止再启动
		err = scheduleService.RestartJob(uint(clusterID))
		if err != nil {
			logger.Error("启动集群任务失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("集群任务启动成功", c)
	}
}

// StopClusterJob 停止集群任务 - 停止指定集群的定时收集任务
func StopClusterJob(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterIDStr := c.Param("cluster_id")
		clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的集群ID", c)
			return
		}

		// 需要在ScheduleService中添加停止单个任务的方法
		err = scheduleService.StopJob(uint(clusterID))
		if err != nil {
			logger.Error("停止集群任务失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("集群任务停止成功", c)
	}
}

// ExecuteJobNow 立即执行集群任务 - 触发指定集群的即时数据收集
func ExecuteJobNow(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterIDStr := c.Param("cluster_id")
		clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的集群ID", c)
			return
		}

		err = scheduleService.ExecuteJobNow(uint(clusterID))
		if err != nil {
			logger.Error("立即执行任务失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("任务已触发执行", c)
	}
}

// UpdateJobSettings 更新集群任务设置 - 修改指定集群的调度任务配置
func UpdateJobSettings(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusterIDStr := c.Param("cluster_id")
		clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
		if err != nil {
			response.BadRequest("无效的集群ID", c)
			return
		}

		var settings service.JobSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			response.BadRequest("请求参数格式错误: "+err.Error(), c)
			return
		}

		err = scheduleService.UpdateJobSettings(uint(clusterID), &settings)
		if err != nil {
			logger.Error("更新任务设置失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithMessage("任务设置更新成功", c)
	}
}

// UpdateScheduleSettings 更新调度设置 - 更新全局调度服务的配置参数
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

// RefreshJobsFromClusters 从集群配置刷新调度任务 - 手动同步集群配置与调度任务
func RefreshJobsFromClusters(scheduleService *service.ScheduleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := scheduleService.RefreshJobsFromClusters()
		if err != nil {
			logger.Error("刷新集群任务失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		// 返回刷新后的任务列表
		jobs := scheduleService.GetAllJobs()
		response.OkWithData(gin.H{
			"data":         jobs,
			"count":        len(jobs),
			"sync_message": fmt.Sprintf("成功同步 %d 个集群任务", len(jobs)),
		}, c)
	}
}
