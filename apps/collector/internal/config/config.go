package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	API        APIConfig        `mapstructure:"api"`
	Collection CollectionConfig `mapstructure:"collection"`
	Telemetry  TelemetryConfig  `mapstructure:"telemetry"`
	Logging    LoggingConfig    `mapstructure:"logging"`
}

// APIConfig represents API-related configuration
type APIConfig struct {
	Endpoint string        `mapstructure:"endpoint"`
	Timeout  time.Duration `mapstructure:"timeout"`
	Retries  int           `mapstructure:"retries"`
}

// CollectionConfig represents data collection configuration
type CollectionConfig struct {
	Interval    time.Duration `mapstructure:"interval"`
	BatchSize   int           `mapstructure:"batch_size"`
	MaxWorkers  int           `mapstructure:"max_workers"`
	BufferSize  int           `mapstructure:"buffer_size"`
}

// TelemetryConfig represents telemetry configuration
type TelemetryConfig struct {
	Endpoint string  `mapstructure:"endpoint"`
	Enabled  bool    `mapstructure:"enabled"`
	SampleRate float64 `mapstructure:"sample_rate"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/streamforge/collector")

	// Environment variables
	viper.SetEnvPrefix("STREAMFORGE_COLLECTOR")
	viper.AutomaticEnv()

	// Default values
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("api.endpoint", "http://localhost:8080")
	viper.SetDefault("api.timeout", "30s")
	viper.SetDefault("api.retries", 3)

	viper.SetDefault("collection.interval", "1m")
	viper.SetDefault("collection.batch_size", 100)
	viper.SetDefault("collection.max_workers", 4)
	viper.SetDefault("collection.buffer_size", 1000)

	viper.SetDefault("telemetry.endpoint", "http://localhost:4317")
	viper.SetDefault("telemetry.enabled", true)
	viper.SetDefault("telemetry.sample_rate", 1.0)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
} 