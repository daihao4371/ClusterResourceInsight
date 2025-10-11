package api

import (
	"net/http"
	"strconv"

	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"

	"github.com/gin-gonic/gin"
)

// CreateApi 创建API
func CreateApi(apiService *service.ApiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.CreateApiRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}

		api, err := apiService.CreateApi(&req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, api)
	}
}

// GetApiByID 根据ID获取API
func GetApiByID(apiService *service.ApiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "API ID格式错误")
			return
		}

		api, err := apiService.GetApiByID(uint(id))
		if err != nil {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Success(c, api)
	}
}

// ListApis 获取API列表
func ListApis(apiService *service.ApiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

		// 解析筛选参数
		keyword := c.Query("keyword")
		group := c.Query("group")
		method := c.Query("method")
		status := c.Query("status")

		var isPublic *bool
		if isPublicStr := c.Query("is_public"); isPublicStr != "" {
			val := isPublicStr == "true"
			isPublic = &val
		}

		// 参数校验
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		apis, total, err := apiService.ListApis(page, pageSize, keyword, group, method, status, isPublic)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.SuccessWithData(c, gin.H{
			"list":      apis,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

// UpdateApi 更新API信息
func UpdateApi(apiService *service.ApiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "API ID格式错误")
			return
		}

		var req service.UpdateApiRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}

		api, err := apiService.UpdateApi(uint(id), &req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, api)
	}
}

// DeleteApi 删除API
func DeleteApi(apiService *service.ApiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "API ID格式错误")
			return
		}

		if err := apiService.DeleteApi(uint(id)); err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, gin.H{"message": "API删除成功"})
	}
}

// GetApiGroups 获取所有API分组
func GetApiGroups(apiService *service.ApiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		groups, err := apiService.GetApiGroups()
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		// 直接返回groups数组
		response.Success(c, groups)
	}
}

// BatchCreateApis 批量创建API
func BatchCreateApis(apiService *service.ApiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Apis []service.CreateApiRequest `json:"apis" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}

		if err := apiService.BatchCreateApis(req.Apis); err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, gin.H{"message": "批量创建API成功"})
	}
}

// SyncApis 同步API
func SyncApis(apiService *service.ApiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Apis []service.CreateApiRequest `json:"apis" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}

		if err := apiService.SyncApis(req.Apis); err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, gin.H{"message": "API同步成功"})
	}
}

// GetApiStatistics 获取API统计信息
func GetApiStatistics(apiService *service.ApiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := apiService.GetApiStatistics()
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, stats)
	}
}