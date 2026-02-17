package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"github.com/leleo886/lopic/internal/config"
	"github.com/leleo886/lopic/internal/log"
	"gorm.io/driver/mysql"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接实例
var DB *gorm.DB

// Connect 连接数据库
func Connect(config *config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	// 根据配置选择数据库驱动
	switch config.Type {
	case "mysql":
		dialector = mysql.Open(config.GetDSN())
	case "sqlite":
		// 创建SQLite数据库文件（如果不存在）
		if _, err := os.Stat(config.DBName); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(config.DBName), 0755); err != nil {
				return nil, fmt.Errorf("failed to create sqlite directory: %v", err)
			}
		}
		dialector = sqlite.Open(config.GetDSN())
	default:
		log.Fatalf("Not supported database type: %s", config.Type)
		return nil, fmt.Errorf("not supported database type: %s", config.Type)
	}

	// 配置GORM日志级别
	logLevel := logger.Info
	if config.Type == "sqlite" {
		logLevel = logger.Error
	}

	// 创建自定义GORM logger，将日志输出到我们的zap日志系统
	customLogger := logger.New(
		&gormLoggerWriter{},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // 慢SQL阈值
			LogLevel:                  logLevel,               // 日志级别
			IgnoreRecordNotFoundError: false,                  // 不忽略ErrRecordNotFound错误
			Colorful:                  false,                  // 禁用彩色输出
		},
	)

	// 连接数据库
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: customLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	// 设置全局DB实例
	DB = db

	log.Infof("Database connected successfully (Type: %s)", config.Type)
	return db, nil
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	return DB
}

// gormLoggerWriter 实现gorm.io/gorm/logger.Writer接口，将GORM日志输出到自定义日志系统
type gormLoggerWriter struct{}

// Printf 实现logger.Writer接口，将GORM日志输出到自定义日志系统
func (w *gormLoggerWriter) Printf(format string, args ...interface{}) {
	// 将GORM日志写入到自定义日志系统
	log.Infof(format, args...)
}
