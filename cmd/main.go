package main

import (
	"flag"
	"log"

	"cluster-resource-insight/internal/collector"
	"cluster-resource-insight/internal/config"
	"cluster-resource-insight/internal/database"
	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/server"
)

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
	defer logger.Close() // 确保程序退出时关闭日志文件

	// 从此处开始使用自定义logger
	logger.Info("启动 K8s 多集群资源监控系统...")

	// 初始化数据库连接
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
	if err := database.CheckAndAutoMigrate(); err != nil {
		logger.Fatal("数据库表检查和自动迁移失败: %v", err)
	}

	// 创建空的资源收集器（多集群模式下会从数据库动态创建）
	logger.Info("使用多集群模式，将从数据库动态创建资源收集器")
	resourceCollector := &collector.ResourceCollector{}

	// 创建并配置服务器
	srv := server.New(appConfig, resourceCollector)
	srv.Setup()

	// 启动服务器
	if err := srv.Start(); err != nil {
		logger.Fatal("服务器启动失败: %v", err)
	}
}
