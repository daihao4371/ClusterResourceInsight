package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"cluster-resource-insight/internal/api"
	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/config"
	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/models"

	"github.com/gin-gonic/gin"
)

// checkAndAutoMigrate 检查数据库表是否存在，如果不存在则自动执行迁移
func checkAndAutoMigrate() error {
	db := database.GetDB()
	
	// 检查关键表是否存在
	log.Println("正在检查数据库表是否存在...")
	if !db.Migrator().HasTable(&models.ClusterConfig{}) {
		log.Println("检测到数据库表不存在，正在自动执行迁移...")
		if err := database.MigrateDatabase(); err != nil {
			return fmt.Errorf("自动迁移失败: %v\n\n" +
				"请手动执行以下命令来创建数据库表:\n" +
				"  ./bin/cluster-resource-insight --migrate\n" +
				"或者:\n" +
				"  go run cmd/main.go --migrate", err)
		}
		log.Println("数据库表自动创建完成")
	} else {
		log.Println("数据库表已存在，跳过迁移")
	}
	
	return nil
}

func main() {
	var configPath = flag.String("config", "config.toml", "配置文件路径")
	var migrate = flag.Bool("migrate", false, "执行数据库迁移")
	flag.Parse()

	log.Println("启动 K8s 多集群资源监控系统...")

	// 加载配置文件
	log.Printf("加载配置文件: %s", *configPath)
	appConfig, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("配置文件加载失败: %v", err)
	}

	// 初始化数据库连接
	log.Println("正在初始化数据库连接...")
	dbConfig := &database.DatabaseConfig{
		Host:     appConfig.Database.Host,
		Port:     appConfig.Database.Port,
		Username: appConfig.Database.Username,
		Password: appConfig.Database.Password,
		DBName:   appConfig.Database.Database,
		Charset:  appConfig.Database.Charset,
	}

	if err := database.InitDatabase(dbConfig); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.CloseDatabase()

	// 执行数据库迁移
	if *migrate {
		log.Println("正在执行数据库迁移...")
		if err := database.MigrateDatabase(); err != nil {
			log.Fatalf("数据库迁移失败: %v", err)
		}
		log.Println("数据库迁移完成")
		return
	}

	// 检查必要的数据库表是否存在，如果不存在则自动执行迁移
	if err := checkAndAutoMigrate(); err != nil {
		log.Fatalf("数据库表检查和自动迁移失败: %v", err)
	}

	// 设置 Gin 模式
	if !appConfig.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建空的资源收集器（多集群模式下会从数据库动态创建）
	log.Println("使用多集群模式，将从数据库动态创建资源收集器")
	resourceCollector := &collector.ResourceCollector{}

	// 设置路由
	r := gin.Default()
	
	// 静态文件服务
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// API 路由
	apiGroup := r.Group("/api/v1")
	api.SetupRoutes(apiGroup, resourceCollector)

	// 首页路由
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "K8s 多集群资源监控系统",
		})
	})

	// 集群管理页面路由
	r.GET("/clusters", func(c *gin.Context) {
		c.HTML(http.StatusOK, "clusters.html", gin.H{
			"title": "集群管理",
		})
	})

	// 启动服务器
	addr := fmt.Sprintf(":%d", appConfig.App.Port)
	log.Printf("服务器启动在端口 %d", appConfig.App.Port)
	log.Println("访问地址:")
	log.Printf("  - 资源监控: http://localhost:%d", appConfig.App.Port)
	log.Printf("  - 集群管理: http://localhost:%d/clusters", appConfig.App.Port)
	log.Printf("  - API文档: http://localhost:%d/api/v1/health", appConfig.App.Port)
	
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}