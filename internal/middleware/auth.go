package middleware

import (
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
// 从请求头中提取token，验证并注入用户信息到context
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, "未提供认证令牌")
			c.Abort()
			return
		}

		// Bearer token格式: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, 401, "认证令牌格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			response.Error(c, 401, "无效的认证令牌")
			c.Abort()
			return
		}

		// 将用户信息注入到context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("real_name", claims.RealName)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// RequirePermission 权限验证中间件
// 检查用户是否拥有指定权限
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		roles, exists := c.Get("roles")
		if !exists {
			response.Error(c, 403, "权限不足")
			c.Abort()
			return
		}

		roleCodes := roles.([]string)

		// 超级管理员拥有所有权限
		for _, role := range roleCodes {
			if role == "admin" || role == "superadmin" {
				c.Next()
				return
			}
		}

		// TODO: 这里应该查询数据库，检查用户角色是否包含所需权限
		// 由于时间关系，暂时只检查超级管理员
		// 在实际项目中，应该：
		// 1. 从数据库查询用户的所有角色
		// 2. 合并所有角色的权限列表
		// 3. 检查是否包含所需权限

		response.Error(c, 403, "权限不足")
		c.Abort()
	}
}