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

	// API 路由 - 先设置API路由，避免被静态文件路由覆盖
	apiGroup := r.Group("/api")
	router.SetupRoutes(apiGroup, resourceCollector)

	// 添加v1兼容路由
	v1ApiGroup := r.Group("/api/v1")
	router.SetupRoutes(v1ApiGroup, resourceCollector)

	// 静态文件服务配置
	// 首先检查是否存在构建后的dist目录
	if _, err := http.Dir("./web/dist").Open("/"); err != nil {
		// 如果没有dist目录，说明是开发模式，使用web-legacy作为备用
		logger.Info("未找到Vue.js构建文件，使用传统web页面作为备用")
		r.Static("/static", "./web-legacy/static")
		r.LoadHTMLGlob("web-legacy/templates/*")

		// 传统模式的路由
		r.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"title": "K8s 多集群资源监控系统",
			})
		})

		r.GET("/clusters", func(c *gin.Context) {
			c.HTML(http.StatusOK, "clusters.html", gin.H{
				"title": "集群管理",
			})
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
