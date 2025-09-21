package utils

import (
	"cluster-resource-insight/internal/collector"
	"math"
	"sort"
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

// ApplyPaginationToSlice 对切片数据应用分页
// 通用的切片分页处理方法，返回分页后的数据和分页信息
func (handler *HttpPaginationHandler) ApplyPaginationToSlice(data []collector.PodResourceInfo, params PaginationParams) ([]collector.PodResourceInfo, PaginationResult) {
	total := len(data)
	totalPages := (total + params.Size - 1) / params.Size
	start := (params.Page - 1) * params.Size
	end := start + params.Size

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var pagedData []collector.PodResourceInfo
	if start < end {
		pagedData = data[start:end]
	}

	result := PaginationResult{
		Page:       params.Page,
		Size:       params.Size,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    end < total,
		HasPrev:    params.Page > 1,
	}

	return pagedData, result
}

// PodSorter Pod排序器
type PodSorter struct{}

// NewPodSorter 创建Pod排序器实例
func NewPodSorter() *PodSorter {
	return &PodSorter{}
}

// SortProblems 对问题Pod进行排序
// 支持按CPU浪费、内存浪费、总浪费程度排序
func (sorter *PodSorter) SortProblems(problems []collector.PodResourceInfo, sortBy string) {
	switch sortBy {
	case "cpu_waste":
		sort.Slice(problems, func(i, j int) bool {
			cpuWasteI := sorter.calculateCPUWaste(problems[i])
			cpuWasteJ := sorter.calculateCPUWaste(problems[j])
			return cpuWasteI > cpuWasteJ
		})
	case "memory_waste":
		sort.Slice(problems, func(i, j int) bool {
			memoryWasteI := sorter.calculateMemoryWaste(problems[i])
			memoryWasteJ := sorter.calculateMemoryWaste(problems[j])
			return memoryWasteI > memoryWasteJ
		})
	case "total_waste":
		fallthrough
	default:
		sort.Slice(problems, func(i, j int) bool {
			totalWasteI := sorter.calculateTotalWaste(problems[i])
			totalWasteJ := sorter.calculateTotalWaste(problems[j])
			return totalWasteI > totalWasteJ
		})
	}
}

// calculateCPUWaste 计算CPU浪费程度
func (sorter *PodSorter) calculateCPUWaste(pod collector.PodResourceInfo) float64 {
	if pod.CPUReqPct <= 0 {
		return 0
	}
	return math.Max(0, 100-pod.CPUReqPct)
}

// calculateMemoryWaste 计算内存浪费程度
func (sorter *PodSorter) calculateMemoryWaste(pod collector.PodResourceInfo) float64 {
	if pod.MemoryReqPct <= 0 {
		return 0
	}
	return math.Max(0, 100-pod.MemoryReqPct)
}

// calculateTotalWaste 计算总浪费程度
func (sorter *PodSorter) calculateTotalWaste(pod collector.PodResourceInfo) float64 {
	cpuWaste := sorter.calculateCPUWaste(pod)
	memoryWaste := sorter.calculateMemoryWaste(pod)
	return (cpuWaste + memoryWaste) / 2
}

// PaginationHelper 分页辅助器（保留向后兼容）
type PaginationHelper struct{}

// NewPaginationHelper 创建分页辅助器实例
func NewPaginationHelper() *PaginationHelper {
	return &PaginationHelper{}
}

// ApplyPagination 应用分页逻辑
// 根据页码和每页大小对数据进行分页处理
func (helper *PaginationHelper) ApplyPagination(data []collector.PodResourceInfo, page, size int) ([]collector.PodResourceInfo, int, int, int) {
	total := len(data)
	totalPages := (total + size - 1) / size
	start := (page - 1) * size
	end := start + size

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	var pagedData []collector.PodResourceInfo
	if start < end {
		pagedData = data[start:end]
	}

	return pagedData, total, totalPages, end
}

// MinInt 返回两个整数中的较小值
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}