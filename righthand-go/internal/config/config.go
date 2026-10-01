package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Services ServicesConfig `mapstructure:"services"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parseTime"`
	Loc             string `mapstructure:"loc"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
	Issuer string `mapstructure:"issuer"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// ServicesConfig 服务配置
type ServicesConfig struct {
	Auth ServiceConfig `mapstructure:"auth"`
	Memo ServiceConfig `mapstructure:"memo"`
	Task ServiceConfig `mapstructure:"task"`
}

// ServiceConfig 单个服务配置
type ServiceConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

var GlobalConfig *Config

// LoadConfig 加载配置文件（支持多种路径）
func LoadConfig(configPath string) (*Config, error) {
	// 尝试多个可能的配置文件路径
	possiblePaths := []string{
		configPath,                  // 传入的路径
		"configs/config.yaml",       // 相对于当前目录
		"../configs/config.yaml",    // 上一级目录的configs
		"../../configs/config.yaml", // 上两级目录的configs
		filepath.Join(getProjectRoot(), "configs", "config.yaml"), // 项目根目录
	}

	var configFile string
	var found bool

	// 查找存在的配置文件
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			found = true
			break
		}
	}

	if !found {
		log.Printf("警告: 未找到配置文件，将使用默认配置")
		// 使用默认值
	} else {
		viper.SetConfigFile(configFile)
		log.Printf("成功加载配置文件: %s", configFile)
	}

	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	viper.SetDefault("server.port", 9080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.host", "127.0.0.1")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.username", "root")
	viper.SetDefault("database.password", "root")
	viper.SetDefault("database.dbname", "righthand")
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("jwt.secret", "righthand-secret-key-2024-very-long-and-secure")
	viper.SetDefault("jwt.expire", 7200)

	if err := viper.ReadInConfig(); err != nil {
		if found {
			log.Printf("警告: 读取配置文件失败 %s: %v", configFile, err)
		}
		// 如果配置文件不存在或读取失败，使用默认值继续
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("无法解析配置文件: %v", err)
	}

	GlobalConfig = &config
	return &config, nil
}

// getProjectRoot 获取项目根目录
func getProjectRoot() string {
	// 尝试通过go.mod文件定位项目根目录
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// 向上查找直到找到go.mod文件
	for dir != filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}

	return ""
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&clientFoundRows=true",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
		c.Charset,
		c.ParseTime,
		url.QueryEscape(c.Loc),
	)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
