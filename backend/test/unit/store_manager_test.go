package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"

	"github.com/wcygan/todo/backend/internal/config"
	"github.com/wcygan/todo/backend/internal/store"
)

func TestStoreManager_Unit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping store manager unit tests in short mode")
	}

	t.Run("NewManager_SuccessfulConnection", func(t *testing.T) {
		// Setup test MariaDB
		ctx := context.Background()
		container, dbConfig := setupTestMariaDB(t, ctx)
		defer container.Terminate(ctx)

		cfg := &config.Config{
			Database: *dbConfig,
		}

		// Test manager creation
		manager, err := store.NewManager(cfg)
		require.NoError(t, err)
		require.NotNil(t, manager)

		// Test health check
		healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		err = manager.HealthCheck(healthCtx)
		assert.NoError(t, err)

		// Test task store access
		taskStore := manager.TaskStore()
		assert.NotNil(t, taskStore)

		// Cleanup
		err = manager.Close()
		assert.NoError(t, err)
	})

	t.Run("NewManager_DatabaseConnectionTimeout", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Host:            "nonexistent-host",
				Port:            3306,
				User:            "testuser",
				Password:        "testpass",
				Database:        "testdb",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				SSLMode:         "false",
			},
		}

		// Should fail quickly in tests
		start := time.Now()
		_, err := store.NewManager(cfg)
		duration := time.Since(start)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to wait for MySQL database")
		assert.Less(t, duration, 65*time.Second) // Should timeout around 60s for development
	})

	t.Run("WaitForDatabase_Success", func(t *testing.T) {
		ctx := context.Background()
		container, dbConfig := setupTestMariaDB(t, ctx)
		defer container.Terminate(ctx)

		err := store.WaitForDatabase(dbConfig, 30*time.Second)
		assert.NoError(t, err)
	})

	t.Run("WaitForDatabase_Timeout", func(t *testing.T) {
		dbConfig := &config.DatabaseConfig{
			Host:            "nonexistent-host",
			Port:            3306,
			User:            "testuser",
			Password:        "testpass",
			Database:        "testdb",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 5 * time.Minute,
			SSLMode:         "false",
		}

		start := time.Now()
		err := store.WaitForDatabase(dbConfig, 2*time.Second)
		duration := time.Since(start)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "timeout waiting for database")
		assert.GreaterOrEqual(t, duration, 2*time.Second)
		assert.Less(t, duration, 3*time.Second)
	})

	t.Run("HealthCheck_DatabaseDown", func(t *testing.T) {
		// Start container, create manager, then stop container
		ctx := context.Background()
		container, dbConfig := setupTestMariaDB(t, ctx)

		cfg := &config.Config{
			Database: *dbConfig,
		}

		manager, err := store.NewManager(cfg)
		require.NoError(t, err)

		// Stop the database
		err = container.Stop(ctx, nil)
		require.NoError(t, err)

		// Health check should fail
		time.Sleep(100 * time.Millisecond) // Give it time to stop
		healthCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		
		err = manager.HealthCheck(healthCtx)
		assert.Error(t, err)

		manager.Close()
		container.Terminate(ctx)
	})
}

func TestStoreManager_ConfigValidation(t *testing.T) {
	testCases := []struct {
		name      string
		config    *config.Config
		shouldErr bool
		errMsg    string
	}{
		{
			name: "ValidConfig",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Host:            "localhost",
					Port:            3306,
					User:            "testuser",
					Password:        "testpass",
					Database:        "testdb",
					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxLifetime: 5 * time.Minute,
					ConnMaxIdleTime: 5 * time.Minute,
					SSLMode:         "false",
				},
			},
			shouldErr: false,
		},
		{
			name: "InvalidConfig_EmptyHost",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Host:     "",
					Port:     3306,
					User:     "testuser",
					Password: "testpass",
					Database: "testdb",
				},
			},
			shouldErr: true,
			errMsg:    "database host cannot be empty",
		},
		{
			name: "InvalidConfig_InvalidPort",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Host:     "localhost",
					Port:     0,
					User:     "testuser",
					Password: "testpass",
					Database: "testdb",
				},
			},
			shouldErr: true,
			errMsg:    "invalid database port",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.shouldErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to setup test MariaDB
func setupTestMariaDB(t *testing.T, ctx context.Context) (*mariadb.MariaDBContainer, *config.DatabaseConfig) {
	t.Helper()

	container, err := mariadb.Run(ctx,
		"mariadb:11.5",
		mariadb.WithDatabase("testdb"),
		mariadb.WithUsername("testuser"),
		mariadb.WithPassword("testpass"),
	)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)

	dbConfig := &config.DatabaseConfig{
		Host:            host,
		Port:            port.Int(),
		User:            "testuser",
		Password:        "testpass",
		Database:        "testdb",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		SSLMode:         "false",
	}

	return container, dbConfig
}