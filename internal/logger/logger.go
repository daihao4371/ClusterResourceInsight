package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// levelNames 日志级别名称映射
var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger 统一日志管理器
type Logger struct {
	level      LogLevel
	fileWriter io.Writer
	consoleLog *log.Logger
	fileLog    *log.Logger
	mu         sync.RWMutex
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init 初始化日志系统
func Init(logLevel string, logFile string) error {
	var err error
	once.Do(func() {
		defaultLogger, err = NewLogger(logLevel, logFile)
	})
	return err
}

// NewLogger 创建新的日志实例
func NewLogger(logLevel string, logFile string) (*Logger, error) {
	level := parseLogLevel(logLevel)
	
	logger := &Logger{
		level:      level,
		consoleLog: log.New(os.Stdout, "", log.LstdFlags),
	}

	// 如果指定了日志文件，创建文件写入器
	if logFile != "" {
		if err := logger.setupFileLogger(logFile); err != nil {
			return nil, fmt.Errorf("设置文件日志失败: %v", err)
		}
	}

	return logger, nil
}

// setupFileLogger 设置文件日志输出
func (l *Logger) setupFileLogger(logFile string) error {
	// 确保日志目录存在
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	l.fileWriter = file
	l.fileLog = log.New(file, "", log.LstdFlags)
	return nil
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// formatMessage 格式化日志消息
func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("[%s] %s", levelNames[level], message)
}

// log 内部日志记录方法
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	message := l.formatMessage(level, format, args...)

	// 输出到控制台
	l.consoleLog.Print(message)

	// 输出到文件（如果配置了文件输出）
	if l.fileLog != nil {
		l.fileLog.Print(message)
	}

	// FATAL级别直接退出程序
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug 调试级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info 信息级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn 警告级别日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error 错误级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal 致命错误级别日志
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

// Println 兼容标准log.Println的方法
func (l *Logger) Println(args ...interface{}) {
	l.log(INFO, "%s", fmt.Sprint(args...))
}

// Printf 兼容标准log.Printf的方法
func (l *Logger) Printf(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// 全局日志方法
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Fatal(format, args...)
	}
}

func Println(args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Println(args...)
	}
}

func Printf(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Printf(format, args...)
	}
}

// GetLogger 获取默认日志实例
func GetLogger() *Logger {
	return defaultLogger
}

// Close 关闭日志文件
func Close() error {
	if defaultLogger != nil && defaultLogger.fileWriter != nil {
		if closer, ok := defaultLogger.fileWriter.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}