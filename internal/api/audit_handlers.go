package api

import (
	"net/http"
	"strconv"
	"time"

	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/service"

	"github.com/gin-gonic/gin"
)

// QueryAuditLogs 查询审计日志
func QueryAuditLogs(auditService *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

		// 解析筛选参数
		var userID *uint
		if userIDStr := c.Query("user_id"); userIDStr != "" {
			if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
				uid := uint(id)
				userID = &uid
			}
		}

		username := c.Query("username")
		operateType := c.Query("operate_type")
		module := c.Query("module")
		resource := c.Query("resource")

		// 解析时间范围
		var startTime, endTime *time.Time
		if startTimeStr := c.Query("start_time"); startTimeStr != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
				startTime = &t
			}
		}
		if endTimeStr := c.Query("end_time"); endTimeStr != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
				endTime = &t
			}
		}

		// 解析耗时范围
		var minDuration, maxDuration *int64
		if minDurationStr := c.Query("min_duration"); minDurationStr != "" {
			if d, err := strconv.ParseInt(minDurationStr, 10, 64); err == nil {
				minDuration = &d
			}
		}
		if maxDurationStr := c.Query("max_duration"); maxDurationStr != "" {
			if d, err := strconv.ParseInt(maxDurationStr, 10, 64); err == nil {
				maxDuration = &d
			}
		}

		// 解析响应码(支持status_code和response_code两个参数名)
		var responseCode *int
		statusCodeStr := c.Query("status_code")
		if statusCodeStr == "" {
			statusCodeStr = c.Query("response_code")
		}
		if statusCodeStr != "" {
			if code, err := strconv.Atoi(statusCodeStr); err == nil {
				responseCode = &code
			}
		}

		// 参数校验
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		logs, total, err := auditService.QueryAuditLogs(
			page, pageSize,
			userID, username, operateType, module, resource,
			startTime, endTime,
			minDuration, maxDuration,
			responseCode,
		)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.SuccessWithData(c, gin.H{
			"list":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

// GetAuditLogByID 根据ID获取审计日志详情
func GetAuditLogByID(auditService *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "审计日志ID格式错误")
			return
		}

		log, err := auditService.GetAuditLogByID(uint(id))
		if err != nil {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Success(c, log)
	}
}

// GetUserAuditLogs 获取指定用户的审计日志
func GetUserAuditLogs(auditService *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "用户ID格式错误")
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

		logs, total, err := auditService.GetUserAuditLogs(uint(userID), page, pageSize)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.SuccessWithData(c, gin.H{
			"list":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}

// GetAuditStatistics 获取审计统计信息
func GetAuditStatistics(auditService *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析时间范围
		var startTime, endTime *time.Time
		if startTimeStr := c.Query("start_time"); startTimeStr != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
				startTime = &t
			}
		}
		if endTimeStr := c.Query("end_time"); endTimeStr != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
				endTime = &t
			}
		}

		stats, err := auditService.GetAuditStatistics(startTime, endTime)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, stats)
	}
}

// GetOperationTrend 获取操作趋势
func GetOperationTrend(auditService *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
		if days < 1 {
			days = 7
		}
		if days > 90 {
			days = 90
		}

		trend, err := auditService.GetOperationTrend(days)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, gin.H{"trend": trend})
	}
}

// CleanupOldAuditLogs 清理旧的审计日志
func CleanupOldAuditLogs(auditService *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))
		if days < 30 {
			response.Error(c, http.StatusBadRequest, "保留天数不能少于30天")
			return
		}

		deletedCount, err := auditService.CleanupOldLogs(days)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Success(c, gin.H{
			"message":       "清理完成",
			"deleted_count": deletedCount,
		})
	}
}

// ExportAuditLogs 导出审计日志
func ExportAuditLogs(auditService *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析筛选参数
		var userID *uint
		if userIDStr := c.Query("user_id"); userIDStr != "" {
			if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
				uid := uint(id)
				userID = &uid
			}
		}

		operateType := c.Query("operate_type")
		module := c.Query("module")

		// 解析时间范围
		var startTime, endTime *time.Time
		if startTimeStr := c.Query("start_time"); startTimeStr != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
				startTime = &t
			}
		}
		if endTimeStr := c.Query("end_time"); endTimeStr != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
				endTime = &t
			}
		}

		csvData, err := auditService.ExportAuditLogs(userID, operateType, module, startTime, endTime)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		// 设置响应头
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename=audit_logs_"+time.Now().Format("20060102150405")+".csv")

		// 写入BOM头，使Excel正确识别UTF-8编码
		c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

		// 写入CSV数据
		for _, row := range csvData {
			for i, col := range row {
				if i > 0 {
					c.Writer.WriteString(",")
				}
				// 处理包含逗号和引号的字段
				if containsSpecialChar(col) {
					c.Writer.WriteString("\"" + escapeQuotes(col) + "\"")
				} else {
					c.Writer.WriteString(col)
				}
			}
			c.Writer.WriteString("\n")
		}
	}
}

// containsSpecialChar 检查字符串是否包含特殊字符
func containsSpecialChar(s string) bool {
	for _, c := range s {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			return true
		}
	}
	return false
}

// escapeQuotes 转义引号
func escapeQuotes(s string) string {
	result := ""
	for _, c := range s {
		if c == '"' {
			result += "\"\""
		} else {
			result += string(c)
		}
	}
	return result
}