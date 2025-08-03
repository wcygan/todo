package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear environment variables
	clearEnvVars()
	
	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Test default values
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 60*time.Second, config.Server.IdleTimeout)
	assert.Equal(t, 15*time.Second, config.Server.ShutdownTimeout)
	assert.Equal(t, []string{"*"}, config.Server.CORS.AllowedOrigins)
	assert.Equal(t, "info", config.Logger.Level)
	assert.Equal(t, "json", config.Logger.Format)
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	setEnvVars(map[string]string{
		"SERVER_PORT":            "9090",
		"SERVER_READ_TIMEOUT":    "45s",
		"SERVER_WRITE_TIMEOUT":   "45s",
		"SERVER_IDLE_TIMEOUT":    "90s",
		"SERVER_SHUTDOWN_TIMEOUT": "20s",
		"LOG_LEVEL":              "debug",
		"LOG_FORMAT":             "text",
	})
	defer clearEnvVars()
	
	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)
	
	// Test environment values
	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, 45*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 45*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 90*time.Second, config.Server.IdleTimeout)
	assert.Equal(t, 20*time.Second, config.Server.ShutdownTimeout)
	assert.Equal(t, "debug", config.Logger.Level)
	assert.Equal(t, "text", config.Logger.Format)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_config",
			config: &Config{
				Server: ServerConfig{
					Port:            8080,
					ReadTimeout:     30 * time.Second,
					WriteTimeout:    30 * time.Second,
					IdleTimeout:     60 * time.Second,
					ShutdownTimeout: 15 * time.Second,
				},
				Logger: LoggerConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid_port_zero",
			config: &Config{
				Server: ServerConfig{
					Port:            0,
					ReadTimeout:     30 * time.Second,
					WriteTimeout:    30 * time.Second,
					IdleTimeout:     60 * time.Second,
					ShutdownTimeout: 15 * time.Second,
				},
				Logger: LoggerConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid_port_too_high",
			config: &Config{
				Server: ServerConfig{
					Port:            70000,
					ReadTimeout:     30 * time.Second,
					WriteTimeout:    30 * time.Second,
					IdleTimeout:     60 * time.Second,
					ShutdownTimeout: 15 * time.Second,
				},
				Logger: LoggerConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid_log_level",
			config: &Config{
				Server: ServerConfig{
					Port:            8080,
					ReadTimeout:     30 * time.Second,
					WriteTimeout:    30 * time.Second,
					IdleTimeout:     60 * time.Second,
					ShutdownTimeout: 15 * time.Second,
				},
				Logger: LoggerConfig{
					Level:  "invalid",
					Format: "json",
				},
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
		{
			name: "invalid_log_format",
			config: &Config{
				Server: ServerConfig{
					Port:            8080,
					ReadTimeout:     30 * time.Second,
					WriteTimeout:    30 * time.Second,
					IdleTimeout:     60 * time.Second,
					ShutdownTimeout: 15 * time.Second,
				},
				Logger: LoggerConfig{
					Level:  "info",
					Format: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid log format",
		},
		{
			name: "invalid_read_timeout",
			config: &Config{
				Server: ServerConfig{
					Port:            8080,
					ReadTimeout:     0,
					WriteTimeout:    30 * time.Second,
					IdleTimeout:     60 * time.Second,
					ShutdownTimeout: 15 * time.Second,
				},
				Logger: LoggerConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr: true,
			errMsg:  "invalid read timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "default_development",
			envValue: "",
			expected: true,
		},
		{
			name:     "explicit_development",
			envValue: "development",
			expected: true,
		},
		{
			name:     "production",
			envValue: "production",
			expected: false,
		},
		{
			name:     "staging",
			envValue: "staging",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			if tt.envValue != "" {
				os.Setenv("ENVIRONMENT", tt.envValue)
			}
			defer clearEnvVars()
			
			config := &Config{}
			assert.Equal(t, tt.expected, config.IsDevelopment())
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "default_development",
			envValue: "",
			expected: false,
		},
		{
			name:     "development",
			envValue: "development",
			expected: false,
		},
		{
			name:     "production",
			envValue: "production",
			expected: true,
		},
		{
			name:     "staging",
			envValue: "staging",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars()
			if tt.envValue != "" {
				os.Setenv("ENVIRONMENT", tt.envValue)
			}
			defer clearEnvVars()
			
			config := &Config{}
			assert.Equal(t, tt.expected, config.IsProduction())
		})
	}
}

// Helper functions

func setEnvVars(vars map[string]string) {
	for key, value := range vars {
		os.Setenv(key, value)
	}
}

func clearEnvVars() {
	envVars := []string{
		"SERVER_PORT",
		"SERVER_READ_TIMEOUT",
		"SERVER_WRITE_TIMEOUT",
		"SERVER_IDLE_TIMEOUT",
		"SERVER_SHUTDOWN_TIMEOUT",
		"CORS_ALLOWED_ORIGINS",
		"CORS_ALLOWED_METHODS",
		"CORS_ALLOWED_HEADERS",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"ENVIRONMENT",
	}
	
	for _, key := range envVars {
		os.Unsetenv(key)
	}
}