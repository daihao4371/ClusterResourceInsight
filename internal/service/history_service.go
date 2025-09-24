package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/models"
	"cluster-resource-insight/pkg/pagination"

	"gorm.io/gorm"
)

// PodResourceInfo 简化的Pod资源信息（避免循环导入）
type PodResourceInfo struct {
	PodName        string    `json:"pod_name"`
	Namespace      string    `json:"namespace"`
	NodeName       string    `json:"node_name"`
	ClusterName    string    `json:"cluster_name"`
	MemoryUsage    int64     `json:"memory_usage"`
	MemoryRequest  int64     `json:"memory_request"`
	MemoryLimit    int64     `json:"memory_limit"`
	MemoryReqPct   float64   `json:"memory_req_pct"`
	MemoryLimitPct float64   `json:"memory_limit_pct"`
	CPUUsage       int64     `json:"cpu_usage"`
	CPURequest     int64     `json:"cpu_request"`
	CPULimit       int64     `json:"cpu_limit"`
	CPUReqPct      float64   `json:"cpu_req_pct"`
	CPULimitPct    float64   `json:"cpu_limit_pct"`
	Status         string    `json:"status"`
	Issues         []string  `json:"issues"`
	CreationTime   time.Time `json:"creation_time"`
}

// HistoryService 历史数据服务
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService 创建历史数据服务实例
func NewHistoryService() *HistoryService {
	return &HistoryService{
		db: database.GetDB(),
	}
}

// HistoryQueryRequest 历史数据查询请求
type HistoryQueryRequest struct {
	ClusterID   uint      `form:"cluster_id"`   // 集群ID筛选
	Namespace   string    `form:"namespace"`    // 命名空间筛选
	PodName     string    `form:"pod_name"`     // Pod名称筛选
	StartTime   time.Time `form:"start_time"`   // 开始时间
	EndTime     time.Time `form:"end_time"`     // 结束时间
	Page        int       `form:"page"`         // 页码
	Size        int       `form:"size"`         // 每页大小
	OrderBy     string    `form:"order_by"`     // 排序字段
	OrderDesc   bool      `form:"order_desc"`   // 是否降序
}

