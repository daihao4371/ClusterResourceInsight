package middleware

import (
	"net/http"

	"cluster-resource-insight/internal/logger"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		
		// 设置允许的域名，开发环境允许所有域名
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		// 允许类型校验
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok")
			c.Abort()
			return
		}

		logger.Debug("CORS: origin=%s, method=%s", origin, method)
		c.Next()
	}
}