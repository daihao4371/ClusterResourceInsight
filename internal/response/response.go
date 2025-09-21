package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构体
type Response struct {
	Code int         `json:"code"` // 响应状态码
	Data interface{} `json:"data"` // 响应数据
	Msg  string      `json:"msg"`  // 响应消息
}

const (
	ERROR   = 7 // 错误状态码
	SUCCESS = 0 // 成功状态码
)

// Result 通用响应方法，构建统一格式的JSON响应
func Result(code int, data interface{}, msg string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}

// Ok 成功响应，无数据返回
func Ok(c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, "操作成功", c)
}

// OkWithMessage 成功响应，自定义消息
func OkWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, message, c)
}

// OkWithData 成功响应，返回数据
func OkWithData(data interface{}, c *gin.Context) {
	Result(SUCCESS, data, "操作成功", c)
}

// OkWithDetailed 成功响应，自定义数据和消息
func OkWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(SUCCESS, data, message, c)
}

// Fail 失败响应，无数据返回
func Fail(c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, "操作失败", c)
}

// FailWithMessage 失败响应，自定义错误消息
func FailWithMessage(message string, c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, message, c)
}

// FailWithDetailed 失败响应，自定义数据和消息
func FailWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(ERROR, data, message, c)
}

// NoAuth 未授权响应，返回401状态码
func NoAuth(message string, c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		Code: ERROR,
		Data: nil,
		Msg:  message,
	})
}

// BadRequest 请求参数错误响应，返回400状态码
func BadRequest(message string, c *gin.Context) {
	c.JSON(http.StatusBadRequest, Response{
		Code: ERROR,
		Data: nil,
		Msg:  message,
	})
}

// NotFound 资源未找到响应，返回404状态码
func NotFound(message string, c *gin.Context) {
	c.JSON(http.StatusNotFound, Response{
		Code: ERROR,
		Data: nil,
		Msg:  message,
	})
}

// InternalServerError 服务器内部错误响应，返回500状态码
func InternalServerError(message string, c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Response{
		Code: ERROR,
		Data: nil,
		Msg:  message,
	})
}