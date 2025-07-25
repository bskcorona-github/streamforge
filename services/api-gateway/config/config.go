package config

import (
	"os"
	"strconv"
	"time"
)

// Config represents the application configuration
type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Logging     LoggingConfig
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MaxHeaderBytes  int
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level string
	File  string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Address:        getEnv("SERVER_ADDRESS", ":8080"),
			ReadTimeout:    getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:   getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:    getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
			MaxHeaderBytes: getIntEnv("SERVER_MAX_HEADER_BYTES", 1<<20),
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getIntEnv("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "streamforge"),
			Password: getEnv("POSTGRES_PASSWORD", "streamforge"),
			DBName:   getEnv("POSTGRES_DB", "streamforge"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getIntEnv("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			Expiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			File:  getEnv("LOG_FILE", ""),
		},
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv gets an integer environment variable with a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getDurationEnv gets a duration environment variable with a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
} 