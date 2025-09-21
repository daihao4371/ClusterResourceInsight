package pagination

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

// DatabasePaginationHandler 数据库分页处理器 - 专用于service层的分页逻辑
type DatabasePaginationHandler struct{}

// NewDatabasePaginationHandler 创建数据库分页处理器实例
func NewDatabasePaginationHandler() *DatabasePaginationHandler {
	return &DatabasePaginationHandler{}
}

// ParsePaginationParams 解析分页参数并设置默认值
// 统一处理分页参数的验证和默认值设置
func (handler *DatabasePaginationHandler) ParsePaginationParams(page, size int, defaultSize int) PaginationParams {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = defaultSize
	}
	
	return PaginationParams{
		Page: page,
		Size: size,
	}
}

// CalculatePaginationOffset 计算分页偏移量和限制
// 返回用于数据库查询的offset和limit值
func (handler *DatabasePaginationHandler) CalculatePaginationOffset(params PaginationParams) (offset int, limit int) {
	offset = (params.Page - 1) * params.Size
	limit = params.Size
	return offset, limit
}

// BuildPaginationResult 构建分页结果
// 根据查询结果和总数构建完整的分页信息
func (handler *DatabasePaginationHandler) BuildPaginationResult(params PaginationParams, total int64) PaginationResult {
	totalPages := int((total + int64(params.Size) - 1) / int64(params.Size))
	
	return PaginationResult{
		Page:       params.Page,
		Size:       params.Size,
		Total:      int(total),
		TotalPages: totalPages,
		HasNext:    params.Page < totalPages,
		HasPrev:    params.Page > 1,
	}
}

// BuildQueryResponse 构建标准化的查询响应
// 组合数据和分页信息为统一的响应格式
func (handler *DatabasePaginationHandler) BuildQueryResponse(data interface{}, paginationResult PaginationResult) map[string]interface{} {
	return map[string]interface{}{
		"data":        data,
		"total":       paginationResult.Total,
		"page":        paginationResult.Page,
		"size":        paginationResult.Size,
		"total_pages": paginationResult.TotalPages,
	}
}