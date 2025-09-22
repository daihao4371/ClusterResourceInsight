package api

import (
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
