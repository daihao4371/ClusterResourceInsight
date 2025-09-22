package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/config"
	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/models"
	"cluster-resource-insight/internal/response"
	"cluster-resource-insight/internal/router"

	"github.com/gin-gonic/gin"
)

// checkAndAutoMigrate 检查数据库表是否存在，如果不存在则自动执行迁移
func checkAndAutoMigrate() error {
	db := database.GetDB()

	// 检查关键表是否存在
	logger.Info("正在检查数据库表是否存在...")
	if !db.Migrator().HasTable(&models.ClusterConfig{}) {
		logger.Info("检测到数据库表不存在，正在自动执行迁移...")
		if err := database.MigrateDatabase(); err != nil {
			return fmt.Errorf("自动迁移失败: %v\n\n"+
				"请手动执行以下命令来创建数据库表:\n"+
				"  ./bin/cluster-resource-insight --migrate\n"+
				"或者:\n"+
				"  go run cmd/main.go --migrate", err)
		}
		logger.Info("数据库表自动创建完成")
	} else {
		logger.Info("数据库表已存在，跳过迁移")
	}

	return nil
}

// corsMiddleware 跨域中间件
func corsMiddleware() gin.HandlerFunc {
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

func main() {
	var configPath = flag.String("config", "config.toml", "配置文件路径")
	var migrate = flag.Bool("migrate", false, "执行数据库迁移")
	flag.Parse()

	// 使用标准log进行初始化阶段的日志输出
	log.Println("启动 K8s 多集群资源监控系统...")

	// 加载配置文件
	log.Printf("加载配置文件: %s", *configPath)
	appConfig, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("配置文件加载失败: %v", err)
	}

	// 初始化自定义日志系统
	if err := logger.Init(appConfig.Logging.Level, appConfig.Logging.File); err != nil {
		log.Fatalf("日志系统初始化失败: %v", err)
	}
	logger.Info("日志系统初始化完成，日志文件: %s", appConfig.Logging.File)

	// 从此处开始使用自定义logger
	logger.Info("启动 K8s 多集群资源监控系统...")

	// 初始化数据库连接
	logger.Info("正在初始化数据库连接...")
	dbConfig := &database.DatabaseConfig{
		Host:     appConfig.Database.Host,
		Port:     appConfig.Database.Port,
		Username: appConfig.Database.Username,
		Password: appConfig.Database.Password,
		DBName:   appConfig.Database.Database,
		Charset:  appConfig.Database.Charset,
	}

	if err := database.InitDatabase(dbConfig); err != nil {
		logger.Fatal("数据库初始化失败: %v", err)
	}
	defer database.CloseDatabase()
	defer logger.Close() // 确保程序退出时关闭日志文件

	// 执行数据库迁移
	if *migrate {
		logger.Info("正在执行数据库迁移...")
		if err := database.MigrateDatabase(); err != nil {
			logger.Fatal("数据库迁移失败: %v", err)
		}
		logger.Info("数据库迁移完成")
		return
	}

	// 检查必要的数据库表是否存在，如果不存在则自动执行迁移
	if err := checkAndAutoMigrate(); err != nil {
		logger.Fatal("数据库表检查和自动迁移失败: %v", err)
	}

	// 设置 Gin 模式
	if !appConfig.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建空的资源收集器（多集群模式下会从数据库动态创建）
	logger.Info("使用多集群模式，将从数据库动态创建资源收集器")
	resourceCollector := &collector.ResourceCollector{}

	// 设置路由
	r := gin.Default()
	
	// 添加CORS中间件
	r.Use(corsMiddleware())

	// API 路由 - 先设置API路由，避免被静态文件路由覆盖
	apiGroup := r.Group("/api")
	router.SetupRoutes(apiGroup, resourceCollector)

	// 添加v1兼容路由
	v1ApiGroup := r.Group("/api/v1")
	router.SetupRoutes(v1ApiGroup, resourceCollector)

	// 静态文件服务配置
	// 首先检查是否存在构建后的dist目录
	if _, err := http.Dir("./web/dist").Open("/"); err != nil {
		logger.Info("未找到Vue.js构建文件，跳过模板加载专注测试API")
		r.GET("/", func(c *gin.Context) {
			response.OkWithData(gin.H{
				"message": "API服务运行正常",
				"service": "cluster-resource-insight",
			}, c)
		})
	} else {
		// Vue.js SPA模式
		logger.Info("使用Vue.js前端应用")
		r.Static("/assets", "./web/dist/assets")
		r.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

		// SPA首页
		r.GET("/", func(c *gin.Context) {
			c.File("./web/dist/index.html")
		})

		// SPA路由支持 - 所有非API路由都返回index.html
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				response.NotFound("API endpoint not found", c)
				return
			}
			c.File("./web/dist/index.html")
		})
	}

	// 启动服务器
	addr := fmt.Sprintf(":%d", appConfig.App.Port)
	logger.Info("服务器启动在端口 %d", appConfig.App.Port)
	logger.Info("访问地址:")
	logger.Info("  - 资源监控: http://localhost:%d", appConfig.App.Port)
	logger.Info("  - 集群管理: http://localhost:%d/clusters", appConfig.App.Port)
	logger.Info("  - API文档: http://localhost:%d/api/v1/health", appConfig.App.Port)

	if err := r.Run(addr); err != nil {
		logger.Fatal("服务器启动失败: %v", err)
	}
}
