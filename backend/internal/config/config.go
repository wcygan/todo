package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server ServerConfig `json:"server"`
	Logger LoggerConfig `json:"logger"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port            int           `json:"port"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	CORS            CORSConfig    `json:"cors"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"` // "json" or "text"
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", "30s"),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", "30s"),
			IdleTimeout:     getEnvAsDuration("SERVER_IDLE_TIMEOUT", "60s"),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", "15s"),
			CORS: CORSConfig{
				AllowedOrigins: getEnvAsStringSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
				AllowedMethods: getEnvAsStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
				AllowedHeaders: getEnvAsStringSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Connect-Protocol-Version", "Connect-Timeout-Ms"}),
			},
		},
		Logger: LoggerConfig{
			Level:  getEnvAsString("LOG_LEVEL", "info"),
			Format: getEnvAsString("LOG_FORMAT", "json"),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server port
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be between 1 and 65535)", c.Server.Port)
	}

	// Validate timeouts
	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("invalid read timeout: %v (must be positive)", c.Server.ReadTimeout)
	}
	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("invalid write timeout: %v (must be positive)", c.Server.WriteTimeout)
	}
	if c.Server.IdleTimeout <= 0 {
		return fmt.Errorf("invalid idle timeout: %v (must be positive)", c.Server.IdleTimeout)
	}
	if c.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("invalid shutdown timeout: %v (must be positive)", c.Server.ShutdownTimeout)
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Logger.Level] {
		return fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error)", c.Logger.Level)
	}

	// Validate log format
	if c.Logger.Format != "json" && c.Logger.Format != "text" {
		return fmt.Errorf("invalid log format: %s (must be 'json' or 'text')", c.Logger.Format)
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return getEnvAsString("ENVIRONMENT", "development") == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return getEnvAsString("ENVIRONMENT", "development") == "production"
}

// Helper functions for environment variable parsing

func getEnvAsString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := time.ParseDuration(valueStr); err == nil {
			return value
		}
	}
	// Parse default value
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 30 * time.Second // fallback
}

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	if valueStr, exists := os.LookupEnv(key); exists {
		// Simple implementation: assume comma-separated values
		// In production, you might want to use a more sophisticated parser
		return []string{valueStr}
	}
	return defaultValue
}