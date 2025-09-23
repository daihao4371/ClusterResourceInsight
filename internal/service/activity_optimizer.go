package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/models"

	"gorm.io/gorm"
)

// ActivityOptimizer 活动优化器，负责去重和降噪
type ActivityOptimizer struct {
	db *gorm.DB
}

// OptimizationConfig 优化配置
type OptimizationConfig struct {
	DeduplicationWindow  int     `json:"deduplication_window"`   // 去重时间窗口（分钟）
	SimilarityThreshold  float64 `json:"similarity_threshold"`   // 相似度阈值
	MaxDuplicateCount    int     `json:"max_duplicate_count"`    // 最大重复次数
	EnableAggregation    bool    `json:"enable_aggregation"`     // 是否启用聚合
	AggregationThreshold int     `json:"aggregation_threshold"`  // 聚合阈值
	NoiseFilterEnabled   bool    `json:"noise_filter_enabled"`   // 是否启用噪音过滤
	AutoCleanupEnabled   bool    `json:"auto_cleanup_enabled"`   // 是否启用自动清理
	CleanupRetentionDays int     `json:"cleanup_retention_days"` // 清理保留天数
}

// ActivityAggregation 活动聚合信息
type ActivityAggregation struct {
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Count      int       `json:"count"`
	FirstTime  time.Time `json:"first_time"`
	LastTime   time.Time `json:"last_time"`
	ClusterIDs []uint    `json:"cluster_ids"`
	Sources    []string  `json:"sources"`
}

// OptimizationResult 优化结果
type OptimizationResult struct {
	DuplicatesRemoved int64                 `json:"duplicates_removed"`
	NoiseFiltered     int64                 `json:"noise_filtered"`
	Aggregations      []ActivityAggregation `json:"aggregations"`
	ProcessedAt       time.Time             `json:"processed_at"`
}

// NewActivityOptimizer 创建活动优化器实例
func NewActivityOptimizer() *ActivityOptimizer {
	return &ActivityOptimizer{
		db: database.GetDB(),
	}
}

// LoadOptimizationConfig 从系统配置加载优化配置
func (o *ActivityOptimizer) LoadOptimizationConfig() (*OptimizationConfig, error) {
	// 默认配置
	config := &OptimizationConfig{
		DeduplicationWindow:  30,
		SimilarityThreshold:  0.8,
		MaxDuplicateCount:    3,
		EnableAggregation:    true,
		AggregationThreshold: 5,
		NoiseFilterEnabled:   true,
		AutoCleanupEnabled:   true,
		CleanupRetentionDays: 7,
	}

	// 从数据库加载配置
	configKeys := []string{
		"activity_dedup_window",
		"activity_similarity_threshold",
		"activity_max_duplicates",
		"activity_enable_aggregation",
		"activity_aggregation_threshold",
		"activity_enable_noise_filter",
		"activity_auto_cleanup",
		"activity_retention_days",
	}

	var settings []models.SystemSettings
	err := o.db.Where("`key` IN ?", configKeys).Find(&settings).Error
	if err != nil {
		logger.Warn("加载活动优化配置失败，使用默认配置: %v", err)
		return config, nil
	}

	// 应用数据库中的配置值
	for _, setting := range settings {
		switch setting.Key {
		case "activity_dedup_window":
			if val, err := strconv.Atoi(setting.Value); err == nil && val > 0 {
				config.DeduplicationWindow = val
			}
		case "activity_similarity_threshold":
			if val, err := strconv.ParseFloat(setting.Value, 64); err == nil && val > 0 && val <= 1 {
				config.SimilarityThreshold = val
			}
		case "activity_max_duplicates":
			if val, err := strconv.Atoi(setting.Value); err == nil && val > 0 {
				config.MaxDuplicateCount = val
			}
		case "activity_enable_aggregation":
			if val, err := strconv.ParseBool(setting.Value); err == nil {
				config.EnableAggregation = val
			}
		case "activity_aggregation_threshold":
			if val, err := strconv.Atoi(setting.Value); err == nil && val > 0 {
				config.AggregationThreshold = val
			}
		case "activity_enable_noise_filter":
			if val, err := strconv.ParseBool(setting.Value); err == nil {
				config.NoiseFilterEnabled = val
			}
		case "activity_auto_cleanup":
			if val, err := strconv.ParseBool(setting.Value); err == nil {
				config.AutoCleanupEnabled = val
			}
		case "activity_retention_days":
			if val, err := strconv.Atoi(setting.Value); err == nil && val > 0 {
				config.CleanupRetentionDays = val
			}
		}
	}

	return config, nil
}

