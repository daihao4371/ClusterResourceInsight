package api

import (
	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetTopResourceNamespaces 获取资源使用最高的命名空间 - 按资源使用量排序返回
func GetTopResourceNamespaces(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取查询参数
		limitStr := c.DefaultQuery("limit", "10")
		sortBy := c.DefaultQuery("sort_by", "combined") // memory, cpu, combined
		
		limit := 10
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}

		summaries, err := multiCollector.GetTopResourceNamespaces(c.Request.Context(), limit, sortBy)
		if err != nil {
			logger.Error("获取资源使用最高的命名空间失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(gin.H{
			"data":    summaries,
			"count":   len(summaries),
			"limit":   limit,
			"sort_by": sortBy,
		}, c)
	}
}

// GetAllNamespaces 获取所有命名空间列表 - 提取去重后的命名空间名称列表
func GetAllNamespaces(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取所有namespace数据（不限制数量）
		summaries, err := multiCollector.GetTopResourceNamespaces(c.Request.Context(), -1, "combined")
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

// GetNamespacePods 获取指定命名空间的Pod列表 - 查询特定命名空间下的所有Pod信息
func GetNamespacePods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
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

// GetNamespaceTreeData 获取命名空间树状数据 - 获取指定命名空间的层级结构数据
func GetNamespaceTreeData(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
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
