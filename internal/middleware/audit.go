package middleware

import (
	"bytes"
	"cluster-resource-insight/internal/service"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware 审计日志中间件
// 记录所有API请求的详细信息，用于安全审计和操作追踪
func AuditMiddleware(auditService *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 读取请求体（需要保存以便后续handler使用）
		var requestBody string
		if c.Request.Body != nil && c.Request.Method != "GET" {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// 重新设置请求体，使后续handler可以读取
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 使用响应写入器包装以捕获响应状态码
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			statusCode:     200,
		}
		c.Writer = writer

		// 继续处理请求
		c.Next()

		// 请求处理完成后记录审计日志
		duration := time.Since(startTime).Milliseconds()

		// 从上下文获取认证用户信息
		var userID uint
		var username string
		var realName string
		if userIDValue, exists := c.Get("user_id"); exists {
			userID = userIDValue.(uint)
		}
		if usernameValue, exists := c.Get("username"); exists {
			username = usernameValue.(string)
		}
		if realNameValue, exists := c.Get("real_name"); exists {
			realName = realNameValue.(string)
		}

		// 确定操作类型
		operateType := determineOperateType(c.Request.Method, c.Request.URL.Path)

		// 确定模块
		module := determineModule(c.Request.URL.Path)

		// 脱敏处理请求体（移除敏感信息）
		requestBody = sanitizeRequestBody(requestBody)

		// 构建审计日志请求
		auditLogReq := &service.CreateAuditLogRequest{
			UserID:       userID,
			Username:     username,
			RealName:     realName,
			OperateType:  operateType,
			Module:       module,
			Resource:     c.Request.URL.Path,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			RequestBody:  requestBody,
			ResponseCode: writer.statusCode,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Duration:     duration,
		}

		// 如果请求失败，记录错误信息
		if len(c.Errors) > 0 {
			auditLogReq.ErrorMsg = c.Errors.String()
		}

		// 异步保存审计日志（避免影响请求响应时间）
		go func() {
			if err := auditService.CreateAuditLog(auditLogReq); err != nil {
				// 日志记录失败不应影响主业务流程，只记录错误
				println("审计日志记录失败:", err.Error())
			}
		}()
	}
}

// responseWriter 自定义响应写入器，用于捕获HTTP状态码
type responseWriter struct {
	gin.ResponseWriter
	statusCode int
}

// WriteHeader 重写WriteHeader方法以捕获状态码
func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// determineOperateType 根据HTTP方法和路径判断操作类型
func determineOperateType(method, path string) string {
	// 登录/登出特殊处理
	if strings.Contains(path, "/login") {
		return "login"
	}
	if strings.Contains(path, "/logout") {
		return "logout"
	}

	// 根据HTTP方法判断
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	case "GET":
		return "query"
	default:
		return "unknown"
	}
}

// determineModule 根据URL路径判断所属模块
func determineModule(path string) string {
	if strings.Contains(path, "/users") {
		return "user"
	}
	if strings.Contains(path, "/roles") {
		return "role"
	}
	if strings.Contains(path, "/apis") {
		return "api"
	}
	if strings.Contains(path, "/clusters") {
		return "cluster"
	}
	if strings.Contains(path, "/pods") {
		return "pod"
	}
	if strings.Contains(path, "/schedules") || strings.Contains(path, "/tasks") {
		return "schedule"
	}
	if strings.Contains(path, "/alerts") {
		return "alert"
	}
	if strings.Contains(path, "/audit") {
		return "audit"
	}
	if strings.Contains(path, "/auth") {
		return "auth"
	}
	return "system"
}

// sanitizeRequestBody 脱敏处理请求体，移除敏感字段
func sanitizeRequestBody(body string) string {
	if body == "" {
		return ""
	}

	// 尝试解析为JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// 如果不是JSON，直接返回（可能需要其他格式的处理）
		return body
	}

	// 定义需要脱敏的字段列表
	sensitiveFields := []string{
		"password", "passwd", "pwd",
		"token", "access_token", "refresh_token",
		"secret", "api_key", "apikey",
		"old_password", "new_password",
	}

	// 替换敏感字段值
	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			data[field] = "******"
		}
	}

	// 重新序列化为JSON
	sanitized, err := json.Marshal(data)
	if err != nil {
		return body
	}

	return string(sanitized)
}