// SaveOptimizationConfig 保存优化配置到数据库
func (o *ActivityOptimizer) SaveOptimizationConfig(config *OptimizationConfig) error {
	configMappings := map[string]interface{}{
		"activity_dedup_window":          config.DeduplicationWindow,
		"activity_similarity_threshold":  config.SimilarityThreshold,
		"activity_max_duplicates":        config.MaxDuplicateCount,
		"activity_enable_aggregation":    config.EnableAggregation,
		"activity_aggregation_threshold": config.AggregationThreshold,
		"activity_enable_noise_filter":   config.NoiseFilterEnabled,
		"activity_auto_cleanup":          config.AutoCleanupEnabled,
		"activity_retention_days":        config.CleanupRetentionDays,
	}

	for key, value := range configMappings {
		valueStr := fmt.Sprintf("%v", value)
		valueType := "string"

		switch value.(type) {
		case int:
			valueType = "int"
		case float64:
			valueType = "float"
		case bool:
			valueType = "bool"
		}

		// 更新或插入配置
		setting := models.SystemSettings{
			Key:       key,
			Value:     valueStr,
			ValueType: valueType,
		}

		err := o.db.Where("`key` = ?", key).
			Assign(setting).
			FirstOrCreate(&setting).Error
		if err != nil {
			return fmt.Errorf("保存配置 %s 失败: %w", key, err)
		}
	}

	logger.Info("活动优化配置保存成功")
	return nil
}

// OptimizeActivities 执行活动优化
func (o *ActivityOptimizer) OptimizeActivities() (*OptimizationResult, error) {
	config, err := o.LoadOptimizationConfig()
	if err != nil {
		return nil, fmt.Errorf("加载优化配置失败: %w", err)
	}

	result := &OptimizationResult{
		ProcessedAt: time.Now(),
	}

	// 1. 去除重复活动
	if duplicatesRemoved, err := o.removeDuplicateActivities(config); err != nil {
		logger.Error("去除重复活动失败: %v", err)
	} else {
		result.DuplicatesRemoved = duplicatesRemoved
	}

	// 2. 过滤噪音活动
	if config.NoiseFilterEnabled {
		if noiseFiltered, err := o.filterNoiseActivities(config); err != nil {
			logger.Error("过滤噪音活动失败: %v", err)
		} else {
			result.NoiseFiltered = noiseFiltered
		}
	}

	// 3. 创建聚合活动
	if config.EnableAggregation {
		if aggregations, err := o.createAggregatedActivities(config); err != nil {
			logger.Error("创建聚合活动失败: %v", err)
		} else {
			result.Aggregations = aggregations
		}
	}

	// 4. 自动清理过期数据
	if config.AutoCleanupEnabled {
		if err := o.cleanupExpiredActivities(config.CleanupRetentionDays); err != nil {
			logger.Error("自动清理失败: %v", err)
		}
	}

	logger.Info("活动优化完成: 去重=%d, 降噪=%d, 聚合=%d",
		result.DuplicatesRemoved, result.NoiseFiltered, len(result.Aggregations))

	return result, nil
}

