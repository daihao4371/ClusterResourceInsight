package database

import (
	"fmt"
	"time"

	"cluster-resource-insight/internal/logger"
	"cluster-resource-insight/internal/models"
	
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	Host     string // 数据库主机地址
	Port     int    // 数据库端口
	Username string // 用户名
	Password string // 密码
	DBName   string // 数据库名称
	Charset  string // 字符集
}

// DefaultDatabaseConfig 默认数据库配置
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password", 
		DBName:   "cluster_resource_insight",
		Charset:  "utf8mb4",
	}
}

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(config *DatabaseConfig) error {
	// 构建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.Charset,
	)

	logger.Info("正在连接数据库: %s:%d/%s", config.Host, config.Port, config.DBName)

	// 打开数据库连接
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info), // 启用SQL日志
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 获取底层的sql.DB对象进行连接池配置
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接池失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)           // 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxOpenConns(100)          // 设置打开数据库连接的最大数量
	sqlDB.SetConnMaxLifetime(time.Hour) // 设置连接可复用的最大时间

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	logger.Info("数据库连接成功")
	return nil
}

// MigrateDatabase 执行数据库迁移，创建所有表
func MigrateDatabase() error {
	logger.Info("开始执行数据库迁移...")

	// 自动迁移所有模型
	err := DB.AutoMigrate(
		&models.ClusterConfig{},
		&models.PodMetricsHistory{},
		&models.SystemSettings{},
		&models.AlertRule{},
		&models.AlertHistory{},
	)
	if err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	// 初始化默认系统配置
	if err := initDefaultSettings(); err != nil {
		return fmt.Errorf("初始化默认配置失败: %v", err)
	}

	logger.Info("数据库迁移完成")
	return nil
}

// initDefaultSettings 初始化默认系统配置
func initDefaultSettings() error {
	defaultSettings := []models.SystemSettings{
		{
			Key:         "data_retention_days",
			Value:       "30",
			ValueType:   "int",
			Description: "历史数据保留天数",
		},
		{
			Key:         "default_collect_interval",
			Value:       "30",
			ValueType:   "int", 
			Description: "默认采集间隔（分钟）",
		},
		{
			Key:         "max_concurrent_collections",
			Value:       "10",
			ValueType:   "int",
			Description: "最大并发采集数",
		},
		{
			Key:         "memory_usage_threshold_low",
			Value:       "20",
			ValueType:   "int",
			Description: "内存利用率过低阈值（百分比）",
		},
		{
			Key:         "cpu_usage_threshold_low", 
			Value:       "15",
			ValueType:   "int",
			Description: "CPU利用率过低阈值（百分比）",
		},
		{
			Key:         "alert_enabled",
			Value:       "true",
			ValueType:   "bool",
			Description: "是否启用告警功能",
		},
	}

	// 使用FirstOrCreate确保不重复插入
	for _, setting := range defaultSettings {
		var existingSetting models.SystemSettings
		result := DB.Where("key = ?", setting.Key).First(&existingSetting)
		if result.Error == gorm.ErrRecordNotFound {
			// 配置不存在，创建新配置
			if err := DB.Create(&setting).Error; err != nil {
				return fmt.Errorf("创建默认配置 %s 失败: %v", setting.Key, err)
			}
			logger.Info("创建默认配置: %s = %s", setting.Key, setting.Value)
		}
	}

	return nil
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}