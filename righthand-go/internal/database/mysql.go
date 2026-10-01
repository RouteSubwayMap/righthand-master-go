package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"righthand-go/internal/config"
)

var DB *gorm.DB

// TestDBName 自动测试唯一允许使用的独立测试库名（强约束，见 docs/ai/rules.md §9）
const TestDBName = "righthand_test"

// isTestProcess 判断当前进程是否由 go test 启动（testing.Testing 为主，二进制名后缀兜底）
func isTestProcess() bool {
	if testing.Testing() {
		return true
	}
	name := strings.ToLower(os.Args[0])
	return strings.HasSuffix(name, ".test") || strings.HasSuffix(name, ".test.exe")
}

// InitDB 初始化数据库连接
func InitDB(cfg *config.DatabaseConfig) error {
	// 强约束：测试进程只允许连接独立测试库，禁止触碰开发/生产主库（见 docs/ai/rules.md §9）
	if isTestProcess() && !strings.HasSuffix(strings.ToLower(cfg.DBName), "_test") {
		return fmt.Errorf("自动测试禁止连接非测试库 %q：请使用独立测试库 %s（初始化命令见 docs/ai/rules.md §9）", cfg.DBName, TestDBName)
	}
	dsn := cfg.GetDSN()
	
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	
	if err != nil {
		return fmt.Errorf("无法连接数据库: %v", err)
	}
	
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("无法获取数据库实例: %v", err)
	}
	
	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	
	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}
	
	log.Println("数据库连接成功")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
