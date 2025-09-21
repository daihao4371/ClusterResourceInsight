package api

import (
	"strconv"

	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/pkg/utils"

	"github.com/gin-gonic/gin"
)

// GetPodsData 获取Pod数据 - 支持按限制数量和问题筛选获取Pod信息
func GetPodsData(resourceCollector *collector.ResourceCollector) gin.HandlerFunc {
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
			response.InternalServerError(err.Error(), c)
			return
		}

		// 根据参数返回不同的数据
		if onlyProblems == "true" {
			pods := result.Top50Problems
			if len(pods) > limit {
				pods = pods[:limit]
			}
			response.OkWithData(gin.H{
				"pods":        pods,
				"total_count": len(result.Top50Problems),
				"limit":       limit,
			}, c)
		} else {
			// 返回所有数据，这里需要重新收集
			response.OkWithData(gin.H{
				"total_pods":        result.TotalPods,
				"unreasonable_pods": result.UnreasonablePods,
				"problems":          result.Top50Problems[:utils.MinInt(limit, len(result.Top50Problems))],
			}, c)
		}
	}
}

// SearchPods 搜索Pod - 支持多条件搜索和分页的Pod查询功能
func SearchPods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req collector.PodSearchRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			response.BadRequest("请求参数格式错误: "+err.Error(), c)
			return
		}

		// 设置默认值
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.Size <= 0 {
			req.Size = 10
		}

		podsResponse, err := multiCollector.SearchPods(c.Request.Context(), req)
		if err != nil {
			logger.Error("搜索Pod失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(podsResponse, c)
	}
}

// ListPods 获取Pod列表 - 使用统一分页处理器获取Pod列表数据
func ListPods(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用统一的分页处理器解析分页参数
		paginationHandler := utils.NewHttpPaginationHandler()
		paginationParams := paginationHandler.ParsePaginationParams(c, 10)

		req := collector.PodSearchRequest{
			Page: paginationParams.Page,
			Size: paginationParams.Size,
		}

		podsResponse, err := multiCollector.SearchPods(c.Request.Context(), req)
		if err != nil {
			logger.Error("获取Pod列表失败: %v", err)
			response.InternalServerError(err.Error(), c)
			return
		}

		response.OkWithData(podsResponse, c)
	}
}

// GetProblemsWithPagination 获取问题Pod分页数据 - 支持筛选、排序和分页的问题Pod查询
func GetProblemsWithPagination(multiCollector *collector.MultiClusterResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用统一的分页处理器解析分页参数
		paginationHandler := utils.NewHttpPaginationHandler()
		paginationParams := paginationHandler.ParsePaginationParams(c, 10)

		clusterName := c.Query("cluster_name")
		sortBy := c.DefaultQuery("sort_by", "total_waste")

		// 获取所有问题Pod数据
		result, err := multiCollector.CollectAllClustersData(c.Request.Context())
		if err != nil {
			logger.Error("获取问题Pod数据失败: %v", err)
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

		// 应用排序
		podSorter := utils.NewPodSorter()
		podSorter.SortProblems(filteredProblems, sortBy)

		// 使用统一的分页处理器应用分页
		pagedProblems, paginationResult := paginationHandler.ApplyPaginationToSlice(filteredProblems, paginationParams)

		logger.Info("问题Pod分页查询完成: total=%d, page=%d, size=%d, cluster=%s, sort=%s",
			paginationResult.Total, paginationParams.Page, paginationParams.Size, clusterName, sortBy)

		// 使用统一的分页响应构建器
		responseData := paginationHandler.BuildPaginationResponse(paginationParams, paginationResult.Total, pagedProblems)
		responseData["cluster_name"] = clusterName
		responseData["sort_by"] = sortBy

		response.OkWithData(responseData, c)
	}
}
