package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Drivers  DriversConfig  `mapstructure:"drivers"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// DatabaseConfig contains database-related configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// DriversConfig contains all driver configurations
type DriversConfig struct {
	Drive115 Drive115Config `mapstructure:"115driver"`
}

// Drive115Config contains 115drive specific configuration
type Drive115Config struct {
	UID  string `mapstructure:"uid"`
	CID  string `mapstructure:"cid"`
	SEID string `mapstructure:"seid"`
	KID  string `mapstructure:"kid"`
}

// Load loads and validates the application configuration
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set default values
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.name", "cloud_driver")
	viper.SetDefault("database.ssl_mode", "disable")

	// Environment variable support
	viper.SetEnvPrefix("CLOUD_DRIVER")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using defaults and environment variables: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validate checks if the configuration is valid
func validate(cfg *Config) error {
	// Database validation
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host not configured")
	}
	if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", cfg.Database.Port)
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database user not configured")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database name not configured")
	}

	// Server validation
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	return nil
}
