package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the operator configuration
type Config struct {
	Operator   OperatorConfig   `mapstructure:"operator"`
	Kubernetes KubernetesConfig `mapstructure:"kubernetes"`
	Logging    LoggingConfig    `mapstructure:"logging"`
}

// OperatorConfig represents operator-specific configuration
type OperatorConfig struct {
	ReconcilePeriod time.Duration `mapstructure:"reconcile_period"`
	MaxConcurrent   int           `mapstructure:"max_concurrent"`
	LeaderElection  bool          `mapstructure:"leader_election"`
}

// KubernetesConfig represents Kubernetes-related configuration
type KubernetesConfig struct {
	Namespace string `mapstructure:"namespace"`
	WatchAll  bool   `mapstructure:"watch_all"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load loads the configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Set default values
	setDefaults()

	// Read environment variables
	viper.AutomaticEnv()

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
	viper.SetDefault("operator.reconcile_period", "30s")
	viper.SetDefault("operator.max_concurrent", 1)
	viper.SetDefault("operator.leader_election", false)
	viper.SetDefault("kubernetes.namespace", "")
	viper.SetDefault("kubernetes.watch_all", false)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
} 