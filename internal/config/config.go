package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/labstack/gommon/bytes"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server              ServerConfig `mapstructure:"server"`
	UploadPartBodyLimit string       `mapstructure:"upload_part_body_limit"`
	UploadSessionSecret string       `mapstructure:"upload_session_secret"`
	AllowedOrigins      []string     `mapstructure:"allowed_origins"`
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
	viper.SetDefault("upload_part_body_limit", "17M")
	viper.SetDefault("upload_session_secret", "")
	viper.SetDefault("allowed_origins", []string{"https://drive.syzroy.com", "http://localhost:3012", "http://127.0.0.1:3012"})

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
	if origins := os.Getenv("CLOUD_DRIVER_ALLOWED_ORIGINS"); origins != "" {
		config.AllowedOrigins = strings.Split(origins, ",")
		for index := range config.AllowedOrigins {
			config.AllowedOrigins[index] = strings.TrimSpace(config.AllowedOrigins[index])
		}
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
	if _, err := bytes.Parse(cfg.UploadPartBodyLimit); err != nil {
		return fmt.Errorf("invalid upload part body limit: %q", cfg.UploadPartBodyLimit)
	}
	if len(cfg.UploadSessionSecret) < 32 {
		return fmt.Errorf("upload_session_secret must be at least 32 characters")
	}
	for _, origin := range cfg.AllowedOrigins {
		if origin != "*" && !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			return fmt.Errorf("invalid allowed origin: %q", origin)
		}
	}

	return nil
}