// removeDuplicateActivities 去除重复活动
func (o *ActivityOptimizer) removeDuplicateActivities(config *OptimizationConfig) (int64, error) {
	windowStart := time.Now().Add(-time.Duration(config.DeduplicationWindow) * time.Minute)

	// 查找在时间窗口内的重复活动组
	var duplicateGroups []struct {
		Type      string
		Title     string
		ClusterID uint
		Count     int64
	}

	err := o.db.Model(&models.SystemActivity{}).
		Select("type, title, cluster_id, COUNT(*) as count").
		Where("created_at > ?", windowStart).
		Group("type, title, cluster_id").
		Having("COUNT(*) > ?", config.MaxDuplicateCount).
		Find(&duplicateGroups).Error

	if err != nil {
		return 0, fmt.Errorf("查找重复活动失败: %w", err)
	}

	var totalRemoved int64

	// 对每个重复组保留最新的记录，删除旧的
	for _, group := range duplicateGroups {
		var activityIDs []uint
		err := o.db.Model(&models.SystemActivity{}).
			Select("id").
			Where("type = ? AND title = ? AND cluster_id = ? AND created_at > ?",
				group.Type, group.Title, group.ClusterID, windowStart).
			Order("created_at DESC").
			Offset(config.MaxDuplicateCount). // 跳过要保留的记录
			Pluck("id", &activityIDs).Error

		if err != nil {
			logger.Error("查询重复活动ID失败: %v", err)
			continue
		}

		if len(activityIDs) > 0 {
			result := o.db.Where("id IN ?", activityIDs).Delete(&models.SystemActivity{})
			if result.Error != nil {
				logger.Error("删除重复活动失败: %v", result.Error)
				continue
			}
			totalRemoved += result.RowsAffected
		}
	}

	return totalRemoved, nil
}

// filterNoiseActivities 过滤噪音活动
func (o *ActivityOptimizer) filterNoiseActivities(_ *OptimizationConfig) (int64, error) {
	var totalFiltered int64

	// 定义噪音过滤规则
	noiseRules := []struct {
		name      string
		condition func() *gorm.DB
	}{
		{
			name: "测试相关活动",
			condition: func() *gorm.DB {
				return o.db.Where("title LIKE ? OR message LIKE ? OR title LIKE ?",
					"%测试%", "%test%", "%demo%")
			},
		},
		{
			name: "过期的系统启动活动",
			condition: func() *gorm.DB {
				return o.db.Where("title = ? AND created_at < ?",
					"系统监控启动", time.Now().Add(-2*time.Hour))
			},
		},
		{
			name: "无效集群的连接活动",
			condition: func() *gorm.DB {
				return o.db.Where("cluster_id = 0 AND type IN ? AND title LIKE ?",
					[]string{"warning", "error"}, "%集群%连接%")
			},
		},
		{
			name: "空内容活动",
			condition: func() *gorm.DB {
				return o.db.Where("message = '' OR title = ''")
			},
		},
	}

	// 应用噪音过滤规则
	for _, rule := range noiseRules {
		result := rule.condition().Delete(&models.SystemActivity{})
		if result.Error != nil {
			logger.Error("应用噪音过滤规则失败 (%s): %v", rule.name, result.Error)
			continue
		}

		if result.RowsAffected > 0 {
			totalFiltered += result.RowsAffected
			logger.Info("噪音过滤: %s, 删除=%d条", rule.name, result.RowsAffected)
		}
	}

	return totalFiltered, nil
}

