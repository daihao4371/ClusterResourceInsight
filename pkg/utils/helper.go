package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams 分页参数结构
type PaginationParams struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

// PaginationResult 分页结果结构
type PaginationResult struct {
	Page       int  `json:"page"`
	Size       int  `json:"size"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// HttpPaginationHandler HTTP分页处理器
type HttpPaginationHandler struct{}

// NewHttpPaginationHandler 创建HTTP分页处理器实例
func NewHttpPaginationHandler() *HttpPaginationHandler {
	return &HttpPaginationHandler{}
}

// ParsePaginationParams 解析HTTP请求中的分页参数
// 从gin.Context中提取page和size参数，并设置默认值和验证
func (handler *HttpPaginationHandler) ParsePaginationParams(c *gin.Context, defaultSize int) PaginationParams {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", strconv.Itoa(defaultSize))

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		size = defaultSize
	}

	// 限制最大分页大小，避免性能问题
	maxSize := 100
	if size > maxSize {
		size = maxSize
	}

	return PaginationParams{
		Page: page,
		Size: size,
	}
}

// BuildPaginationResponse 构建标准分页响应
// 根据分页参数和数据总数构建统一的分页响应格式
func (handler *HttpPaginationHandler) BuildPaginationResponse(params PaginationParams, total int, data interface{}) gin.H {
	totalPages := (total + params.Size - 1) / params.Size
	start := (params.Page - 1) * params.Size
	end := start + params.Size

	return gin.H{
		"data": data,
		"pagination": gin.H{
			"page":        params.Page,
			"size":        params.Size,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    end < total,
			"has_prev":    params.Page > 1,
		},
	}
}

// GormPaginationHandler GORM数据库分页处理器
type GormPaginationHandler struct{}

// NewGormPaginationHandler 创建GORM分页处理器实例
func NewGormPaginationHandler() *GormPaginationHandler {
	return &GormPaginationHandler{}
}

// ApplyPaginationToQuery 对GORM查询应用分页逻辑
// 自动处理计数查询、偏移量计算和结果构建
func (handler *GormPaginationHandler) ApplyPaginationToQuery(params PaginationParams) (offset int, limit int, result PaginationResult) {
	// 计算偏移量和限制
	offset = (params.Page - 1) * params.Size
	limit = params.Size

	// 构建分页结果（总数需要外部设置）
	result = PaginationResult{
		Page:    params.Page,
		Size:    params.Size,
		Total:   0, // 将由调用方设置
		HasPrev: params.Page > 1,
	}

	return offset, limit, result
}

// BuildQueryPaginationResponse 构建查询分页响应 - 标准化数据库查询分页结果格式
func (handler *GormPaginationHandler) BuildQueryPaginationResponse(params PaginationParams, total int64, data interface{}) map[string]interface{} {
	totalPages := int((total + int64(params.Size) - 1) / int64(params.Size))

	return map[string]interface{}{
		"data":        data,
		"total":       total,
		"page":        params.Page,
		"size":        params.Size,
		"total_pages": totalPages,
	}
}

// MinInt 返回两个整数中的较小值
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}