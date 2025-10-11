package service

import (
	"errors"
	"fmt"

	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/models"

	"gorm.io/gorm"
)

// ApiService API管理服务
type ApiService struct {
	db *gorm.DB
}

// NewApiService 创建API服务实例
func NewApiService() *ApiService {
	return &ApiService{
		db: database.GetDB(),
	}
}

// CreateApiRequest 创建API请求结构
type CreateApiRequest struct {
	Path        string `json:"path" binding:"required"`
	Method      string `json:"method" binding:"required,oneof=GET POST PUT DELETE PATCH"`
	Group       string `json:"group"`
	Description string `json:"description"`
	Permission  string `json:"permission"`
	IsPublic    bool   `json:"is_public"`
}

// UpdateApiRequest 更新API请求结构
type UpdateApiRequest struct {
	Path        string `json:"path"`
	Method      string `json:"method" binding:"omitempty,oneof=GET POST PUT DELETE PATCH"`
	Group       string `json:"group"`
	Description string `json:"description"`
	Permission  string `json:"permission"`
	IsPublic    *bool  `json:"is_public"`
	Status      string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// CreateApi 创建API接口
func (s *ApiService) CreateApi(req *CreateApiRequest) (*models.ApiEndpoint, error) {
	// 检查相同路径和方法的API是否已存在
	var existingApi models.ApiEndpoint
	if err := s.db.Where("path = ? AND method = ?", req.Path, req.Method).First(&existingApi).Error; err == nil {
		return nil, errors.New("相同路径和方法的API已存在")
	}

	// 创建API对象
	api := &models.ApiEndpoint{
		Path:        req.Path,
		Method:      req.Method,
		Group:       req.Group,
		Description: req.Description,
		Permission:  req.Permission,
		IsPublic:    req.IsPublic,
		Status:      "active",
	}

	// 保存到数据库
	if err := s.db.Create(api).Error; err != nil {
		return nil, fmt.Errorf("创建API失败: %v", err)
	}

	return api, nil
}

// GetApiByID 根据ID获取API
func (s *ApiService) GetApiByID(id uint) (*models.ApiEndpoint, error) {
	var api models.ApiEndpoint
	if err := s.db.First(&api, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API不存在")
		}
		return nil, fmt.Errorf("查询API失败: %v", err)
	}
	return &api, nil
}

// GetApiByPathAndMethod 根据路径和方法获取API
func (s *ApiService) GetApiByPathAndMethod(path, method string) (*models.ApiEndpoint, error) {
	var api models.ApiEndpoint
	if err := s.db.Where("path = ? AND method = ?", path, method).First(&api).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API不存在")
		}
		return nil, fmt.Errorf("查询API失败: %v", err)
	}
	return &api, nil
}

// ListApis 获取API列表（支持分页和筛选）
func (s *ApiService) ListApis(page, pageSize int, keyword, group, method, status string, isPublic *bool) ([]models.ApiEndpoint, int64, error) {
	var apis []models.ApiEndpoint
	var total int64

	// 构建查询条件
	query := s.db.Model(&models.ApiEndpoint{})

	// 关键字搜索 - 支持路径、描述
	if keyword != "" {
		query = query.Where("path LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%")
	}

	// 分组筛选
	if group != "" {
		query = query.Where("group = ?", group)
	}

	// 方法筛选
	if method != "" {
		query = query.Where("method = ?", method)
	}

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 公开接口筛选
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计API数量失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("group ASC, path ASC, method ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&apis).Error; err != nil {
		return nil, 0, fmt.Errorf("查询API列表失败: %v", err)
	}

	return apis, total, nil
}

// UpdateApi 更新API信息
func (s *ApiService) UpdateApi(id uint, req *UpdateApiRequest) (*models.ApiEndpoint, error) {
	var api models.ApiEndpoint
	if err := s.db.First(&api, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API不存在")
		}
		return nil, fmt.Errorf("查询API失败: %v", err)
	}

	// 构建更新数据
	updates := make(map[string]interface{})

	if req.Path != "" && (req.Path != api.Path || req.Method != api.Method) {
		// 检查新的路径和方法组合是否已存在
		var existingApi models.ApiEndpoint
		checkPath := req.Path
		if req.Path == "" {
			checkPath = api.Path
		}
		checkMethod := req.Method
		if req.Method == "" {
			checkMethod = api.Method
		}
		if err := s.db.Where("path = ? AND method = ? AND id != ?", checkPath, checkMethod, id).First(&existingApi).Error; err == nil {
			return nil, errors.New("相同路径和方法的API已存在")
		}
		updates["path"] = req.Path
	}

	if req.Method != "" {
		updates["method"] = req.Method
	}

	if req.Group != "" {
		updates["group"] = req.Group
	}

	if req.Description != "" {
		updates["description"] = req.Description
	}

	if req.Permission != "" {
		updates["permission"] = req.Permission
	}

	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	if req.Status != "" {
		updates["status"] = req.Status
	}

	// 执行更新
	if len(updates) > 0 {
		if err := s.db.Model(&api).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新API失败: %v", err)
		}
	}

	// 重新加载API信息
	s.db.First(&api, id)
	return &api, nil
}

