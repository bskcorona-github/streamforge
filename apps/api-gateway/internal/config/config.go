package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config はアプリケーション全体の設定を表します
type Config struct {
	Environment     string       `mapstructure:"environment"`
	Version         string       `mapstructure:"version"`
	Port            int          `mapstructure:"port"`
	Database        Database     `mapstructure:"database"`
	Redis           Redis        `mapstructure:"redis"`
	JaegerEndpoint  string       `mapstructure:"jaeger_endpoint"`
	RateLimit       RateLimit    `mapstructure:"rate_limit"`
	Security        Security     `mapstructure:"security"`
	Logging         Logging      `mapstructure:"logging"`
}

// Database はデータベース設定を表します
type Database struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"`
	MaxConns int    `mapstructure:"max_conns"`
}

// Redis はRedis設定を表します
type Redis struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// RateLimit はレート制限設定を表します
type RateLimit struct {
	Enabled bool `mapstructure:"enabled"`
	Limit   int  `mapstructure:"limit"`
	Window  int  `mapstructure:"window"` // 秒単位
}

// Security はセキュリティ設定を表します
type Security struct {
	JWTSecret     string   `mapstructure:"jwt_secret"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	CORSEnabled   bool     `mapstructure:"cors_enabled"`
}

// Logging はログ設定を表します
type Logging struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
}

// Load は設定を読み込みます
func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/streamforge")

	// 環境変数の設定
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// デフォルト値の設定
	setDefaults()

	// 設定ファイルの読み込み
	if err := viper.ReadInConfig(); err != nil {
		// 設定ファイルが見つからない場合は環境変数のみを使用
		fmt.Printf("Warning: Config file not found, using environment variables: %v\n", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal config: %v", err))
	}

	// 設定の検証
	if err := validateConfig(&config); err != nil {
		panic(fmt.Sprintf("Invalid config: %v", err))
	}

	return &config
}

// setDefaults はデフォルト値を設定します
func setDefaults() {
	// アプリケーション設定
	viper.SetDefault("environment", "development")
	viper.SetDefault("version", "1.0.0")
	viper.SetDefault("port", 8080)

	// データベース設定
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "streamforge")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.name", "streamforge")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_conns", 10)

	// Redis設定
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// Jaeger設定
	viper.SetDefault("jaeger_endpoint", "http://localhost:14268/api/traces")

	// レート制限設定
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.limit", 100)
	viper.SetDefault("rate_limit.window", 60)

	// セキュリティ設定
	viper.SetDefault("security.jwt_secret", "your-secret-key")
	viper.SetDefault("security.allowed_origins", []string{"*"})
	viper.SetDefault("security.cors_enabled", true)

	// ログ設定
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output_path", "")
}

// validateConfig は設定の妥当性を検証します
func validateConfig(config *Config) error {
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid port: %d", config.Port)
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Database.Port <= 0 || config.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", config.Database.Port)
	}

	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if config.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if config.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	if config.Redis.Port <= 0 || config.Redis.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d", config.Redis.Port)
	}

	if config.RateLimit.Enabled {
		if config.RateLimit.Limit <= 0 {
			return fmt.Errorf("rate limit must be positive")
		}
		if config.RateLimit.Window <= 0 {
			return fmt.Errorf("rate limit window must be positive")
		}
	}

	return nil
}

// GetDSN はデータベース接続文字列を返します
func (d *Database) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// GetRedisAddr はRedis接続アドレスを返します
func (r *Redis) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
} 