package server

import (
	"fmt"
	"net/http"
	"strings"

	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/config"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/middleware"
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/router"

	"github.com/gin-gonic/gin"
)

// Server HTTP服务器结构
type Server struct {
	engine   *gin.Engine
	config   *config.AppConfig
	collector *collector.ResourceCollector
}

// New 创建新的服务器实例
func New(appConfig *config.AppConfig, resourceCollector *collector.ResourceCollector) *Server {
	return &Server{
		config:    appConfig,
		collector: resourceCollector,
	}
}

// Setup 设置服务器配置
func (s *Server) Setup() {
	// 设置 Gin 模式
	if !s.config.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎
	s.engine = gin.Default()
	
	// 添加CORS中间件
	s.engine.Use(middleware.CORS())

	// 设置路由
	s.setupRoutes()
	
	// 设置静态文件服务
	s.setupStaticFiles()
}

// setupRoutes 设置API路由
func (s *Server) setupRoutes() {
	// API 路由 - 先设置API路由，避免被静态文件路由覆盖
	apiGroup := s.engine.Group("/api")
	router.SetupRoutes(apiGroup, s.collector)

	// 添加v1兼容路由
	v1ApiGroup := s.engine.Group("/api/v1")
	router.SetupRoutes(v1ApiGroup, s.collector)
}

// setupStaticFiles 设置静态文件服务
func (s *Server) setupStaticFiles() {
	// 首先检查是否存在构建后的dist目录
	if _, err := http.Dir("./web/dist").Open("/"); err != nil {
		logger.Info("未找到Vue.js构建文件，跳过模板加载专注测试API")
		s.engine.GET("/", func(c *gin.Context) {
			response.OkWithData(gin.H{
				"message": "API服务运行正常",
				"service": "cluster-resource-insight",
			}, c)
		})
	} else {
		// Vue.js SPA模式
		logger.Info("使用Vue.js前端应用")
		s.engine.Static("/assets", "./web/dist/assets")
		s.engine.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

		// SPA首页
		s.engine.GET("/", func(c *gin.Context) {
			c.File("./web/dist/index.html")
		})

		// SPA路由支持 - 所有非API路由都返回index.html
		s.engine.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				response.NotFound("API endpoint not found", c)
				return
			}
			c.File("./web/dist/index.html")
		})
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.App.Port)
	logger.Info("服务器启动在端口 %d", s.config.App.Port)
	logger.Info("访问地址:")
	logger.Info("  - 资源监控: http://localhost:%d", s.config.App.Port)
	logger.Info("  - 集群管理: http://localhost:%d/clusters", s.config.App.Port)
	logger.Info("  - API文档: http://localhost:%d/api/v1/health", s.config.App.Port)

	return s.engine.Run(addr)
}