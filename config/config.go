package config

import (
	"os"
	"strconv"
	"time"
)

// Config 包含所有应用配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Captcha  CaptchaConfig
	SMTP     SMTPConfig
	Admin    AdminConfig
	App      AppConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	BodyLimit    int
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int           // 最大连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	ConnMaxIdleTime time.Duration // 连接最大空闲时间
	QueryTimeout    time.Duration // 查询超时时间
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  int
	RefreshTokenExpiry int
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	Enabled     bool
	InstanceURL string
	SiteKey     string
	Secret      string
}

// SMTPConfig SMTP 邮件配置
type SMTPConfig struct {
	Enabled  bool
	Host     string
	Port     int
	UseSSL   bool
	User     string
	Password string
	From     string
	FromName string
}

// AdminConfig 默认管理员配置
type AdminConfig struct {
	Email    string
	Password string
	Username string
}

// AppConfig 应用配置
type AppConfig struct {
	Name             string
	Env              string
	LogLevel         string
	Origin           string
	OriginFrontend   string
	VerifyAPIEnabled bool
	IDCardAPIURL     string
	IDCardAppCode    string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
			BodyLimit:    getIntEnv("BODY_LIMIT", 4*1024*1024), // 4MB
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			Name:            getEnv("DB_NAME", "oauth_server"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 1*time.Hour),
			ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
			QueryTimeout:    getDurationEnv("DB_QUERY_TIMEOUT", 5*time.Second),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", ""),
			AccessTokenExpiry:  getIntEnv("JWT_ACCESS_TOKEN_EXPIRY", 3600),    // 1 hour
			RefreshTokenExpiry: getIntEnv("JWT_REFRESH_TOKEN_EXPIRY", 604800), // 7 days
		},
		Captcha: CaptchaConfig{
			Enabled:     getBoolEnv("CAPTCHA_ENABLED", false),
			InstanceURL: getEnv("CAPTCHA_INSTANCE_URL", ""),
			SiteKey:     getEnv("CAPTCHA_SITE_KEY", ""),
			Secret:      getEnv("CAPTCHA_SECRET", ""),
		},
		SMTP: SMTPConfig{
			Enabled:  getBoolEnv("SMTP_ENABLED", false),
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getIntEnv("SMTP_PORT", 587),
			UseSSL:   getBoolEnv("SMTP_USE_SSL", false),
			User:     getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
			FromName: getEnv("SMTP_FROM_NAME", "OAuth Server"),
		},
		Admin: AdminConfig{
			Email:    getEnv("ADMIN_EMAIL", ""),
			Password: getEnv("ADMIN_PASSWORD", ""),
			Username: getEnv("ADMIN_USERNAME", ""),
		},
		App: AppConfig{
			Name:             getEnv("APP_NAME", "OAuth Server"),
			Env:              getEnv("APP_ENV", "production"),
			LogLevel:         getEnv("LOG_LEVEL", "info"),
			Origin:           getEnv("ORIGIN", "http://localhost:8080"),
			OriginFrontend:   getEnv("ORIGIN_FRONTEND", "http://localhost:8080"),
			VerifyAPIEnabled: getBoolEnv("VERIFY_API_ENABLED", false),
			IDCardAPIURL:     getEnv("IDCARD_API_URL", "https://sfzsmyxb.market.alicloudapi.com/get/idcard/checkV3"),
			IDCardAppCode:    getEnv("IDCARD_APP_CODE", ""),
		},
	}, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv 获取整数类型的环境变量
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getBoolEnv 获取布尔类型的环境变量
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getDurationEnv 获取时间间隔类型的环境变量
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
