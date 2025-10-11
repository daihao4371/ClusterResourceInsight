package api

import (
	"net/http"
	"strconv"

	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"

	"github.com/gin-gonic/gin"
)

// CreateUser 创建用户
func CreateUser(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}

		user, err := userService.CreateUser(&req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, user)
	}
}

// GetUserByID 根据ID获取用户
func GetUserByID(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "用户ID格式错误")
			return
		}

		user, err := userService.GetUserByID(uint(id))
		if err != nil {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Success(c, user)
	}
}

// ListUsers 获取用户列表
func ListUsers(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

		// 解析筛选参数
		keyword := c.Query("keyword")
		status := c.Query("status")

		// 参数校验
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		users, total, err := userService.ListUsers(page, pageSize, keyword, status)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.SuccessWithData(c, gin.H{
			"list":      users,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

// UpdateUser 更新用户信息
func UpdateUser(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "用户ID格式错误")
			return
		}

		var req service.UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}

		user, err := userService.UpdateUser(uint(id), &req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, user)
	}
}

// DeleteUser 删除用户
func DeleteUser(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "用户ID格式错误")
			return
		}

		if err := userService.DeleteUser(uint(id)); err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, gin.H{"message": "用户删除成功"})
	}
}

// ChangePassword 修改密码
func ChangePassword(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "用户ID格式错误")
			return
		}

		var req service.ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}

		if err := userService.ChangePassword(uint(id), &req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		response.Success(c, gin.H{"message": "密码修改成功"})
	}
}

// ResetPassword 重置密码（管理员操作）
func ResetPassword(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "用户ID格式错误")
			return
		}

		var req service.ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}

		if err := userService.ResetPassword(uint(id), &req); err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, gin.H{"message": "密码重置成功"})
	}
}