// HistoryQueryResponse 历史数据查询响应
type HistoryQueryResponse struct {
	Data       []models.PodMetricsHistory `json:"data"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	Size       int                        `json:"size"`
	TotalPages int                        `json:"total_pages"`
}

// SavePodMetrics 批量保存Pod监控数据
func (hs *HistoryService) SavePodMetrics(clusterID uint, pods []PodResourceInfo) error {
	if len(pods) == 0 {
		return nil
	}

	// 转换为数据库模型
	var historyRecords []models.PodMetricsHistory
	collectedAt := time.Now()

	for _, pod := range pods {
		// 序列化问题列表为JSON
		issuesJSON, _ := json.Marshal(pod.Issues)

		record := models.PodMetricsHistory{
			ClusterID:      clusterID,
			Namespace:      pod.Namespace,
			PodName:        pod.PodName,
			NodeName:       pod.NodeName,
			MemoryUsage:    pod.MemoryUsage,
			MemoryRequest:  pod.MemoryRequest,
			MemoryLimit:    pod.MemoryLimit,
			MemoryReqPct:   pod.MemoryReqPct,
			MemoryLimitPct: pod.MemoryLimitPct,
			CPUUsage:       pod.CPUUsage,
			CPURequest:     pod.CPURequest,
			CPULimit:       pod.CPULimit,
			CPUReqPct:      pod.CPUReqPct,
			CPULimitPct:    pod.CPULimitPct,
			Status:         pod.Status,
			Issues:         string(issuesJSON),
			CollectedAt:    collectedAt,
		}

		historyRecords = append(historyRecords, record)
	}

	// 批量插入数据
	if err := hs.db.CreateInBatches(historyRecords, 100).Error; err != nil {
		return fmt.Errorf("保存Pod监控历史数据失败: %v", err)
	}

	return nil
}

// QueryHistory 查询历史数据 - 使用统一分页逻辑
func (hs *HistoryService) QueryHistory(req HistoryQueryRequest) (*HistoryQueryResponse, error) {
	// 使用统一的分页处理器
	paginationHandler := pagination.NewDatabasePaginationHandler()
	paginationParams := paginationHandler.ParsePaginationParams(req.Page, req.Size, 20)
	
	// 设置默认排序
	if req.OrderBy == "" {
		req.OrderBy = "collected_at"
		req.OrderDesc = true
	}

	// 构建查询条件
	query := hs.db.Model(&models.PodMetricsHistory{})

	// 应用筛选条件
	if req.ClusterID > 0 {
		query = query.Where("cluster_id = ?", req.ClusterID)
	}
	if req.Namespace != "" {
		query = query.Where("namespace = ?", req.Namespace)
	}
	if req.PodName != "" {
		query = query.Where("pod_name LIKE ?", "%"+req.PodName+"%")
	}
	if !req.StartTime.IsZero() {
		query = query.Where("collected_at >= ?", req.StartTime)
	}
	if !req.EndTime.IsZero() {
		query = query.Where("collected_at <= ?", req.EndTime)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询历史数据总数失败: %v", err)
	}

	// 应用排序
	orderClause := req.OrderBy
	if req.OrderDesc {
		orderClause += " DESC"
	}
	query = query.Order(orderClause)

	// 使用统一分页逻辑计算偏移量
	offset, limit := paginationHandler.CalculatePaginationOffset(paginationParams)
	query = query.Offset(offset).Limit(limit)

	// 执行查询
	var data []models.PodMetricsHistory
	if err := query.Preload("Cluster").Find(&data).Error; err != nil {
		return nil, fmt.Errorf("查询历史数据失败: %v", err)
	}

	// 构建分页结果
	paginationResult := paginationHandler.BuildPaginationResult(paginationParams, total)

	return &HistoryQueryResponse{
		Data:       data,
		Total:      total,
		Page:       paginationResult.Page,
		Size:       paginationResult.Size,
		TotalPages: paginationResult.TotalPages,
	}, nil
}

// GetTrendData 获取趋势数据
func (hs *HistoryService) GetTrendData(clusterID uint, namespace, podName string, hours int) ([]models.PodMetricsHistory, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	query := hs.db.Model(&models.PodMetricsHistory{}).
		Where("collected_at >= ?", startTime).
		Order("collected_at ASC")

	if clusterID > 0 {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}
	if podName != "" {
		query = query.Where("pod_name = ?", podName)
	}

	var data []models.PodMetricsHistory
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("查询趋势数据失败: %v", err)
	}

	return data, nil
}

// CleanupOldData 清理过期数据
func (hs *HistoryService) CleanupOldData(ctx context.Context, retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	result := hs.db.Where("collected_at < ?", cutoffTime).Delete(&models.PodMetricsHistory{})
	if result.Error != nil {
		return fmt.Errorf("清理过期数据失败: %v", result.Error)
	}

	if result.RowsAffected > 0 {
		fmt.Printf("清理了 %d 条过期历史记录（超过 %d 天）\n", result.RowsAffected, retentionDays)
	}

	return nil
}

// GetStatistics 获取统计信息
func (hs *HistoryService) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总记录数
	var totalRecords int64
	if err := hs.db.Model(&models.PodMetricsHistory{}).Count(&totalRecords).Error; err != nil {
		return nil, fmt.Errorf("查询总记录数失败: %v", err)
	}
	stats["total_records"] = totalRecords

	// 集群数量
	var clusterCount int64
	if err := hs.db.Model(&models.PodMetricsHistory{}).Distinct("cluster_id").Count(&clusterCount).Error; err != nil {
		return nil, fmt.Errorf("查询集群数量失败: %v", err)
	}
	stats["cluster_count"] = clusterCount

	// 命名空间数量
	var namespaceCount int64
	if err := hs.db.Model(&models.PodMetricsHistory{}).Distinct("namespace").Count(&namespaceCount).Error; err != nil {
		return nil, fmt.Errorf("查询命名空间数量失败: %v", err)
	}
	stats["namespace_count"] = namespaceCount

	// 最早和最新的记录时间
	var earliestTime, latestTime time.Time
	if err := hs.db.Model(&models.PodMetricsHistory{}).Select("MIN(collected_at)").Scan(&earliestTime).Error; err == nil {
		stats["earliest_record"] = earliestTime
	}
	if err := hs.db.Model(&models.PodMetricsHistory{}).Select("MAX(collected_at)").Scan(&latestTime).Error; err == nil {
		stats["latest_record"] = latestTime
	}

	return stats, nil
}

// SystemTrendData 系统级趋势数据结构
type SystemTrendData struct {
	Time       string  `json:"time"`
	CPUUsage   float64 `json:"cpu"`
	MemoryUsage float64 `json:"memory"`
	PodCount   int     `json:"pods"`
}

// GetSystemTrendData 获取系统级聚合趋势数据 - 为Dashboard提供图表数据
// 聚合所有集群的CPU、内存使用率平均值和Pod总数按时间分组
func (hs *HistoryService) GetSystemTrendData(hours int) ([]SystemTrendData, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	// 计算时间分组间隔 - 根据时间范围动态调整
	var intervalMinutes int
	switch {
	case hours <= 1:
		intervalMinutes = 5  // 1小时内，5分钟一个点
	case hours <= 6:
		intervalMinutes = 15 // 6小时内，15分钟一个点
	case hours <= 24:
		intervalMinutes = 60 // 24小时内，1小时一个点
	default:
		intervalMinutes = 240 // 7天内，4小时一个点
	}
	
	// 构建SQL查询 - 按时间区间分组聚合数据
	query := `
		SELECT 
			DATE_FORMAT(MIN(collected_at), '%H:%i') as time_label,
			AVG(cpu_req_pct) as avg_cpu_usage,
			AVG(memory_req_pct) as avg_memory_usage,
			COUNT(DISTINCT pod_name) as pod_count,
			MIN(collected_at) as collected_at
		FROM pod_metrics_history 
		WHERE collected_at >= ? 
		GROUP BY 
			FLOOR(UNIX_TIMESTAMP(collected_at) / (? * 60))
		ORDER BY collected_at ASC
	`
	
	type QueryResult struct {
		TimeLabel      string    `json:"time_label"`
		AvgCPUUsage    float64   `json:"avg_cpu_usage"`
		AvgMemoryUsage float64   `json:"avg_memory_usage"`
		PodCount       int       `json:"pod_count"`
		CollectedAt    time.Time `json:"collected_at"`
	}
	
	var results []QueryResult
	if err := hs.db.Raw(query, startTime, intervalMinutes).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("查询系统趋势数据失败: %v", err)
	}
	
	// 转换为前端期望的格式
	var trendData []SystemTrendData
	for _, result := range results {
		trendData = append(trendData, SystemTrendData{
			Time:        result.TimeLabel,
			CPUUsage:    result.AvgCPUUsage,
			MemoryUsage: result.AvgMemoryUsage,
			PodCount:    result.PodCount,
		})
	}
	
	// 如果没有历史数据，返回空数组
	if len(trendData) == 0 {
		return []SystemTrendData{}, nil
	}
	
	return trendData, nil
}

// GetSystemTrendDataWithCluster 获取系统级聚合趋势数据，支持集群筛选 - 为Dashboard提供图表数据
// 聚合指定集群或所有集群的CPU、内存使用率平均值和Pod总数按时间分组
// 参数:
//   - hours: 时间范围（小时）
//   - clusterID: 集群ID，为nil时查询所有集群
// 返回:
//   - []SystemTrendData: 系统趋势数据数组
//   - error: 查询过程中的错误
func (hs *HistoryService) GetSystemTrendDataWithCluster(hours int, clusterID *uint) ([]SystemTrendData, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	// 计算时间分组间隔 - 根据时间范围动态调整
	var intervalMinutes int
	switch {
	case hours <= 1:
		intervalMinutes = 5  // 1小时内，5分钟一个点
	case hours <= 6:
		intervalMinutes = 15 // 6小时内，15分钟一个点
	case hours <= 24:
		intervalMinutes = 60 // 24小时内，1小时一个点
	default:
		intervalMinutes = 240 // 7天内，4小时一个点
	}
	
	// 构建SQL查询 - 按时间区间分组聚合数据，支持集群筛选
	var query string
	var queryParams []interface{}
	
	if clusterID != nil {
		// 筛选特定集群的数据
		query = `
			SELECT 
				DATE_FORMAT(MIN(collected_at), '%H:%i') as time_label,
				AVG(cpu_req_pct) as avg_cpu_usage,
				AVG(memory_req_pct) as avg_memory_usage,
				COUNT(DISTINCT pod_name) as pod_count,
				MIN(collected_at) as collected_at
			FROM pod_metrics_history 
			WHERE collected_at >= ? AND cluster_id = ?
			GROUP BY 
				FLOOR(UNIX_TIMESTAMP(collected_at) / (? * 60))
			ORDER BY collected_at ASC
		`
		queryParams = []interface{}{startTime, *clusterID, intervalMinutes}
	} else {
		// 聚合所有集群的数据
		query = `
			SELECT 
				DATE_FORMAT(MIN(collected_at), '%H:%i') as time_label,
				AVG(cpu_req_pct) as avg_cpu_usage,
				AVG(memory_req_pct) as avg_memory_usage,
				COUNT(DISTINCT pod_name) as pod_count,
				MIN(collected_at) as collected_at
			FROM pod_metrics_history 
			WHERE collected_at >= ? 
			GROUP BY 
				FLOOR(UNIX_TIMESTAMP(collected_at) / (? * 60))
			ORDER BY collected_at ASC
		`
		queryParams = []interface{}{startTime, intervalMinutes}
	}
	
	type QueryResult struct {
		TimeLabel      string    `json:"time_label"`
		AvgCPUUsage    float64   `json:"avg_cpu_usage"`
		AvgMemoryUsage float64   `json:"avg_memory_usage"`
		PodCount       int       `json:"pod_count"`
		CollectedAt    time.Time `json:"collected_at"`
	}
	
	var results []QueryResult
	if err := hs.db.Raw(query, queryParams...).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("查询系统趋势数据失败: %v", err)
	}
	
	// 转换为前端期望的格式
	var trendData []SystemTrendData
	for _, result := range results {
		trendData = append(trendData, SystemTrendData{
			Time:        result.TimeLabel,
			CPUUsage:    result.AvgCPUUsage,
			MemoryUsage: result.AvgMemoryUsage,
			PodCount:    result.PodCount,
		})
	}
	
	// 如果没有历史数据，返回空数组
	if len(trendData) == 0 {
		return []SystemTrendData{}, nil
	}
	
	return trendData, nil
}



// GetLatestPodMetrics 获取最新的Pod指标数据 - 按Pod聚合获取最近时间范围内每个Pod的最新记录
// 参数:
//   - clusterID: 集群ID
//   - namespace: 命名空间筛选条件，为空时查询所有命名空间
//   - duration: 时间范围，查询这个时间段内的最新数据
//
// 返回:
//   - []models.PodMetricsHistory: 最新的Pod指标数据列表
//   - error: 查询过程中的错误信息
func (hs *HistoryService) GetLatestPodMetrics(clusterID uint, namespace string, duration time.Duration) ([]models.PodMetricsHistory, error) {
	startTime := time.Now().Add(-duration)
	
	// 使用更简单的查询方式：直接获取指定时间范围内的最新记录
	// 按Pod分组并获取每组的最新记录
	query := hs.db.Model(&models.PodMetricsHistory{}).
		Where("collected_at >= ? AND cluster_id = ?", startTime, clusterID)
	
	// 添加命名空间筛选条件
	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}
	
	// 获取所有符合条件的记录，然后在应用层去重
	var allPods []models.PodMetricsHistory
	err := query.Order("collected_at DESC").Find(&allPods).Error
	if err != nil {
		return nil, fmt.Errorf("查询Pod指标数据失败: %v", err)
	}
	
	// 在应用层按Pod名称去重，保留最新记录
	podMap := make(map[string]models.PodMetricsHistory)
	for _, pod := range allPods {
		key := fmt.Sprintf("%d_%s_%s", pod.ClusterID, pod.Namespace, pod.PodName)
		if existing, exists := podMap[key]; !exists || pod.CollectedAt.After(existing.CollectedAt) {
			podMap[key] = pod
		}
	}
	
	// 转换为切片返回
	var latestPods []models.PodMetricsHistory
	for _, pod := range podMap {
		latestPods = append(latestPods, pod)
	}
	
	return latestPods, nil
}
