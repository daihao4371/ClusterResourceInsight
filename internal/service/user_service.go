package service

import (
	"errors"
	"fmt"
	"time"

	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户管理服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{
		db: database.GetDB(),
	}
}

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50"`
	Password string   `json:"password" binding:"required,min=6,max=50"`
	Email    string   `json:"email" binding:"required,email"`
	RealName string   `json:"real_name"`
	Phone    string   `json:"phone"`
	RoleIDs  []uint   `json:"role_ids"`
}

// UpdateUserRequest 更新用户请求结构
type UpdateUserRequest struct {
	Email    string   `json:"email" binding:"omitempty,email"`
	RealName string   `json:"real_name"`
	Phone    string   `json:"phone"`
	Status   string   `json:"status" binding:"omitempty,oneof=active inactive locked"`
	RoleIDs  []uint   `json:"role_ids"`
}

// ChangePasswordRequest 修改密码请求结构
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
}

// ResetPasswordRequest 重置密码请求结构
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
}

// CreateUser 创建用户
// 1. 校验用户名和邮箱唯一性
// 2. 密码加密存储
// 3. 关联角色
func (s *UserService) CreateUser(req *CreateUserRequest) (*models.User, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("邮箱已被使用")
	}

	// 密码加密 - 使用bcrypt保证安全性
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 创建用户对象
	user := &models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		RealName: req.RealName,
		Phone:    req.Phone,
		Status:   "active", // 默认激活状态
	}

	// 使用事务确保数据一致性
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 创建用户
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("创建用户失败: %v", err)
		}

		// 关联角色
		if len(req.RoleIDs) > 0 {
			var roles []models.Role
			if err := tx.Where("id IN ?", req.RoleIDs).Find(&roles).Error; err != nil {
				return fmt.Errorf("查询角色失败: %v", err)
			}
			if len(roles) != len(req.RoleIDs) {
				return errors.New("部分角色ID不存在")
			}
			if err := tx.Model(user).Association("Roles").Append(&roles); err != nil {
				return fmt.Errorf("关联角色失败: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 加载角色信息返回
	s.db.Preload("Roles").First(user, user.ID)
	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Roles").Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return &user, nil
}

// ListUsers 获取用户列表（支持分页和筛选）
func (s *UserService) ListUsers(page, pageSize int, keyword, status string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 构建查询条件
	query := s.db.Model(&models.User{})

	// 关键字搜索 - 支持用户名、真实姓名、邮箱
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ? OR email LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计用户数量失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("Roles").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %v", err)
	}

	return users, total, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(id uint, req *UpdateUserRequest) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	// 使用事务更新
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 更新基本信息
		updates := make(map[string]interface{})
		if req.Email != "" && req.Email != user.Email {
			// 检查新邮箱是否已被使用
			var existingUser models.User
			if err := tx.Where("email = ? AND id != ?", req.Email, id).First(&existingUser).Error; err == nil {
				return errors.New("邮箱已被使用")
			}
			updates["email"] = req.Email
		}
		if req.RealName != "" {
			updates["real_name"] = req.RealName
		}
		if req.Phone != "" {
			updates["phone"] = req.Phone
		}
		if req.Status != "" {
			updates["status"] = req.Status
		}

		if len(updates) > 0 {
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新用户信息失败: %v", err)
			}
		}

		// 更新角色关联
		if req.RoleIDs != nil {
			var roles []models.Role
			if len(req.RoleIDs) > 0 {
				if err := tx.Where("id IN ?", req.RoleIDs).Find(&roles).Error; err != nil {
					return fmt.Errorf("查询角色失败: %v", err)
				}
				if len(roles) != len(req.RoleIDs) {
					return errors.New("部分角色ID不存在")
				}
			}
			if err := tx.Model(&user).Association("Roles").Replace(&roles); err != nil {
				return fmt.Errorf("更新角色关联失败: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 重新加载用户信息
	s.db.Preload("Roles").First(&user, id)
	return &user, nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(id uint) error {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %v", err)
	}

	// 软删除用户
	if err := s.db.Delete(&user).Error; err != nil {
		return fmt.Errorf("删除用户失败: %v", err)
	}

	return nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(id uint, req *ChangePasswordRequest) error {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %v", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	// 更新密码
	if err := s.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("更新密码失败: %v", err)
	}

	return nil
}

// ResetPassword 重置密码（管理员操作）
func (s *UserService) ResetPassword(id uint, req *ResetPasswordRequest) error {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %v", err)
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	// 更新密码
	if err := s.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("更新密码失败: %v", err)
	}

	return nil
}

// UpdateLoginInfo 更新登录信息
func (s *UserService) UpdateLoginInfo(id uint, ip string) error {
	now := time.Now()
	if err := s.db.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": ip,
	}).Error; err != nil {
		return fmt.Errorf("更新登录信息失败: %v", err)
	}
	return nil
}

// ValidatePassword 验证密码
func (s *UserService) ValidatePassword(username, password string) (*models.User, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("用户已被禁用或锁定")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("密码错误")
	}

	return user, nil
}