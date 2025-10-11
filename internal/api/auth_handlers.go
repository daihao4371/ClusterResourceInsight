package api

import (
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"
	"cluster-resource-insight/internal/utils"

	"github.com/gin-gonic/gin"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token    string                 `json:"token"`
	UserInfo map[string]interface{} `json:"user_info"`
}

// Login 用户登录
// 验证用户名密码，生成JWT token
func Login(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, 400, "请求参数错误")
			return
		}

		// 根据用户名查询用户
		user, err := userService.GetUserByUsername(req.Username)
		if err != nil {
			response.Error(c, 401, "用户名或密码错误")
			return
		}

		// 检查用户状态
		if user.Status != "active" {
			response.Error(c, 403, "账户已被禁用或锁定")
			return
		}

		// 验证密码
		if !utils.CheckPassword(req.Password, user.Password) {
			response.Error(c, 401, "用户名或密码错误")
			return
		}

		// 更新最后登录信息（登录时间和IP）
		clientIP := c.ClientIP()
		if err := userService.UpdateLoginInfo(user.ID, clientIP); err != nil {
			// 更新登录信息失败不影响登录流程，只记录日志
			// 在生产环境应该使用专门的日志系统
			println("更新登录信息失败:", err.Error())
		}

		// 提取角色代码列表
		roleCodes := make([]string, 0, len(user.Roles))
		for _, role := range user.Roles {
			roleCodes = append(roleCodes, role.RoleCode)
		}

		// 生成JWT token
		token, err := utils.GenerateToken(user.ID, user.Username, roleCodes)
		if err != nil {
			response.Error(c, 500, "生成Token失败")
			return
		}

		// 构造响应数据
		userInfo := map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"realName": user.RealName,
			"phone":    user.Phone,
			"status":   user.Status,
			"roles":    user.Roles,
		}

		response.SuccessWithData(c, LoginResponse{
			Token:    token,
			UserInfo: userInfo,
		})
	}
}

// GetCurrentUser 获取当前登录用户信息
// 从JWT token中提取用户信息并返回完整用户数据
func GetCurrentUser(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从context中获取用户ID（由JWT中间件设置）
		userID, exists := c.Get("user_id")
		if !exists {
			response.Error(c, 401, "未登录")
			return
		}

		// 查询用户信息
		user, err := userService.GetUserByID(userID.(uint))
		if err != nil {
			response.Error(c, 404, "用户不存在")
			return
		}

		// 构造响应数据
		userInfo := map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"realName": user.RealName,
			"phone":    user.Phone,
			"status":   user.Status,
			"roles":    user.Roles,
		}

		response.SuccessWithData(c, userInfo)
	}
}

// Logout 用户登出
// 实际只是前端清除token，后端无需处理
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Success(c, nil)
	}
}
