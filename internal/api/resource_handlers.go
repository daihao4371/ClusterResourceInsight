package api

import (
	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/pkg/utils"

	"github.com/gin-gonic/gin"
)

// GetResourceAnalysis 获取资源分析数据 - 分析Pod资源配置合理性并支持分页和集群筛选
func GetResourceAnalysis(resourceCollector *collector.ResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Starting resource analysis...")

		// 使用统一的分页处理器解析分页参数
		paginationHandler := utils.NewHttpPaginationHandler()
		paginationParams := paginationHandler.ParsePaginationParams(c, 50)

		clusterName := c.Query("cluster_name")

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
		pagedProblems, total, _, _ := collector.ApplyPagination(filteredProblems, paginationParams.Page, paginationParams.Size)

		// 构造分页响应
		pagedResult := &collector.AnalysisResult{
			TotalPods:        result.TotalPods,
			UnreasonablePods: result.UnreasonablePods,
			Top50Problems:    pagedProblems,
			GeneratedAt:      result.GeneratedAt,
			ClustersAnalyzed: result.ClustersAnalyzed,
		}

		logger.Info("Analysis complete: total_pods=%d, unreasonable_pods=%d, page=%d, size=%d, cluster=%s",
			result.TotalPods, result.UnreasonablePods, paginationParams.Page, paginationParams.Size, clusterName)

		// 使用统一的分页响应构建器
		responseData := paginationHandler.BuildPaginationResponse(paginationParams, total, pagedResult)
		responseData["filter"] = gin.H{
			"cluster_name": clusterName,
		}

		response.OkWithData(responseData, c)
	}
}
