package config

import (
	"crypto/rand"
	_ "embed"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/internal/log"
	"github.com/leleo886/lopic/models"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

//go:embed config.yaml
var defaultConfigYAML string

// Config 应用程序配置结构体
type Config struct {
	Server         ServerConfig          `mapstructure:"server"`
	Database       DatabaseConfig        `mapstructure:"database"`
	JWT            JWTConfig             `mapstructure:"jwt"`
	Swagger        SwaggerConfig         `mapstructure:"swagger"`
	SystemSettings models.SystemSettings `mapstructure:"systemSettings"`
	Log            LogConfig             `mapstructure:"log"`
}

// LogConfig 日志配置结构体
type LogConfig struct {
	Level         string `mapstructure:"level"`          // 日志级别: debug, info, warn, error
	OutputPath    string `mapstructure:"output_path"`    // 日志输出路径
	MaxSize       int    `mapstructure:"max_size"`       // 单个日志文件最大大小（MB）
	MaxBackups    int    `mapstructure:"max_backups"`    // 保留的最大旧日志文件数量
	MaxAge        int    `mapstructure:"max_age"`        // 旧日志文件的最大留存天数
	Compress      bool   `mapstructure:"compress"`       // 是否压缩旧日志文件
	ConsoleOutput bool   `mapstructure:"console_output"` // 是否同时输出到控制台
}

// ServerConfig 服务器配置结构体
type ServerConfig struct {
	Port         int      `mapstructure:"port"`
	Mode         string   `mapstructure:"mode"`
	StaticPath   string   `mapstructure:"static_path"`
	UploadDir    string   `mapstructure:"upload_dir"`
	AllowOrigins []string `mapstructure:"allowOrigins"`
}

// DatabaseConfig 数据库配置结构体
type DatabaseConfig struct {
	Type      string `mapstructure:"type"`
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	User      string `mapstructure:"user"`
	Password  string `mapstructure:"password"`
	DBName    string `mapstructure:"dbname"`
	Charset   string `mapstructure:"charset"`
	ParseTime bool   `mapstructure:"parseTime"`
	Loc       string `mapstructure:"loc"`
}

// JWTConfig JWT配置结构体
type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	Expire             int    `mapstructure:"expire"`
	RefreshTokenExpire int    `mapstructure:"refresh_token_expire"`
	Issuer             string `mapstructure:"issuer"`
	TokenSecret        string `mapstructure:"token_secret"`
}

// SwaggerConfig Swagger配置结构体
type SwaggerConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Path        string `mapstructure:"path"`
	Title       string `mapstructure:"title"`
	Description string `mapstructure:"description"`
	Version     string `mapstructure:"version"`
}

// 从数据库加载系统设置
func LoadSystemSettingsFromDatabase(db *gorm.DB) (models.SystemSettings, error) {
	var systemSetting models.SystemSetting

	// 从数据库获取单条系统设置记录
	if err := db.First(&systemSetting).Error; err != nil {
		// 如果没有记录，返回默认配置
		if err == gorm.ErrRecordNotFound {
			return models.SystemSettings{}, nil
		}
		return models.SystemSettings{}, err
	}

	return systemSetting.Value, nil
}

// 从配置结构体导入系统设置到数据库
func ImportSystemSettingsToDatabase(db *gorm.DB, settings models.SystemSettings) error {
	var systemSetting models.SystemSetting
	if err := db.First(&systemSetting, 1).Error; err != nil {
		return err
	}
	systemSetting.Value = settings
	return db.Save(&systemSetting).Error
}

// 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 创建配置目录
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Errorf("failed to create config directory: %v", err)
			return nil, fmt.Errorf("failed to create config directory: %v", err)
		}

		// 从内嵌的默认配置创建文件
		if err := os.WriteFile(configPath, []byte(defaultConfigYAML), 0644); err != nil {
			log.Errorf("failed to create default config file: %v", err)
			return nil, fmt.Errorf("failed to create default config file: %v", err)
		}
		log.Infof("created default config file: %s", configPath)
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("failed to read config file: %v", err)
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// 解析配置到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Errorf("failed to unmarshal config: %v", err)
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	if config.Server.Mode == gin.ReleaseMode || !isSecureSecret(config.JWT.Secret) || !isSecureSecret(config.JWT.TokenSecret) {
		if !isSecureSecret(config.JWT.Secret) {
			log.Warn("JWT Secret is insecure, generating random secret")
			config.JWT.Secret = randomString(32)
		}
		if !isSecureSecret(config.JWT.TokenSecret) {
			log.Warn("JWT TokenSecret is insecure, generating random secret")
			config.JWT.TokenSecret = randomString(32)
		}
	}

	return &config, nil
}

// 获取数据库DSN字符串
func (c *DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
			c.User, c.Password, c.Host, c.Port, c.DBName, c.Charset, c.ParseTime, c.Loc)
	case "sqlite":
		return c.DBName
	default:
		return ""
	}
}

func randomString(length int) string {
	// 生成 JWT 密钥
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randomByte, err := secureRandomInt(len(charset))
		if err != nil {
			// 如果生成失败，使用备用方案
			log.Errorf("failed to generate secure random number: %v", err)
			return fallbackRandomString(length)
		}
		b[i] = charset[randomByte]
	}
	return string(b)
}

// 使用 crypto/rand 生成随机整数
func secureRandomInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max must be positive")
	}

	// 计算需要的位数
	n := uint32(max)
	if n&(n-1) == 0 {
		// max 是 2 的幂，可以直接取模
		var b [4]byte
		_, err := rand.Read(b[:])
		if err != nil {
			return 0, err
		}
		return int(binary.BigEndian.Uint32(b[:]) & (n - 1)), nil
	}

	// 拒绝采样法，确保均匀分布
	mask := uint32(1)
	for mask < n {
		mask <<= 1
	}
	mask--

	for {
		var b [4]byte
		_, err := rand.Read(b[:])
		if err != nil {
			return 0, err
		}
		result := binary.BigEndian.Uint32(b[:]) & mask
		if result < n {
			return int(result), nil
		}
	}
}

// fallbackRandomString 备用随机字符串生成
func fallbackRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		// 使用简单的循环获取随机字节
		for {
			var randomByte [1]byte
			_, err := rand.Read(randomByte[:])
			if err != nil {
				continue
			}
			idx := int(randomByte[0]) % len(charset)
			b[i] = charset[idx]
			break
		}
	}
	return string(b)
}

func isSecureSecret(secret string) bool {
	if len(secret) < 16 {
		return false
	}
	unsafePatterns := []string{
		"your-secret",
		"your-token",
		"password",
		"123456",
		"admin",
		"secret",
		"key",
	}
	lowerSecret := strings.ToLower(secret)
	for _, pattern := range unsafePatterns {
		if strings.Contains(lowerSecret, pattern) {
			return false
		}
	}
	return true
}
