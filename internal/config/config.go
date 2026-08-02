package config

import (
	"fmt"
	"log"

	"github.com/labstack/gommon/bytes"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server          ServerConfig `mapstructure:"server"`
	UploadBodyLimit string       `mapstructure:"upload_body_limit"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
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
	viper.SetDefault("upload_body_limit", "20G")

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
	// Server validation
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}
	if _, err := bytes.Parse(cfg.UploadBodyLimit); err != nil {
		return fmt.Errorf("invalid upload body limit: %q", cfg.UploadBodyLimit)
	}

	return nil
}