// createAggregatedActivities 创建聚合活动
func (o *ActivityOptimizer) createAggregatedActivities(config *OptimizationConfig) ([]ActivityAggregation, error) {
	windowStart := time.Now().Add(-24 * time.Hour) // 聚合过去24小时的数据

	// 查找需要聚合的活动组
	var aggregationCandidates []struct {
		Type      string
		Title     string
		Count     int64
		FirstTime time.Time
		LastTime  time.Time
	}

	err := o.db.Model(&models.SystemActivity{}).
		Select("type, title, COUNT(*) as count, MIN(created_at) as first_time, MAX(created_at) as last_time").
		Where("created_at > ? AND source != ?", windowStart, "aggregator"). // 排除已聚合的活动
		Group("type, title").
		Having("COUNT(*) >= ?", config.AggregationThreshold).
		Find(&aggregationCandidates).Error

	if err != nil {
		return nil, fmt.Errorf("查找聚合候选失败: %w", err)
	}

	var aggregations []ActivityAggregation

	for _, candidate := range aggregationCandidates {
		// 获取该类型活动涉及的集群和来源
		var clusterIDs []uint
		var sources []string

		o.db.Model(&models.SystemActivity{}).
			Where("type = ? AND title = ? AND created_at > ?",
				candidate.Type, candidate.Title, windowStart).
			Distinct("cluster_id").
			Pluck("cluster_id", &clusterIDs)

		o.db.Model(&models.SystemActivity{}).
			Where("type = ? AND title = ? AND created_at > ?",
				candidate.Type, candidate.Title, windowStart).
			Distinct("source").
			Pluck("source", &sources)

		aggregation := ActivityAggregation{
			Type:       candidate.Type,
			Title:      candidate.Title,
			Count:      int(candidate.Count),
			FirstTime:  candidate.FirstTime,
			LastTime:   candidate.LastTime,
			ClusterIDs: clusterIDs,
			Sources:    sources,
		}

		// 创建聚合活动记录
		if err := o.createAggregationRecord(aggregation); err != nil {
			logger.Error("创建聚合记录失败: %v", err)
			continue
		}

		// 删除被聚合的原始活动（可选）
		// o.db.Where("type = ? AND title = ? AND created_at > ? AND created_at < ?",
		//     candidate.Type, candidate.Title, windowStart, time.Now()).
		//     Delete(&models.SystemActivity{})

		aggregations = append(aggregations, aggregation)
	}

	return aggregations, nil
}

// createAggregationRecord 创建聚合记录
func (o *ActivityOptimizer) createAggregationRecord(agg ActivityAggregation) error {
	details := map[string]interface{}{
		"aggregated_count":    agg.Count,
		"time_range_hours":    agg.LastTime.Sub(agg.FirstTime).Hours(),
		"affected_clusters":   agg.ClusterIDs,
		"involved_sources":    agg.Sources,
		"aggregation_created": time.Now().Unix(),
	}

	detailsJSON, _ := json.Marshal(details)

	message := fmt.Sprintf("在过去%.1f小时内发生了%d次相似活动，涉及%d个集群",
		agg.LastTime.Sub(agg.FirstTime).Hours(), agg.Count, len(agg.ClusterIDs))

	activity := &models.SystemActivity{
		Type:      agg.Type,
		ClusterID: 0, // 聚合活动不关联特定集群
		Title:     fmt.Sprintf("[聚合] %s", agg.Title),
		Message:   message,
		Source:    "aggregator",
		Details:   string(detailsJSON),
		CreatedAt: time.Now(),
	}

	return o.db.Create(activity).Error
}

// cleanupExpiredActivities 清理过期活动
func (o *ActivityOptimizer) cleanupExpiredActivities(retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	result := o.db.Where("created_at < ?", cutoffTime).Delete(&models.SystemActivity{})
	if result.Error != nil {
		return fmt.Errorf("清理过期活动失败: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		logger.Info("清理过期活动: 删除=%d条, 保留天数=%d", result.RowsAffected, retentionDays)
	}

	return nil
}

// CheckBeforeRecord 记录前检查是否应该记录该活动
func (o *ActivityOptimizer) CheckBeforeRecord(activity *models.SystemActivity) (bool, error) {
	config, err := o.LoadOptimizationConfig()
	if err != nil {
		// 配置加载失败时允许记录
		return true, nil
	}

	// 检查是否为重复活动
	windowStart := time.Now().Add(-time.Duration(config.DeduplicationWindow) * time.Minute)

	var count int64
	err = o.db.Model(&models.SystemActivity{}).
		Where("type = ? AND title = ? AND cluster_id = ? AND created_at > ?",
			activity.Type, activity.Title, activity.ClusterID, windowStart).
		Count(&count).Error

	if err != nil {
		// 查询失败时允许记录
		return true, nil
	}

	// 如果已超过最大重复次数，则不记录
	shouldRecord := count < int64(config.MaxDuplicateCount)

	if !shouldRecord {
		logger.Debug("跳过重复活动: type=%s, title=%s, count=%d",
			activity.Type, activity.Title, count)
	}

	return shouldRecord, nil
}
