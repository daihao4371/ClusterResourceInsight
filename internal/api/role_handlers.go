package api

import (
	"net/http"
	"strconv"

	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"

	"github.com/gin-gonic/gin"
)

// CreateRole 创建角色
func CreateRole(roleService *service.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.CreateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}

		role, err := roleService.CreateRole(&req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, role)
	}
}

// GetRoleByID 根据ID获取角色
func GetRoleByID(roleService *service.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "角色ID格式错误")
			return
		}

		role, err := roleService.GetRoleByID(uint(id))
		if err != nil {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Success(c, role)
	}
}

// ListRoles 获取角色列表
func ListRoles(roleService *service.RoleService) gin.HandlerFunc {
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

		roles, total, err := roleService.ListRoles(page, pageSize, keyword, status)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.SuccessWithData(c, gin.H{
			"list":      roles,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

// GetAllActiveRoles 获取所有激活的角色（用于选择器）
func GetAllActiveRoles(roleService *service.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, err := roleService.GetAllActiveRoles()
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, roles)
	}
}

// UpdateRole 更新角色信息
func UpdateRole(roleService *service.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "角色ID格式错误")
			return
		}

		var req service.UpdateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}

		role, err := roleService.UpdateRole(uint(id), &req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, role)
	}
}

// DeleteRole 删除角色
func DeleteRole(roleService *service.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "角色ID格式错误")
			return
		}

		if err := roleService.DeleteRole(uint(id)); err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, gin.H{"message": "角色删除成功"})
	}
}

// GetRoleUsers 获取角色下的用户列表
func GetRoleUsers(roleService *service.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "角色ID格式错误")
			return
		}

		// 解析分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

		// 参数校验
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		users, total, err := roleService.GetRoleUsers(uint(id), page, pageSize)
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

// GetRolePermissions 获取角色的权限列表
func GetRolePermissions(roleService *service.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "角色ID格式错误")
			return
		}

		permissions, err := roleService.GetRolePermissions(uint(id))
		if err != nil {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Success(c, gin.H{"permissions": permissions})
	}
}