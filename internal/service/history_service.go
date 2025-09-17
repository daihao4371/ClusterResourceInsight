package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/models"

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

// QueryHistory 查询历史数据
func (hs *HistoryService) QueryHistory(req HistoryQueryRequest) (*HistoryQueryResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
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

	// 应用分页
	offset := (req.Page - 1) * req.Size
	query = query.Offset(offset).Limit(req.Size)

	// 执行查询
	var data []models.PodMetricsHistory
	if err := query.Preload("Cluster").Find(&data).Error; err != nil {
		return nil, fmt.Errorf("查询历史数据失败: %v", err)
	}

	// 计算总页数
	totalPages := int((total + int64(req.Size) - 1) / int64(req.Size))

	return &HistoryQueryResponse{
		Data:       data,
		Total:      total,
		Page:       req.Page,
		Size:       req.Size,
		TotalPages: totalPages,
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