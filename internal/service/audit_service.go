package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/models"

	"gorm.io/gorm"
)

// AuditService 审计日志服务
type AuditService struct {
	db *gorm.DB
}

// NewAuditService 创建审计服务实例
func NewAuditService() *AuditService {
	return &AuditService{
		db: database.GetDB(),
	}
}

// CreateAuditLogRequest 创建审计日志请求结构
type CreateAuditLogRequest struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	RealName     string `json:"real_name"`
	OperateType  string `json:"operate_type"`
	Module       string `json:"module"`
	Resource     string `json:"resource"`
	ResourceID   string `json:"resource_id"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	RequestBody  string `json:"request_body"`
	ResponseCode int    `json:"response_code"`
	IP           string `json:"ip"`
	UserAgent    string `json:"user_agent"`
	ErrorMsg     string `json:"error_msg"`
	Duration     int64  `json:"duration"`
}

// CreateAuditLog 创建审计日志
// 使用异步写入避免阻塞主流程
func (s *AuditService) CreateAuditLog(req *CreateAuditLogRequest) error {
	// 脱敏请求体中的敏感信息
	sanitizedBody := s.sanitizeRequestBody(req.RequestBody)

	auditLog := &models.AuditLog{
		UserID:       req.UserID,
		Username:     req.Username,
		RealName:     req.RealName,
		OperateType:  req.OperateType,
		Module:       req.Module,
		Resource:     req.Resource,
		ResourceID:   req.ResourceID,
		Method:       req.Method,
		Path:         req.Path,
		RequestBody:  sanitizedBody,
		ResponseCode: req.ResponseCode,
		IP:           req.IP,
		UserAgent:    req.UserAgent,
		ErrorMsg:     req.ErrorMsg,
		Duration:     req.Duration,
	}

	// 异步写入数据库，避免阻塞主请求
	go func() {
		if err := s.db.Create(auditLog).Error; err != nil {
			// 记录到日志系统，但不影响主流程
			fmt.Printf("创建审计日志失败: %v\n", err)
		}
	}()

	return nil
}

// sanitizeRequestBody 脱敏请求体中的敏感信息
func (s *AuditService) sanitizeRequestBody(body string) string {
	if body == "" {
		return ""
	}

	// 尝试解析为JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// 如果不是JSON，返回原始内容（限制长度）
		if len(body) > 1000 {
			return body[:1000] + "..."
		}
		return body
	}

	// 定义需要脱敏的字段
	sensitiveFields := []string{
		"password", "pwd", "secret", "token", "api_key", "apikey",
		"access_token", "refresh_token", "auth", "authorization",
	}

	// 递归脱敏
	s.sanitizeMap(data, sensitiveFields)

	// 序列化回JSON
	sanitized, err := json.Marshal(data)
	if err != nil {
		return body
	}

	return string(sanitized)
}

// sanitizeMap 递归脱敏map中的敏感字段
func (s *AuditService) sanitizeMap(data map[string]interface{}, sensitiveFields []string) {
	for key, value := range data {
		// 检查是否为敏感字段
		lowerKey := strings.ToLower(key)
		for _, field := range sensitiveFields {
			if strings.Contains(lowerKey, field) {
				data[key] = "******"
				continue
			}
		}

		// 递归处理嵌套的map
		if nestedMap, ok := value.(map[string]interface{}); ok {
			s.sanitizeMap(nestedMap, sensitiveFields)
		}
	}
}

// QueryAuditLogs 查询审计日志（支持多条件筛选和分页）
func (s *AuditService) QueryAuditLogs(
	page, pageSize int,
	userID *uint,
	username, operateType, module, resource string,
	startTime, endTime *time.Time,
	minDuration, maxDuration *int64,
	responseCode *int,
) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	// 构建查询条件
	query := s.db.Model(&models.AuditLog{})

	// 用户ID筛选
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// 用户名筛选
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	// 操作类型筛选
	if operateType != "" {
		query = query.Where("operate_type = ?", operateType)
	}

	// 模块筛选
	if module != "" {
		query = query.Where("module = ?", module)
	}

	// 资源筛选
	if resource != "" {
		query = query.Where("resource LIKE ?", "%"+resource+"%")
	}

	// 时间范围筛选
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	// 耗时范围筛选
	if minDuration != nil {
		query = query.Where("duration >= ?", *minDuration)
	}
	if maxDuration != nil {
		query = query.Where("duration <= ?", *maxDuration)
	}

	// 响应码筛选
	if responseCode != nil {
		query = query.Where("response_code = ?", *responseCode)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计审计日志数量失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("User").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询审计日志失败: %v", err)
	}

	return logs, total, nil
}

// GetAuditLogByID 根据ID获取审计日志详情
func (s *AuditService) GetAuditLogByID(id uint) (*models.AuditLog, error) {
	var log models.AuditLog
	if err := s.db.Preload("User").First(&log, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("审计日志不存在")
		}
		return nil, fmt.Errorf("查询审计日志失败: %v", err)
	}
	return &log, nil
}

// GetUserAuditLogs 获取指定用户的审计日志
func (s *AuditService) GetUserAuditLogs(userID uint, page, pageSize int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := s.db.Model(&models.AuditLog{}).Where("user_id = ?", userID)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计审计日志数量失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询审计日志失败: %v", err)
	}

	return logs, total, nil
}

// GetAuditStatistics 获取审计统计信息
func (s *AuditService) GetAuditStatistics(startTime, endTime *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := s.db.Model(&models.AuditLog{})
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	// 总数统计
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("统计审计日志总数失败: %v", err)
	}
	stats["total"] = total

	// 按操作类型统计
	var typeStats []struct {
		OperateType string `json:"operate_type"`
		Count       int64  `json:"count"`
	}
	if err := query.Select("operate_type, COUNT(*) as count").
		Group("operate_type").
		Find(&typeStats).Error; err != nil {
		return nil, fmt.Errorf("统计操作类型失败: %v", err)
	}
	stats["by_operate_type"] = typeStats

	// 按模块统计
	var moduleStats []struct {
		Module string `json:"module"`
		Count  int64  `json:"count"`
	}
	if err := query.Select("module, COUNT(*) as count").
		Where("module != ''").
		Group("module").
		Order("count DESC").
		Limit(10).
		Find(&moduleStats).Error; err != nil {
		return nil, fmt.Errorf("统计模块失败: %v", err)
	}
	stats["by_module"] = moduleStats

	// 按用户统计（Top 10）
	var userStats []struct {
		Username string `json:"username"`
		Count    int64  `json:"count"`
	}
	if err := query.Select("username, COUNT(*) as count").
		Where("username != ''").
		Group("username").
		Order("count DESC").
		Limit(10).
		Find(&userStats).Error; err != nil {
		return nil, fmt.Errorf("统计用户操作失败: %v", err)
	}
	stats["by_user"] = userStats

	// 按响应码统计
	var statusStats []struct {
		ResponseCode int   `json:"response_code"`
		Count        int64 `json:"count"`
	}
	if err := query.Select("response_code, COUNT(*) as count").
		Group("response_code").
		Order("count DESC").
		Find(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("统计响应码失败: %v", err)
	}
	stats["by_status"] = statusStats

	// 平均耗时
	var avgDuration float64
	if err := query.Select("AVG(duration) as avg_duration").
		Scan(&avgDuration).Error; err != nil {
		return nil, fmt.Errorf("统计平均耗时失败: %v", err)
	}
	stats["avg_duration"] = avgDuration

	// 慢请求统计（耗时超过1秒）
	var slowCount int64
	if err := query.Where("duration > ?", 1000).Count(&slowCount).Error; err != nil {
		return nil, fmt.Errorf("统计慢请求失败: %v", err)
	}
	stats["slow_request_count"] = slowCount

	// 错误请求统计（4xx和5xx）
	var errorCount int64
	if err := query.Where("response_code >= ?", 400).Count(&errorCount).Error; err != nil {
		return nil, fmt.Errorf("统计错误请求失败: %v", err)
	}
	stats["error_count"] = errorCount

	return stats, nil
}

// GetOperationTrend 获取操作趋势（按天统计）
func (s *AuditService) GetOperationTrend(days int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	query := `
		SELECT
			DATE(created_at) as date,
			COUNT(*) as count,
			COUNT(DISTINCT user_id) as user_count
		FROM audit_logs
		WHERE created_at >= ?
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`

	if err := s.db.Raw(query, startDate).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("查询操作趋势失败: %v", err)
	}

	return results, nil
}

// CleanupOldLogs 清理旧的审计日志
func (s *AuditService) CleanupOldLogs(days int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	result := s.db.Where("created_at < ?", cutoffDate).Delete(&models.AuditLog{})
	if result.Error != nil {
		return 0, fmt.Errorf("清理审计日志失败: %v", result.Error)
	}

	return result.RowsAffected, nil
}

// ExportAuditLogs 导出审计日志（返回CSV格式数据）
func (s *AuditService) ExportAuditLogs(
	userID *uint,
	operateType, module string,
	startTime, endTime *time.Time,
) ([][]string, error) {
	var logs []models.AuditLog

	query := s.db.Model(&models.AuditLog{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if operateType != "" {
		query = query.Where("operate_type = ?", operateType)
	}
	if module != "" {
		query = query.Where("module = ?", module)
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("查询审计日志失败: %v", err)
	}

	// 构建CSV数据
	data := [][]string{
		{"ID", "用户名", "操作类型", "模块", "资源", "资源ID", "方法", "路径", "响应码", "IP", "耗时(ms)", "操作时间", "错误信息"},
	}

	for _, log := range logs {
		row := []string{
			fmt.Sprintf("%d", log.ID),
			log.Username,
			log.OperateType,
			log.Module,
			log.Resource,
			log.ResourceID,
			log.Method,
			log.Path,
			fmt.Sprintf("%d", log.ResponseCode),
			log.IP,
			fmt.Sprintf("%d", log.Duration),
			log.CreatedAt.Format("2006-01-02 15:04:05"),
			log.ErrorMsg,
		}
		data = append(data, row)
	}

	return data, nil
}