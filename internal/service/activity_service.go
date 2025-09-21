package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/models"

	"gorm.io/gorm"
)

// ActivityService 活动服务管理器
type ActivityService struct {
	db *gorm.DB
}

// NewActivityService 创建活动服务实例
func NewActivityService() *ActivityService {
	return &ActivityService{
		db: database.GetDB(),
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
	Level       string `json:"level"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Time        string `json:"time"`
	Status      string `json:"status"`
}

// RecordActivity 记录系统活动
func (s *ActivityService) RecordActivity(activityType, title, message, source string, clusterID uint, details map[string]interface{}) error {
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

	activity := &models.SystemActivity{
		Type:      activityType,
		ClusterID: clusterID,
		Title:     title,
		Message:   message,
		Source:    source,
		Details:   detailsJSON,
		CreatedAt: time.Now(),
	}

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
	
	// 清理过期告警
	result = s.db.Where("created_at < ?", cutoffTime).Delete(&models.AlertHistory{})
	if result.Error != nil {
		return fmt.Errorf("清理过期告警记录失败: %w", result.Error)
	}
	
	logger.Info("清理过期告警记录完成，删除 %d 条记录", result.RowsAffected)
	
	return nil
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

// CreateAlert 创建系统告警记录
func (s *ActivityService) CreateAlert(clusterID uint, level, title, message, status string) error {
	// 验证告警级别
	validLevels := map[string]bool{
		"info": true, "warning": true, "error": true, "critical": true,
	}
	if !validLevels[level] {
		level = "info" // 默认级别
	}

	// 验证状态
	validStatuses := map[string]bool{
		"active": true, "resolved": true, "suppressed": true,
	}
	if !validStatuses[status] {
		status = "active" // 默认状态
	}

	alert := &models.AlertHistory{
		RuleID:      nil, // 系统生成的告警
		ClusterID:   clusterID,
		AlertLevel:  level,
		Title:       title,
		Message:     message,
		Status:      status,
		TriggeredAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(alert).Error; err != nil {
		return fmt.Errorf("创建系统告警失败: %w", err)
	}

	logger.Info("创建系统告警: level=%s, title=%s, cluster_id=%d", level, title, clusterID)
	return nil
}