// DeleteApi 删除API（软删除）
func (s *ApiService) DeleteApi(id uint) error {
	var api models.ApiEndpoint
	if err := s.db.First(&api, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("API不存在")
		}
		return fmt.Errorf("查询API失败: %v", err)
	}

	// 软删除API
	if err := s.db.Delete(&api).Error; err != nil {
		return fmt.Errorf("删除API失败: %v", err)
	}

	return nil
}

// GetApiGroups 获取所有API分组
func (s *ApiService) GetApiGroups() ([]string, error) {
	var groups []string
	if err := s.db.Model(&models.ApiEndpoint{}).
		Distinct("group").
		Where("group != ''").
		Order("group ASC").
		Pluck("group", &groups).Error; err != nil {
		return nil, fmt.Errorf("查询API分组失败: %v", err)
	}
	return groups, nil
}

// BatchCreateApis 批量创建API（用于系统初始化）
func (s *ApiService) BatchCreateApis(apis []CreateApiRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, req := range apis {
			// 检查是否已存在
			var existingApi models.ApiEndpoint
			if err := tx.Where("path = ? AND method = ?", req.Path, req.Method).First(&existingApi).Error; err == nil {
				// 已存在则跳过
				continue
			}

			api := &models.ApiEndpoint{
				Path:        req.Path,
				Method:      req.Method,
				Group:       req.Group,
				Description: req.Description,
				Permission:  req.Permission,
				IsPublic:    req.IsPublic,
				Status:      "active",
			}

			if err := tx.Create(api).Error; err != nil {
				return fmt.Errorf("创建API %s %s 失败: %v", req.Method, req.Path, err)
			}
		}
		return nil
	})
}

// SyncApis 同步API（删除不存在的，新增缺失的）
func (s *ApiService) SyncApis(apis []CreateApiRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 构建API映射表
		apiMap := make(map[string]bool)
		for _, req := range apis {
			key := fmt.Sprintf("%s:%s", req.Method, req.Path)
			apiMap[key] = true

			// 检查是否已存在
			var existingApi models.ApiEndpoint
			err := tx.Where("path = ? AND method = ?", req.Path, req.Method).First(&existingApi).Error
			if err == nil {
				// 已存在则更新
				updates := map[string]interface{}{
					"group":       req.Group,
					"description": req.Description,
					"permission":  req.Permission,
					"is_public":   req.IsPublic,
				}
				if err := tx.Model(&existingApi).Updates(updates).Error; err != nil {
					return fmt.Errorf("更新API %s %s 失败: %v", req.Method, req.Path, err)
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// 不存在则创建
				api := &models.ApiEndpoint{
					Path:        req.Path,
					Method:      req.Method,
					Group:       req.Group,
					Description: req.Description,
					Permission:  req.Permission,
					IsPublic:    req.IsPublic,
					Status:      "active",
				}
				if err := tx.Create(api).Error; err != nil {
					return fmt.Errorf("创建API %s %s 失败: %v", req.Method, req.Path, err)
				}
			} else {
				return fmt.Errorf("查询API失败: %v", err)
			}
		}

		// 删除不在列表中的API（可选，注释掉则不删除）
		// var allApis []models.ApiEndpoint
		// if err := tx.Find(&allApis).Error; err != nil {
		// 	return fmt.Errorf("查询所有API失败: %v", err)
		// }
		// for _, api := range allApis {
		// 	key := fmt.Sprintf("%s:%s", api.Method, api.Path)
		// 	if !apiMap[key] {
		// 		if err := tx.Delete(&api).Error; err != nil {
		// 			return fmt.Errorf("删除API %s %s 失败: %v", api.Method, api.Path, err)
		// 		}
		// 	}
		// }

		return nil
	})
}

// GetApiStatistics 获取API统计信息
func (s *ApiService) GetApiStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总数统计
	var total int64
	if err := s.db.Model(&models.ApiEndpoint{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("统计API总数失败: %v", err)
	}
	stats["total"] = total

	// 按状态统计
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := s.db.Model(&models.ApiEndpoint{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("统计API状态失败: %v", err)
	}
	stats["by_status"] = statusStats

	// 按分组统计
	var groupStats []struct {
		Group string
		Count int64
	}
	if err := s.db.Model(&models.ApiEndpoint{}).
		Select("`group`, COUNT(*) as count").
		Where("`group` != ''").
		Group("`group`").
		Order("count DESC").
		Limit(10).
		Find(&groupStats).Error; err != nil {
		return nil, fmt.Errorf("统计API分组失败: %v", err)
	}
	stats["by_group"] = groupStats

	// 按方法统计
	var methodStats []struct {
		Method string
		Count  int64
	}
	if err := s.db.Model(&models.ApiEndpoint{}).
		Select("method, COUNT(*) as count").
		Group("method").
		Find(&methodStats).Error; err != nil {
		return nil, fmt.Errorf("统计API方法失败: %v", err)
	}
	stats["by_method"] = methodStats

	// 公开/私有接口统计
	var publicCount, privateCount int64
	s.db.Model(&models.ApiEndpoint{}).Where("is_public = ?", true).Count(&publicCount)
	s.db.Model(&models.ApiEndpoint{}).Where("is_public = ?", false).Count(&privateCount)
	stats["public_count"] = publicCount
	stats["private_count"] = privateCount

	return stats, nil
}