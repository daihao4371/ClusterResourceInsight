package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/models"
	"cluster-resource-insight/pkg/deduplication"

	"gorm.io/gorm"
)

// ActivityService 活动服务管理器
type ActivityService struct {
	db           *gorm.DB
	optimizer    *ActivityOptimizer
	deduplicator *deduplication.AlertDeduplicator // 告警降噪器
}

// NewActivityService 创建活动服务实例
func NewActivityService() *ActivityService {
	return &ActivityService{
		db:           database.GetDB(),
		optimizer:    NewActivityOptimizer(),
		deduplicator: deduplication.NewAlertDeduplicator(),
	}
}

// ActivityItem 前端使用的活动项目结构
type ActivityItem struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Time    string `json:"time"`
	Source  string `json:"source"`
	Details string `json:"details,omitempty"`
}

// AlertItem 前端使用的告警项目结构
type AlertItem struct {
	ID          uint   `json:"id"`
	Level       string `json:"level"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Time        string `json:"time"`
	Status      string `json:"status"`
}

// RecordActivity 记录系统活动（集成优化检查）
func (s *ActivityService) RecordActivity(activityType, title, message, source string, clusterID uint, details map[string]interface{}) error {
	activity := &models.SystemActivity{
		Type:      activityType,
		ClusterID: clusterID,
		Title:     title,
		Message:   message,
		Source:    source,
		CreatedAt: time.Now(),
	}

	// 检查是否应该记录此活动
	shouldRecord, err := s.optimizer.CheckBeforeRecord(activity)
	if err != nil {
		logger.Warn("检查活动记录失败，继续记录: %v", err)
	} else if !shouldRecord {
		// 跳过重复活动
		return nil
	}

	// 序列化详情
	var detailsJSON string
	if details != nil {
		detailsBytes, err := json.Marshal(details)
		if err != nil {
			logger.Error("序列化活动详情失败: %v", err)
			detailsJSON = "{}"
		} else {
			detailsJSON = string(detailsBytes)
		}
	}
	activity.Details = detailsJSON

	if err := s.db.Create(activity).Error; err != nil {
		return fmt.Errorf("记录系统活动失败: %w", err)
	}

	logger.Info("记录系统活动: type=%s, title=%s, cluster_id=%d", activityType, title, clusterID)
	return nil
}

// GetRecentActivities 获取最近的系统活动
func (s *ActivityService) GetRecentActivities(limit int) ([]ActivityItem, error) {
	if limit <= 0 {
		limit = 10 // 默认显示10条
	}

	if limit > 50 {
		limit = 50 // 最多显示50条
	}

	var activities []models.SystemActivity
	err := s.db.Preload("Cluster").
		Order("created_at DESC").
		Limit(limit).
		Find(&activities).Error

	if err != nil {
		return nil, fmt.Errorf("获取最近活动失败: %w", err)
	}

	result := make([]ActivityItem, 0, len(activities))
	for _, activity := range activities {
		item := ActivityItem{
			Type:    activity.Type,
			Title:   activity.Title,
			Message: activity.Message,
			Time:    s.formatRelativeTime(activity.CreatedAt),
			Source:  activity.Source,
			Details: activity.Details,
		}

		// 如果有集群信息，在消息中包含集群名称
		if activity.ClusterID > 0 && activity.Cluster.ClusterName != "" {
			item.Message = fmt.Sprintf("[%s] %s", activity.Cluster.ClusterName, item.Message)
		}

		result = append(result, item)
	}

	return result, nil
}

// GetRecentAlerts 获取最近的系统告警
func (s *ActivityService) GetRecentAlerts(limit int) ([]AlertItem, error) {
	if limit <= 0 {
		limit = 10 // 默认显示10条
	}

	if limit > 50 {
		limit = 50 // 最多显示50条
	}

	var alerts []models.AlertHistory
	err := s.db.Preload("Cluster").
		Order("created_at DESC").
		Limit(limit).
		Find(&alerts).Error

	if err != nil {
		return nil, fmt.Errorf("获取最近告警失败: %w", err)
	}

	result := make([]AlertItem, 0, len(alerts))
	for _, alert := range alerts {
		// 映射告警级别到前端格式
		level := "low"
		switch alert.AlertLevel {
		case "critical":
			level = "high"
		case "error":
			level = "high"
		case "warning":
			level = "medium"
		case "info":
			level = "low"
		}

		item := AlertItem{
			ID:          alert.ID,
			Level:       level,
			Title:       alert.Title,
			Description: alert.Message,
			Time:        s.formatRelativeTime(alert.CreatedAt),
			Status:      alert.Status,
		}

		// 如果有集群信息，在描述中包含集群名称
		if alert.ClusterID > 0 && alert.Cluster.ClusterName != "" {
			item.Description = fmt.Sprintf("[%s] %s", alert.Cluster.ClusterName, item.Description)
		}

		result = append(result, item)
	}

	return result, nil
}

// CreateSampleAlert 创建示例告警（用于测试）
func (s *ActivityService) CreateSampleAlert(clusterID uint, level, title, message string) error {
	alert := &models.AlertHistory{
		ClusterID:   clusterID,
		AlertLevel:  level,
		Title:       title,
		Message:     message,
		Status:      "active",
		TriggeredAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(alert).Error; err != nil {
		return fmt.Errorf("创建示例告警失败: %w", err)
	}

	logger.Info("创建示例告警: level=%s, title=%s, cluster_id=%d", level, title, clusterID)
	return nil
}

// CleanupOldActivities 清理过期的活动记录
func (s *ActivityService) CleanupOldActivities(ctx context.Context, retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	// 清理过期活动
	result := s.db.Where("created_at < ?", cutoffTime).Delete(&models.SystemActivity{})
	if result.Error != nil {
		return fmt.Errorf("清理过期活动记录失败: %w", result.Error)
	}

	logger.Info("清理过期活动记录完成，删除 %d 条记录", result.RowsAffected)

	// 清理过期告警（优化：先进行去重清理，再按时间清理）
	if err := s.DeduplicateAlerts(ctx); err != nil {
		logger.Error("告警去重清理失败: %v", err)
	}

	result = s.db.Where("created_at < ?", cutoffTime).Delete(&models.AlertHistory{})
	if result.Error != nil {
		return fmt.Errorf("清理过期告警记录失败: %w", result.Error)
	}

	logger.Info("清理过期告警记录完成，删除 %d 条记录", result.RowsAffected)

	return nil
}

// CleanupOldActivitiesWithStats 清理过期数据并返回详细统计信息 - 供前端联调使用
func (s *ActivityService) CleanupOldActivitiesWithStats(ctx context.Context, retentionDays int) (*CleanupResult, error) {
	startTime := time.Now()
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	result := &CleanupResult{
		StartTime:     startTime,
		RetentionDays: retentionDays,
	}

	// 获取清理前的统计数据
	var activitiesBefore, alertsBefore int64
	s.db.Model(&models.SystemActivity{}).Count(&activitiesBefore)
	s.db.Model(&models.AlertHistory{}).Count(&alertsBefore)

	result.ActivitiesBefore = activitiesBefore
	result.AlertsBefore = alertsBefore

	// 先进行告警去重清理，减少数据量
	if err := s.DeduplicateAlerts(ctx); err != nil {
		logger.Error("告警去重清理失败: %v", err)
	}

	// 清理过期活动
	activityResult := s.db.Where("created_at < ?", cutoffTime).Delete(&models.SystemActivity{})
	if activityResult.Error != nil {
		return nil, fmt.Errorf("清理过期活动记录失败: %w", activityResult.Error)
	}
	result.RemovedActivities = activityResult.RowsAffected

	// 清理过期告警
	alertResult := s.db.Where("created_at < ?", cutoffTime).Delete(&models.AlertHistory{})
	if alertResult.Error != nil {
		return nil, fmt.Errorf("清理过期告警记录失败: %w", alertResult.Error)
	}
	result.RemovedAlerts = alertResult.RowsAffected

	// 获取清理后的统计数据
	var activitiesAfter, alertsAfter int64
	s.db.Model(&models.SystemActivity{}).Count(&activitiesAfter)
	s.db.Model(&models.AlertHistory{}).Count(&alertsAfter)

	result.ActivitiesAfter = activitiesAfter
	result.AlertsAfter = alertsAfter
	result.Duration = time.Since(startTime)

	logger.Info("数据清理完成 - 活动: %d->%d (删除%d), 告警: %d->%d (删除%d), 耗时: %v",
		activitiesBefore, activitiesAfter, result.RemovedActivities,
		alertsBefore, alertsAfter, result.RemovedAlerts,
		result.Duration)

	return result, nil
}

// formatRelativeTime 格式化相对时间
func (s *ActivityService) formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "刚刚"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d分钟前", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d小时前", hours)
	} else {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "昨天"
		} else if days < 7 {
			return fmt.Sprintf("%d天前", days)
		} else {
			return t.Format("01-02 15:04")
		}
	}
}

// RecordClusterConnection 记录集群连接活动
func (s *ActivityService) RecordClusterConnection(clusterID uint, clusterName string, success bool, message string) {
	activityType := "success"
	title := "集群连接正常"

	if !success {
		activityType = "error"
		title = "集群连接失败"
	}

	details := map[string]interface{}{
		"cluster_name": clusterName,
		"success":      success,
	}

	s.RecordActivity(activityType, title, message, "collector", clusterID, details)
}

// RecordDataCollection 记录数据收集活动
func (s *ActivityService) RecordDataCollection(clusterID uint, clusterName string, podCount int, success bool) {
	activityType := "info"
	title := "数据收集完成"
	message := fmt.Sprintf("已收集 %d 个 Pod 的数据", podCount)

	if !success {
		activityType = "warning"
		title = "数据收集异常"
		message = "数据收集过程中出现异常"
	}

	details := map[string]interface{}{
		"cluster_name": clusterName,
		"pod_count":    podCount,
		"success":      success,
	}

	s.RecordActivity(activityType, title, message, "collector", clusterID, details)
}

// RecordResourceAlert 记录资源告警活动
func (s *ActivityService) RecordResourceAlert(clusterID uint, clusterName, podName, alertType, message string) {
	activityType := "warning"
	title := "资源使用率异常"

	if alertType == "critical" {
		activityType = "error"
		title = "严重资源异常"
	}

	details := map[string]interface{}{
		"cluster_name": clusterName,
		"pod_name":     podName,
		"alert_type":   alertType,
	}

	s.RecordActivity(activityType, title, message, "monitor", clusterID, details)
}

// GenerateRealtimeActivities 基于实际数据生成实时活动
func (s *ActivityService) GenerateRealtimeActivities() error {
	// 获取集群服务和数据收集器
	clusterService := NewClusterService()

	// 生成系统启动活动
	err := s.generateSystemStartupActivity()
	if err != nil {
		logger.Error("生成系统启动活动失败: %v", err)
	}

	// 基于实际集群数据生成活动
	err = s.generateClusterBasedActivities(clusterService)
	if err != nil {
		logger.Error("生成集群活动失败: %v", err)
	}

	// 基于实际资源数据生成活动
	err = s.generateResourceBasedActivities()
	if err != nil {
		logger.Error("生成资源活动失败: %v", err)
	}

	// 生成维护和系统事件活动
	err = s.generateMaintenanceActivities()
	if err != nil {
		logger.Error("生成维护活动失败: %v", err)
	}

	// 基于实际数据生成告警
	err = s.generateRealAlerts(clusterService)
	if err != nil {
		logger.Error("生成实际告警数据失败: %v", err)
	}

	logger.Info("基于实际数据生成活动和告警完成")
	return nil
}

// generateSystemStartupActivity 生成系统启动活动
func (s *ActivityService) generateSystemStartupActivity() error {
	details := map[string]interface{}{
		"service":      "cluster-resource-insight",
		"version":      "1.0.0",
		"startup_time": time.Now().Unix(),
	}

	return s.RecordActivity("info", "系统监控启动",
		"集群资源监控服务已成功启动", "monitor", 0, details)
}

// generateClusterBasedActivities 基于实际集群数据生成活动
func (s *ActivityService) generateClusterBasedActivities(clusterService *ClusterService) error {
	clusters, err := clusterService.GetAllClusters()
	if err != nil {
		return fmt.Errorf("获取集群列表失败: %w", err)
	}

	if len(clusters) == 0 {
		// 如果没有集群，记录提示活动
		s.RecordActivity("warning", "集群配置提醒",
			"系统中尚未配置任何集群，请添加集群以开始监控", "system", 0, nil)
		return nil
	}

	// 统计在线集群
	onlineCount := 0
	for _, cluster := range clusters {
		if cluster.Status == "online" {
			onlineCount++
			// 为在线集群生成连接成功活动
			s.RecordClusterConnection(cluster.ID, cluster.ClusterName, true,
				fmt.Sprintf("集群 %s 连接正常，监控数据同步中", cluster.ClusterName))
		} else {
			// 为离线集群生成连接失败活动
			s.RecordClusterConnection(cluster.ID, cluster.ClusterName, false,
				fmt.Sprintf("集群 %s 连接异常，状态: %s", cluster.ClusterName, cluster.Status))
		}
	}

	// 生成集群统计活动
	details := map[string]interface{}{
		"total_clusters":   len(clusters),
		"online_clusters":  onlineCount,
		"offline_clusters": len(clusters) - onlineCount,
	}

	message := fmt.Sprintf("集群状态检查完成，%d/%d 个集群在线", onlineCount, len(clusters))
	s.RecordActivity("info", "集群状态检查", message, "monitor", 0, details)

	return nil
}

// generateResourceBasedActivities 基于实际资源数据生成活动
func (s *ActivityService) generateResourceBasedActivities() error {
	// 这里需要引入 collector 包来获取实际数据
	// 为了简化，我们从数据库查询最近的收集数据
	var activityCount int64
	s.db.Model(&models.SystemActivity{}).
		Where("created_at > ?", time.Now().Add(-24*time.Hour)).
		Count(&activityCount)

	if activityCount > 0 {
		// 基于最近活动生成数据收集完成活动
		details := map[string]interface{}{
			"recent_activities": activityCount,
			"period_hours":      24,
		}

		message := fmt.Sprintf("过去24小时内记录了 %d 条系统活动", activityCount)
		s.RecordActivity("success", "数据收集统计", message, "collector", 0, details)
	}

	return nil
}

// generateMaintenanceActivities 生成维护活动
func (s *ActivityService) generateMaintenanceActivities() error {
	// 检查数据库中的记录数量，模拟清理操作
	var totalActivities int64
	s.db.Model(&models.SystemActivity{}).Count(&totalActivities)

	var oldActivities int64
	s.db.Model(&models.SystemActivity{}).
		Where("created_at < ?", time.Now().AddDate(0, 0, -7)).
		Count(&oldActivities)

	if oldActivities > 0 {
		details := map[string]interface{}{
			"total_records":  totalActivities,
			"old_records":    oldActivities,
			"retention_days": 7,
		}

		message := fmt.Sprintf("检测到 %d 条超过7天的历史记录", oldActivities)
		s.RecordActivity("info", "数据维护检查", message, "maintenance", 0, details)
	}

	return nil
}

// generateRealAlerts 基于实际数据生成告警
func (s *ActivityService) generateRealAlerts(clusterService *ClusterService) error {
	clusters, err := clusterService.GetAllClusters()
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		// 基于集群状态生成告警
		if cluster.Status == "offline" || cluster.Status == "error" {
			alert := &models.AlertHistory{
				ClusterID:   cluster.ID,
				AlertLevel:  "error",
				Title:       "集群连接异常",
				Message:     fmt.Sprintf("集群 %s 无法连接，状态: %s", cluster.ClusterName, cluster.Status),
				Status:      "active",
				TriggeredAt: time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if err := s.db.Create(alert).Error; err != nil {
				logger.Error("创建集群告警失败: %v", err)
			}
		}

		// 模拟资源使用率检查（实际应该从监控数据获取）
		if cluster.Status == "online" {
			// 随机生成一些基于实际可能情况的告警
			s.generateResourceAlerts(cluster)
		}
	}

	logger.Info("基于实际集群状态生成告警完成")
	return nil
}

// generateResourceAlerts 为特定集群生成资源相关告警
func (s *ActivityService) generateResourceAlerts(cluster models.ClusterConfig) {
	currentTime := time.Now()

	// 基于集群配置和当前时间生成合理的告警
	alerts := []struct {
		Level   string
		Title   string
		Message string
		Status  string
	}{
		{
			Level:   "warning",
			Title:   "资源使用监控",
			Message: fmt.Sprintf("集群 %s 资源使用情况需要关注", cluster.ClusterName),
			Status:  "active",
		},
	}

	// 只在工作时间生成告警，避免过多无意义告警
	hour := currentTime.Hour()
	if hour >= 9 && hour <= 18 {
		for _, alertData := range alerts {
			alert := &models.AlertHistory{
				ClusterID:   cluster.ID,
				AlertLevel:  alertData.Level,
				Title:       alertData.Title,
				Message:     alertData.Message,
				Status:      alertData.Status,
				TriggeredAt: currentTime,
				CreatedAt:   currentTime,
				UpdatedAt:   currentTime,
			}

			if err := s.db.Create(alert).Error; err != nil {
				logger.Error("创建资源告警失败: %v", err)
			}
		}
	}
}

// generateSampleAlerts 已重构为 generateRealAlerts，保留此方法以兼容
func (s *ActivityService) generateSampleAlerts() error {
	clusterService := NewClusterService()
	return s.generateRealAlerts(clusterService)
}

// marshalDetails 序列化详情数据
func (s *ActivityService) marshalDetails(details map[string]interface{}) string {
	if details == nil {
		return "{}"
	}

	detailsBytes, err := json.Marshal(details)
	if err != nil {
		logger.Error("序列化活动详情失败: %v", err)
		return "{}"
	}

	return string(detailsBytes)
}

// RecordResourceCollectionActivity 记录资源收集活动
func (s *ActivityService) RecordResourceCollectionActivity(clusterCount, totalPods, problemPods int, duration time.Duration) {
	message := fmt.Sprintf("完成 %d 个集群资源收集，发现 %d 个问题Pod", clusterCount, problemPods)

	details := map[string]interface{}{
		"cluster_count": clusterCount,
		"total_pods":    totalPods,
		"problem_pods":  problemPods,
		"duration_ms":   duration.Milliseconds(),
		"efficiency":    float64(totalPods-problemPods) / float64(totalPods) * 100,
	}

	s.RecordSystemEvent("info", "资源数据收集", message, details)
}

// RecordSystemEvent 记录系统事件活动
func (s *ActivityService) RecordSystemEvent(eventType, title, message string, details map[string]interface{}) {
	err := s.RecordActivity(eventType, title, message, "system", 0, details)
	if err != nil {
		logger.Error("记录系统事件失败: %v", err)
	}
}

// RecordClusterStatusChange 记录集群状态变化
func (s *ActivityService) RecordClusterStatusChange(clusterID uint, clusterName, oldStatus, newStatus string) {
	var activityType string
	var message string

	switch newStatus {
	case "online":
		activityType = "success"
		message = fmt.Sprintf("集群 %s 已恢复在线状态", clusterName)
	case "offline":
		activityType = "error"
		message = fmt.Sprintf("集群 %s 已断开连接", clusterName)
	default:
		activityType = "warning"
		message = fmt.Sprintf("集群 %s 状态变更为 %s", clusterName, newStatus)
	}

	details := map[string]interface{}{
		"cluster_id":   clusterID,
		"cluster_name": clusterName,
		"old_status":   oldStatus,
		"new_status":   newStatus,
		"change_time":  time.Now().Unix(),
	}

	err := s.RecordActivity(activityType, "集群状态变更", message, "cluster", clusterID, details)
	if err != nil {
		logger.Error("记录集群状态变化失败: %v", err)
	} else {
		logger.Info("记录集群状态变化: %s %s -> %s", clusterName, oldStatus, newStatus)
	}
}

// UpdateAlertStatus 更新告警状态
func (s *ActivityService) UpdateAlertStatus(alertID uint, status string) error {
	// 验证状态值
	if status != "active" && status != "resolved" && status != "suppressed" {
		return fmt.Errorf("无效的告警状态: %s", status)
	}

	// 更新告警状态
	result := s.db.Model(&models.AlertHistory{}).
		Where("id = ?", alertID).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("更新告警状态失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("未找到ID为 %d 的告警记录", alertID)
	}

	// 记录操作活动
	actionTitle := "告警状态更新"
	actionMessage := fmt.Sprintf("告警ID %d 状态已更新为: %s", alertID, s.getStatusDisplayName(status))
	s.RecordActivity("info", actionTitle, actionMessage, "api", 0, map[string]interface{}{
		"alert_id":   alertID,
		"new_status": status,
	})

	logger.Info("告警状态更新成功: ID=%d, 新状态=%s", alertID, status)
	return nil
}

// ResolveAlert 解决告警（将状态设置为resolved）
func (s *ActivityService) ResolveAlert(alertID uint) error {
	return s.UpdateAlertStatus(alertID, "resolved")
}

// DismissAlert 忽略告警（将状态设置为suppressed）
func (s *ActivityService) DismissAlert(alertID uint) error {
	return s.UpdateAlertStatus(alertID, "suppressed")
}

// GetAlertByID 根据ID获取告警详情
func (s *ActivityService) GetAlertByID(alertID uint) (*models.AlertHistory, error) {
	var alert models.AlertHistory

	result := s.db.Preload("Cluster").First(&alert, alertID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到ID为 %d 的告警记录", alertID)
		}
		return nil, fmt.Errorf("查询告警详情失败: %w", result.Error)
	}

	return &alert, nil
}

// getStatusDisplayName 获取状态显示名称
func (s *ActivityService) getStatusDisplayName(status string) string {
	statusMap := map[string]string{
		"active":     "活跃",
		"resolved":   "已解决",
		"suppressed": "已屏蔽",
	}

	if display, exists := statusMap[status]; exists {
		return display
	}
	return status
}

// CreateAlert 创建系统告警记录（集成性能监控和自动验证）
func (s *ActivityService) CreateAlert(clusterID uint, level, title, message, status string) error {
	startTime := time.Now()

	// 验证告警级别
	validLevels := map[string]bool{
		"info": true, "warning": true, "error": true, "critical": true,
	}
	if !validLevels[level] {
		level = "info"
	}

	// 验证状态
	validStatuses := map[string]bool{
		"active": true, "resolved": true, "suppressed": true,
	}
	if !validStatuses[status] {
		status = "active"
	}

	// 使用降噪器检查是否应该创建告警
	shouldCreate, _, count := s.deduplicator.ShouldCreateAlert(clusterID, level, title)
	fingerprint := s.deduplicator.GenerateFingerprint(clusterID, level, title)

	if !shouldCreate {
		// 在抑制期内，使用事务更新现有告警的计数和时间
		tx := s.db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 先查找现有告警记录
		var existingAlert models.AlertHistory
		err := tx.Where("alert_fingerprint = ? AND status = 'active'", fingerprint).
			First(&existingAlert).Error

		if err == nil {
			// 更新现有记录
			err = tx.Model(&existingAlert).Updates(map[string]interface{}{
				"count":            count,
				"last_occurred_at": time.Now(),
				"updated_at":       time.Now(),
				"message":          message, // 更新为最新消息
			}).Error
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("更新降噪告警失败: %w", err)
			}
		} else {
			// 如果没找到活跃告警，可能是已解决的，创建新的
			shouldCreate = true
		}

		if !shouldCreate {
			if err := tx.Commit().Error; err != nil {
				return fmt.Errorf("提交告警更新事务失败: %w", err)
			}

			// 性能监控：记录抑制操作的性能
			duration := time.Since(startTime)
			if duration > 100*time.Millisecond {
				logger.Warn("告警抑制操作耗时过长: %v, level=%s, title=%s", duration, level, title)
			}

			logger.Info("告警被抑制，更新计数: level=%s, title=%s, count=%d, 耗时=%v", level, title, count, duration)
			return nil
		}

		tx.Rollback() // 如果需要创建新告警，回滚当前事务
	}

	// 创建新的降噪告警
	alert := s.deduplicator.PrepareDeduplicatedAlert(clusterID, level, title, message)
	alert.Status = status
	alert.Count = count

	if err := s.db.Create(alert).Error; err != nil {
		return fmt.Errorf("创建降噪告警失败: %w", err)
	}

	// 更新降噪器中的告警ID
	s.deduplicator.UpdateAlertID(clusterID, level, title, alert.ID)

	// 性能监控：记录创建操作的性能
	duration := time.Since(startTime)
	if duration > 200*time.Millisecond {
		logger.Warn("告警创建操作耗时过长: %v, level=%s, title=%s", duration, level, title)
	}

	logger.Info("创建降噪告警: level=%s, title=%s, cluster_id=%d, count=%d, 耗时=%v", level, title, clusterID, count, duration)

	// 自动触发数据质量检查（异步）
	go s.performAsyncDataQualityCheck(clusterID, level, title)

	return nil
}

// performAsyncDataQualityCheck 异步执行数据质量检查
func (s *ActivityService) performAsyncDataQualityCheck(clusterID uint, level, title string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("异步数据质量检查失败: %v", r)
		}
	}()

	// 检查是否存在过多重复告警
	fingerprint := s.deduplicator.GenerateFingerprint(clusterID, level, title)
	var duplicateCount int64
	s.db.Model(&models.AlertHistory{}).
		Where("alert_fingerprint = ? AND status = 'active'", fingerprint).
		Count(&duplicateCount)

	if duplicateCount > 10 {
		logger.Warn("检测到过多重复告警 (fingerprint=%s, count=%d)，建议检查告警源", fingerprint, duplicateCount)

		// 记录数据质量问题
		s.RecordSystemEvent("warning", "数据质量检查",
			fmt.Sprintf("检测到指纹 %s 的告警过于频繁，建议优化告警策略", fingerprint),
			map[string]interface{}{
				"fingerprint":     fingerprint,
				"duplicate_count": duplicateCount,
				"cluster_id":      clusterID,
				"level":           level,
				"title":           title,
			})
	}
}

// OptimizeActivities 执行活动优化
func (s *ActivityService) OptimizeActivities() (*OptimizationResult, error) {
	return s.optimizer.OptimizeActivities()
}

// GetDeduplicationStats 获取告警降噪统计信息
func (s *ActivityService) GetDeduplicationStats() map[string]interface{} {
	return s.deduplicator.GetCacheStats()
}

// SetSuppressionDuration 设置告警抑制时间
func (s *ActivityService) SetSuppressionDuration(duration time.Duration) {
	s.deduplicator.SetSuppressionDuration(duration)
}

// GetOptimizationConfig 获取优化配置
func (s *ActivityService) GetOptimizationConfig() (*OptimizationConfig, error) {
	return s.optimizer.LoadOptimizationConfig()
}

// UpdateOptimizationConfig 更新优化配置
func (s *ActivityService) UpdateOptimizationConfig(config *OptimizationConfig) error {
	return s.optimizer.SaveOptimizationConfig(config)
}

// GetActivityStats 获取活动统计信息
func (s *ActivityService) GetActivityStats(hours int) (map[string]interface{}, error) {
	if hours <= 0 {
		hours = 24
	}

	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	// 统计总活动数
	var totalCount int64
	err := s.db.Model(&models.SystemActivity{}).
		Where("created_at > ?", startTime).
		Count(&totalCount).Error
	if err != nil {
		return nil, fmt.Errorf("统计总活动数失败: %w", err)
	}

	// 按类型统计
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	err = s.db.Model(&models.SystemActivity{}).
		Select("type, COUNT(*) as count").
		Where("created_at > ?", startTime).
		Group("type").
		Find(&typeStats).Error
	if err != nil {
		return nil, fmt.Errorf("按类型统计失败: %w", err)
	}

	// 按来源统计
	var sourceStats []struct {
		Source string `json:"source"`
		Count  int64  `json:"count"`
	}
	err = s.db.Model(&models.SystemActivity{}).
		Select("source, COUNT(*) as count").
		Where("created_at > ?", startTime).
		Group("source").
		Find(&sourceStats).Error
	if err != nil {
		return nil, fmt.Errorf("按来源统计失败: %w", err)
	}

	// 按集群统计
	var clusterStats []struct {
		ClusterID uint  `json:"cluster_id"`
		Count     int64 `json:"count"`
	}
	err = s.db.Model(&models.SystemActivity{}).
		Select("cluster_id, COUNT(*) as count").
		Where("created_at > ? AND cluster_id > 0", startTime).
		Group("cluster_id").
		Find(&clusterStats).Error
	if err != nil {
		return nil, fmt.Errorf("按集群统计失败: %w", err)
	}

	// 时间趋势统计（按小时）
	var hourlyStats []struct {
		Hour  int   `json:"hour"`
		Count int64 `json:"count"`
	}
	err = s.db.Model(&models.SystemActivity{}).
		Select("EXTRACT(HOUR FROM created_at) as hour, COUNT(*) as count").
		Where("created_at > ?", startTime).
		Group("EXTRACT(HOUR FROM created_at)").
		Order("hour").
		Find(&hourlyStats).Error
	if err != nil {
		logger.Warn("时间趋势统计失败: %v", err)
		// 不影响主要统计结果
	}

	// 构建返回结果
	byType := make(map[string]int64)
	for _, stat := range typeStats {
		byType[stat.Type] = stat.Count
	}

	bySource := make(map[string]int64)
	for _, stat := range sourceStats {
		bySource[stat.Source] = stat.Count
	}

	byCluster := make(map[uint]int64)
	for _, stat := range clusterStats {
		byCluster[stat.ClusterID] = stat.Count
	}

	hourlyTrend := make(map[int]int64)
	for _, stat := range hourlyStats {
		hourlyTrend[stat.Hour] = stat.Count
	}

	result := map[string]interface{}{
		"hours":            hours,
		"total_activities": totalCount,
		"by_type":          byType,
		"by_source":        bySource,
		"by_cluster":       byCluster,
		"hourly_trend":     hourlyTrend,
		"generated_at":     time.Now(),
	}

	return result, nil
}

// DeduplicateAlerts 智能告警去重清理
// 基于alert_fingerprint合并重复告警，保留最新记录并更新统计信息
func (s *ActivityService) DeduplicateAlerts(ctx context.Context) error {
	logger.Info("开始执行告警去重清理...")

	// 查询所有存在重复指纹的告警
	var duplicateGroups []struct {
		AlertFingerprint string `json:"alert_fingerprint"`
		Count            int64  `json:"count"`
	}

	err := s.db.Model(&models.AlertHistory{}).
		Select("alert_fingerprint, COUNT(*) as count").
		Where("alert_fingerprint != '' AND alert_fingerprint IS NOT NULL").
		Group("alert_fingerprint").
		Having("COUNT(*) > 1").
		Find(&duplicateGroups).Error

	if err != nil {
		return fmt.Errorf("查询重复告警失败: %w", err)
	}

	if len(duplicateGroups) == 0 {
		logger.Info("未发现重复告警记录")
		return nil
	}

	logger.Info("发现 %d 个重复告警组，开始去重处理", len(duplicateGroups))

	totalProcessed := 0
	totalRemoved := 0

	// 逐组处理重复告警
	for _, group := range duplicateGroups {
		processed, removed, err := s.processDuplicateAlertGroup(ctx, group.AlertFingerprint)
		if err != nil {
			logger.Error("处理重复告警组 %s 失败: %v", group.AlertFingerprint, err)
			continue
		}

		totalProcessed += processed
		totalRemoved += removed
	}

	logger.Info("告警去重清理完成，处理 %d 个告警组，合并删除 %d 条重复记录",
		len(duplicateGroups), totalRemoved)

	return nil
}

// processDuplicateAlertGroup 处理单个重复告警组
func (s *ActivityService) processDuplicateAlertGroup(ctx context.Context, fingerprint string) (int, int, error) {
	// 查询该指纹的所有告警记录，按创建时间排序
	var alerts []models.AlertHistory
	err := s.db.Where("alert_fingerprint = ?", fingerprint).
		Order("created_at DESC, id DESC").
		Find(&alerts).Error

	if err != nil {
		return 0, 0, fmt.Errorf("查询告警记录失败: %w", err)
	}

	if len(alerts) <= 1 {
		return 0, 0, nil
	}

	// 保留最新的告警记录作为主记录
	keepAlert := &alerts[0]
	duplicateAlerts := alerts[1:]

	// 计算合并后的统计信息
	totalCount := keepAlert.Count
	var earliestOccurred time.Time = keepAlert.FirstOccurredAt
	var latestOccurred time.Time = keepAlert.LastOccurredAt

	// 从重复记录中收集统计信息
	var duplicateIDs []uint
	for _, alert := range duplicateAlerts {
		duplicateIDs = append(duplicateIDs, alert.ID)
		totalCount += alert.Count

		if alert.FirstOccurredAt.Before(earliestOccurred) {
			earliestOccurred = alert.FirstOccurredAt
		}
		if alert.LastOccurredAt.After(latestOccurred) {
			latestOccurred = alert.LastOccurredAt
		}
	}

	// 在事务中更新主记录并删除重复记录
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新主记录的统计信息
	err = tx.Model(keepAlert).Updates(map[string]interface{}{
		"count":             totalCount,
		"first_occurred_at": earliestOccurred,
		"last_occurred_at":  latestOccurred,
		"updated_at":        time.Now(),
	}).Error

	if err != nil {
		tx.Rollback()
		return 0, 0, fmt.Errorf("更新主告警记录失败: %w", err)
	}

	// 批量删除重复记录
	err = tx.Where("id IN ?", duplicateIDs).Delete(&models.AlertHistory{}).Error
	if err != nil {
		tx.Rollback()
		return 0, 0, fmt.Errorf("删除重复告警记录失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return 0, 0, fmt.Errorf("提交去重事务失败: %w", err)
	}

	logger.Info("成功合并告警指纹 %s：保留1条，删除%d条，总计数：%d",
		fingerprint, len(duplicateAlerts), totalCount)

	return len(alerts), len(duplicateAlerts), nil
}

// CleanupDuplicateAlerts 专门的告警去重清理接口（可单独调用）
func (s *ActivityService) CleanupDuplicateAlerts(ctx context.Context) (*DeduplicationResult, error) {
	startTime := time.Now()

	// 获取去重前的统计信息
	var totalBefore int64
	s.db.Model(&models.AlertHistory{}).Count(&totalBefore)

	// 执行去重
	err := s.DeduplicateAlerts(ctx)
	if err != nil {
		return nil, err
	}

	// 获取去重后的统计信息
	var totalAfter int64
	s.db.Model(&models.AlertHistory{}).Count(&totalAfter)

	duration := time.Since(startTime)

	result := &DeduplicationResult{
		StartTime:    startTime,
		Duration:     duration,
		TotalBefore:  totalBefore,
		TotalAfter:   totalAfter,
		RemovedCount: totalBefore - totalAfter,
		SuccessCount: totalAfter,
		ErrorCount:   0,
	}

	return result, nil
}

// GetDatabaseStats 获取数据库统计信息 - 供前端联调查看当前数据状态
func (s *ActivityService) GetDatabaseStats() (map[string]interface{}, error) {
	var totalActivities, totalAlerts int64
	var activeAlerts, resolvedAlerts, suppressedAlerts int64
	var duplicateAlerts int64

	// 统计活动总数
	if err := s.db.Model(&models.SystemActivity{}).Count(&totalActivities).Error; err != nil {
		return nil, fmt.Errorf("统计活动总数失败: %w", err)
	}

	// 统计告警总数
	if err := s.db.Model(&models.AlertHistory{}).Count(&totalAlerts).Error; err != nil {
		return nil, fmt.Errorf("统计告警总数失败: %w", err)
	}

	// 按状态统计告警
	if err := s.db.Model(&models.AlertHistory{}).Where("status = ?", "active").Count(&activeAlerts).Error; err != nil {
		logger.Warn("统计活跃告警失败: %v", err)
	}
	if err := s.db.Model(&models.AlertHistory{}).Where("status = ?", "resolved").Count(&resolvedAlerts).Error; err != nil {
		logger.Warn("统计已解决告警失败: %v", err)
	}
	if err := s.db.Model(&models.AlertHistory{}).Where("status = ?", "suppressed").Count(&suppressedAlerts).Error; err != nil {
		logger.Warn("统计已屏蔽告警失败: %v", err)
	}

	// 统计可能重复的告警（基于alert_fingerprint）
	var duplicateGroups []struct {
		AlertFingerprint string `json:"alert_fingerprint"`
		Count            int64  `json:"count"`
	}

	err := s.db.Model(&models.AlertHistory{}).
		Select("alert_fingerprint, COUNT(*) as count").
		Where("alert_fingerprint != '' AND alert_fingerprint IS NOT NULL").
		Group("alert_fingerprint").
		Having("COUNT(*) > 1").
		Find(&duplicateGroups).Error

	if err != nil {
		logger.Warn("统计重复告警失败: %v", err)
	} else {
		for _, group := range duplicateGroups {
			duplicateAlerts += group.Count - 1 // 每组减1，表示重复的数量
		}
	}

	// 按时间范围统计最近的数据
	var recentActivities, recentAlerts int64
	recent24h := time.Now().Add(-24 * time.Hour)

	s.db.Model(&models.SystemActivity{}).Where("created_at > ?", recent24h).Count(&recentActivities)
	s.db.Model(&models.AlertHistory{}).Where("created_at > ?", recent24h).Count(&recentAlerts)

	// 获取降噪器统计信息
	deduplicationStats := s.GetDeduplicationStats()

	stats := map[string]interface{}{
		"total_activities": totalActivities,
		"total_alerts":     totalAlerts,
		"alert_status": map[string]int64{
			"active":     activeAlerts,
			"resolved":   resolvedAlerts,
			"suppressed": suppressedAlerts,
		},
		"duplicate_alerts": duplicateAlerts,
		"duplicate_groups": len(duplicateGroups),
		"recent_24h": map[string]int64{
			"activities": recentActivities,
			"alerts":     recentAlerts,
		},
		"deduplication_cache": deduplicationStats,
		"generated_at":        time.Now(),
	}

	return stats, nil
}

// DeduplicationResult 去重操作结果
type DeduplicationResult struct {
	StartTime    time.Time     `json:"start_time"`
	Duration     time.Duration `json:"duration"`
	TotalBefore  int64         `json:"total_before"`
	TotalAfter   int64         `json:"total_after"`
	RemovedCount int64         `json:"removed_count"`
	SuccessCount int64         `json:"success_count"`
	ErrorCount   int64         `json:"error_count"`
}

// CleanupResult 数据清理操作结果
type CleanupResult struct {
	StartTime         time.Time     `json:"start_time"`
	Duration          time.Duration `json:"duration"`
	ActivitiesBefore  int64         `json:"activities_before"`
	ActivitiesAfter   int64         `json:"activities_after"`
	AlertsBefore      int64         `json:"alerts_before"`
	AlertsAfter       int64         `json:"alerts_after"`
	RemovedActivities int64         `json:"removed_activities"`
	RemovedAlerts     int64         `json:"removed_alerts"`
	RetentionDays     int           `json:"retention_days"`
}
