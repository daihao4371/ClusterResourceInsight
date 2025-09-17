package api

import (
	"log"
	"net/http"
	"strconv"

	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.RouterGroup, resourceCollector *collector.ResourceCollector) {
	// 原有的分析接口
	r.GET("/analysis", getResourceAnalysis(resourceCollector))
	r.GET("/pods", getPodsData(resourceCollector))
	r.GET("/health", healthCheck)

	// 新增的集群管理接口
	clusterService := service.NewClusterService()
	
	// 集群管理路由组
	clusterGroup := r.Group("/clusters")
	{
		clusterGroup.POST("", createCluster(clusterService))                    // 创建集群
		clusterGroup.GET("", getAllClusters(clusterService))                   // 获取所有集群
		clusterGroup.GET("/:id", getClusterByID(clusterService))               // 根据ID获取集群
		clusterGroup.PUT("/:id", updateCluster(clusterService))                // 更新集群配置
		clusterGroup.DELETE("/:id", deleteCluster(clusterService))             // 删除集群
		clusterGroup.POST("/:id/test", testClusterConnection(clusterService))   // 测试集群连接
		clusterGroup.POST("/test", testClusterByConfig(clusterService))        // 测试集群配置（创建前验证）
		clusterGroup.POST("/batch-test", batchTestAllClusters(clusterService)) // 批量测试所有集群
	}
}

// 创建集群
func createCluster(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.CreateClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求参数格式错误: " + err.Error(),
			})
			return
		}

		cluster, err := clusterService.CreateCluster(&req)
		if err != nil {
			log.Printf("创建集群失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "集群创建成功",
			"data":    cluster,
		})
	}
}

// 获取所有集群
func getAllClusters(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clusters, err := clusterService.GetAllClusters()
		if err != nil {
			log.Printf("获取集群列表失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  clusters,
			"count": len(clusters),
		})
	}
}

// 根据ID获取集群
func getClusterByID(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的集群ID",
			})
			return
		}

		cluster, err := clusterService.GetClusterByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": cluster,
		})
	}
}

// 更新集群配置
func updateCluster(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的集群ID",
			})
			return
		}

		var req service.UpdateClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求参数格式错误: " + err.Error(),
			})
			return
		}

		cluster, err := clusterService.UpdateCluster(uint(id), &req)
		if err != nil {
			log.Printf("更新集群失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "集群更新成功",
			"data":    cluster,
		})
	}
}

// 删除集群
func deleteCluster(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的集群ID",
			})
			return
		}

		err = clusterService.DeleteCluster(uint(id))
		if err != nil {
			log.Printf("删除集群失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "集群删除成功",
		})
	}
}

// 测试集群连接
func testClusterConnection(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的集群ID",
			})
			return
		}

		result, err := clusterService.TestClusterConnection(uint(id))
		if err != nil {
			log.Printf("测试集群连接失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": result,
		})
	}
}

// 测试集群配置（创建前验证）
func testClusterByConfig(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.CreateClusterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求参数格式错误: " + err.Error(),
			})
			return
		}

		result, err := clusterService.TestClusterConnectionByConfig(&req)
		if err != nil {
			log.Printf("测试集群配置失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": result,
		})
	}
}

// 批量测试所有集群
func batchTestAllClusters(clusterService *service.ClusterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		results, err := clusterService.BatchTestAllClusters()
		if err != nil {
			log.Printf("批量测试集群失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  results,
			"count": len(results),
		})
	}
}

func getResourceAnalysis(resourceCollector *collector.ResourceCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Starting resource analysis...")
		result, err := resourceCollector.CollectAllPodsData(c.Request.Context())
		if err != nil {
			log.Printf("Error in resource analysis: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		log.Printf("Analysis complete: total_pods=%d, unreasonable_pods=%d", result.TotalPods, result.UnreasonablePods)
		c.JSON(http.StatusOK, result)
	}
}

func getPodsData(resourceCollector *collector.ResourceCollector) gin.HandlerFunc {
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// 根据参数返回不同的数据
		if onlyProblems == "true" {
			pods := result.Top50Problems
			if len(pods) > limit {
				pods = pods[:limit]
			}
			c.JSON(http.StatusOK, gin.H{
				"pods":        pods,
				"total_count": len(result.Top50Problems),
				"limit":       limit,
			})
		} else {
			// 返回所有数据，这里需要重新收集
			c.JSON(http.StatusOK, gin.H{
				"total_pods":        result.TotalPods,
				"unreasonable_pods": result.UnreasonablePods,
				"problems":          result.Top50Problems[:min(limit, len(result.Top50Problems))],
			})
		}
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "cluster-resource-insight",
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}