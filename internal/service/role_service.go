package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/models"

	"gorm.io/gorm"
)

// RoleService 角色管理服务
type RoleService struct {
	db *gorm.DB
}

// NewRoleService 创建角色服务实例
func NewRoleService() *RoleService {
	return &RoleService{
		db: database.GetDB(),
	}
}

// CreateRoleRequest 创建角色请求结构
type CreateRoleRequest struct {
	RoleName    string   `json:"role_name" binding:"required,min=2,max=50"`
	RoleCode    string   `json:"role_code" binding:"required,min=2,max=50"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// UpdateRoleRequest 更新角色请求结构
type UpdateRoleRequest struct {
	RoleName    string   `json:"role_name" binding:"omitempty,min=2,max=50"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	Status      string   `json:"status" binding:"omitempty,oneof=active inactive"`
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(req *CreateRoleRequest) (*models.Role, error) {
	// 检查角色名称是否已存在
	var existingRole models.Role
	if err := s.db.Where("role_name = ?", req.RoleName).First(&existingRole).Error; err == nil {
		return nil, errors.New("角色名称已存在")
	}

	// 检查角色编码是否已存在
	if err := s.db.Where("role_code = ?", req.RoleCode).First(&existingRole).Error; err == nil {
		return nil, errors.New("角色编码已存在")
	}

	// 权限列表转JSON
	permissionsJSON, err := json.Marshal(req.Permissions)
	if err != nil {
		return nil, fmt.Errorf("权限数据序列化失败: %v", err)
	}

	// 创建角色对象
	role := &models.Role{
		RoleName:    req.RoleName,
		RoleCode:    req.RoleCode,
		Description: req.Description,
		Permissions: string(permissionsJSON),
		Status:      "active",
		IsSystem:    false, // 用户创建的角色默认不是系统角色
	}

	// 保存到数据库
	if err := s.db.Create(role).Error; err != nil {
		return nil, fmt.Errorf("创建角色失败: %v", err)
	}

	return role, nil
}

// GetRoleByID 根据ID获取角色
func (s *RoleService) GetRoleByID(id uint) (*models.Role, error) {
	var role models.Role
	if err := s.db.First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, fmt.Errorf("查询角色失败: %v", err)
	}
	return &role, nil
}

// GetRoleByCode 根据角色编码获取角色
func (s *RoleService) GetRoleByCode(code string) (*models.Role, error) {
	var role models.Role
	if err := s.db.Where("role_code = ?", code).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, fmt.Errorf("查询角色失败: %v", err)
	}
	return &role, nil
}

// ListRoles 获取角色列表（支持分页和筛选）
func (s *RoleService) ListRoles(page, pageSize int, keyword, status string) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	// 构建查询条件
	query := s.db.Model(&models.Role{})

	// 关键字搜索 - 支持角色名称、角色编码、描述
	if keyword != "" {
		query = query.Where("role_name LIKE ? OR role_code LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计角色数量失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("is_system DESC, created_at DESC"). // 系统角色优先显示
							Limit(pageSize).
							Offset(offset).
							Find(&roles).Error; err != nil {
		return nil, 0, fmt.Errorf("查询角色列表失败: %v", err)
	}

	return roles, total, nil
}

// UpdateRole 更新角色信息
func (s *RoleService) UpdateRole(id uint, req *UpdateRoleRequest) (*models.Role, error) {
	var role models.Role
	if err := s.db.First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, fmt.Errorf("查询角色失败: %v", err)
	}

	// 系统内置角色不允许修改核心属性
	if role.IsSystem && (req.RoleName != "" || req.Status != "") {
		return nil, errors.New("系统内置角色不允许修改名称和状态")
	}

	// 构建更新数据
	updates := make(map[string]interface{})

	if req.RoleName != "" && req.RoleName != role.RoleName {
		// 检查新角色名是否已存在
		var existingRole models.Role
		if err := s.db.Where("role_name = ? AND id != ?", req.RoleName, id).First(&existingRole).Error; err == nil {
			return nil, errors.New("角色名称已存在")
		}
		updates["role_name"] = req.RoleName
	}

	if req.Description != "" {
		updates["description"] = req.Description
	}

	if req.Status != "" {
		updates["status"] = req.Status
	}

	// 更新权限列表
	if req.Permissions != nil {
		permissionsJSON, err := json.Marshal(req.Permissions)
		if err != nil {
			return nil, fmt.Errorf("权限数据序列化失败: %v", err)
		}
		updates["permissions"] = string(permissionsJSON)
	}

	// 执行更新
	if len(updates) > 0 {
		if err := s.db.Model(&role).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新角色失败: %v", err)
		}
	}

	// 重新加载角色信息
	s.db.First(&role, id)
	return &role, nil
}

