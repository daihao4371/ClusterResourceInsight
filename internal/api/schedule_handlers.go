package api

import (
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
