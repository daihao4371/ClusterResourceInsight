package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

// AppConfig 应用配置结构
type AppConfig struct {
	Database   DatabaseConfig   `mapstructure:"database"`
	App        ApplicationConfig `mapstructure:"app"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Alert      AlertConfig      `mapstructure:"alert"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
}

// ApplicationConfig 应用配置
type ApplicationConfig struct {
	Port  int  `mapstructure:"port"`
	Debug bool `mapstructure:"debug"`
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Key string `mapstructure:"key"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	DefaultCollectInterval    int `mapstructure:"default_collect_interval"`
	MaxConcurrentCollections int `mapstructure:"max_concurrent_collections"`
	DataRetentionDays        int `mapstructure:"data_retention_days"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled                  bool `mapstructure:"enabled"`
	MemoryUsageThresholdLow int  `mapstructure:"memory_usage_threshold_low"`
	CPUUsageThresholdLow    int  `mapstructure:"cpu_usage_threshold_low"`
}

var AppConf *AppConfig

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*AppConfig, error) {
	// 设置配置文件路径
	if configPath == "" {
		configPath = "config.toml"
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 设置 viper 配置
	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	// 设置环境变量前缀
	viper.SetEnvPrefix("CLUSTER_MONITOR")
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析配置到结构体
	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %v", err)
	}

	// 环境变量覆盖（敏感信息）
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}
	if encryptionKey := os.Getenv("ENCRYPTION_KEY"); encryptionKey != "" {
		config.Encryption.Key = encryptionKey
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	AppConf = &config
	log.Printf("配置加载成功: %s", configPath)
	return &config, nil
}

// validateConfig 验证配置的有效性
func validateConfig(config *AppConfig) error {
	// 验证数据库配置
	if config.Database.Host == "" {
		return fmt.Errorf("数据库主机地址不能为空")
	}
	if config.Database.Port <= 0 || config.Database.Port > 65535 {
		return fmt.Errorf("数据库端口无效: %d", config.Database.Port)
	}
	if config.Database.Username == "" {
		return fmt.Errorf("数据库用户名不能为空")
	}
	if config.Database.Database == "" {
		return fmt.Errorf("数据库名称不能为空")
	}

	// 验证应用配置
	if config.App.Port <= 0 || config.App.Port > 65535 {
		return fmt.Errorf("应用端口无效: %d", config.App.Port)
	}

	// 验证加密密钥长度
	if len(config.Encryption.Key) != 32 {
		return fmt.Errorf("加密密钥长度必须为32字节，当前长度: %d", len(config.Encryption.Key))
	}

	// 验证监控配置
	if config.Monitoring.DefaultCollectInterval <= 0 {
		return fmt.Errorf("默认采集间隔必须大于0")
	}
	if config.Monitoring.MaxConcurrentCollections <= 0 {
		return fmt.Errorf("最大并发采集数必须大于0")
	}
	if config.Monitoring.DataRetentionDays <= 0 {
		return fmt.Errorf("数据保留天数必须大于0")
	}

	// 验证告警阈值
	if config.Alert.MemoryUsageThresholdLow < 0 || config.Alert.MemoryUsageThresholdLow > 100 {
		return fmt.Errorf("内存利用率阈值必须在0-100之间")
	}
	if config.Alert.CPUUsageThresholdLow < 0 || config.Alert.CPUUsageThresholdLow > 100 {
		return fmt.Errorf("CPU利用率阈值必须在0-100之间")
	}

	return nil
}

// GetDatabaseConfig 获取数据库配置
func GetDatabaseConfig() *DatabaseConfig {
	if AppConf == nil {
		return nil
	}
	return &AppConf.Database
}

// GetAppConfig 获取应用配置
func GetAppConfig() *ApplicationConfig {
	if AppConf == nil {
		return nil
	}
	return &AppConf.App
}

// GetEncryptionConfig 获取加密配置
func GetEncryptionConfig() *EncryptionConfig {
	if AppConf == nil {
		return nil
	}
	return &AppConf.Encryption
}

// GetMonitoringConfig 获取监控配置
func GetMonitoringConfig() *MonitoringConfig {
	if AppConf == nil {
		return nil
	}
	return &AppConf.Monitoring
}

// GetAlertConfig 获取告警配置
func GetAlertConfig() *AlertConfig {
	if AppConf == nil {
		return nil
	}
	return &AppConf.Alert
}