// DeleteRole 删除角色（软删除）
func (s *RoleService) DeleteRole(id uint) error {
	var role models.Role
	if err := s.db.First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return fmt.Errorf("查询角色失败: %v", err)
	}

	// 系统内置角色不允许删除
	if role.IsSystem {
		return errors.New("系统内置角色不允许删除")
	}

	// 检查是否有用户正在使用该角色
	userCount := s.db.Model(&role).Association("Users").Count()
	if userCount > 0 {
		return fmt.Errorf("该角色正在被 %d 个用户使用，无法删除", userCount)
	}

	// 软删除角色
	if err := s.db.Delete(&role).Error; err != nil {
		return fmt.Errorf("删除角色失败: %v", err)
	}

	return nil
}

// GetRoleUsers 获取角色下的所有用户
func (s *RoleService) GetRoleUsers(id uint, page, pageSize int) ([]models.User, int64, error) {
	var role models.Role
	if err := s.db.First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errors.New("角色不存在")
		}
		return nil, 0, fmt.Errorf("查询角色失败: %v", err)
	}

	var users []models.User

	// 统计总数
	total := s.db.Model(&role).Association("Users").Count()

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Model(&role).
		Limit(pageSize).
		Offset(offset).
		Association("Users").
		Find(&users); err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %v", err)
	}

	return users, total, nil
}

// GetAllActiveRoles 获取所有激活状态的角色（用于角色选择器）
func (s *RoleService) GetAllActiveRoles() ([]models.Role, error) {
	var roles []models.Role
	if err := s.db.Where("status = ?", "active").
		Order("is_system DESC, role_name ASC").
		Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("查询角色列表失败: %v", err)
	}
	return roles, nil
}

// GetRolePermissions 获取角色的权限列表（解析JSON）
func (s *RoleService) GetRolePermissions(id uint) ([]string, error) {
	role, err := s.GetRoleByID(id)
	if err != nil {
		return nil, err
	}

	var permissions []string
	if role.Permissions != "" {
		if err := json.Unmarshal([]byte(role.Permissions), &permissions); err != nil {
			return nil, fmt.Errorf("解析权限数据失败: %v", err)
		}
	}

	return permissions, nil
}

// CheckUserPermission 检查用户是否有指定权限
func (s *RoleService) CheckUserPermission(userID uint, permission string) (bool, error) {
	var user models.User
	if err := s.db.Preload("Roles").First(&user, userID).Error; err != nil {
		return false, fmt.Errorf("查询用户失败: %v", err)
	}

	// 遍历用户的所有角色
	for _, role := range user.Roles {
		if role.Status != "active" {
			continue
		}

		var permissions []string
		if role.Permissions != "" {
			if err := json.Unmarshal([]byte(role.Permissions), &permissions); err != nil {
				continue
			}

			// 检查权限列表中是否包含所需权限
			for _, p := range permissions {
				if p == permission || p == "*" { // * 表示超级权限
					return true, nil
				}
			}
		}
	}

	return false, nil
}