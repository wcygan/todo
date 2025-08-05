package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/wcygan/todo/backend/internal/config"
)

// Manager handles database connections and provides store instances
type Manager struct {
	taskStore TaskRepository
}

// NewManager creates a new store manager with MySQL backend
func NewManager(cfg *config.Config) (*Manager, error) {
	fmt.Println("Connecting to MySQL database...")
	
	// Determine timeout based on environment
	timeout := 120 * time.Second // Default production timeout
	if cfg.IsDevelopment() {
		timeout = 60 * time.Second // Shorter timeout for development
	}
	
	// Wait for database to be available
	if err := WaitForDatabase(&cfg.Database, timeout); err != nil {
		return nil, fmt.Errorf("failed to wait for MySQL database: %w", err)
	}
	
	// Create MySQL task store
	taskStore, err := NewMySQLTaskStore(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL database: %w", err)
	}
	
	envMode := "production"
	if cfg.IsDevelopment() {
		envMode = "development"
	}
	fmt.Printf("Successfully connected to MySQL database in %s mode\n", envMode)

	return &Manager{
		taskStore: taskStore,
	}, nil
}

// TaskStore returns the task repository instance
func (m *Manager) TaskStore() TaskRepository {
	return m.taskStore
}

// Close closes all database connections
func (m *Manager) Close() error {
	if mysqlStore, ok := m.taskStore.(*MySQLTaskStore); ok {
		return mysqlStore.Close()
	}
	return nil
}

// HealthCheck performs a basic health check on the database connection
func (m *Manager) HealthCheck(ctx context.Context) error {
	// Try to list tasks to verify connection is working
	_, err := m.taskStore.ListTasks(ctx)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// GetDB returns the underlying database connection for advanced operations
func (m *Manager) GetDB() (*sql.DB, error) {
	if mysqlStore, ok := m.taskStore.(*MySQLTaskStore); ok {
		return mysqlStore.GetDB(), nil
	}
	return nil, fmt.Errorf("database connection not available")
}

// WaitForDatabase waits for the database to become available
func WaitForDatabase(cfg *config.DatabaseConfig, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	fmt.Printf("Waiting for database at %s:%d to become available (timeout: %v)...\n", 
		cfg.Host, cfg.Port, timeout)

	attempt := 0
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for database to become available after %d attempts", attempt)
		case <-ticker.C:
			attempt++
			
			// Try to create a test connection
			store, err := NewMySQLTaskStore(cfg)
			if err == nil {
				// Test the connection with a simple health check
				healthCtx, healthCancel := context.WithTimeout(context.Background(), 5*time.Second)
				healthErr := store.HealthCheck(healthCtx)
				healthCancel()
				store.Close()
				
				if healthErr == nil {
					fmt.Printf("Database connection successful after %d attempts\n", attempt)
					return nil
				}
				fmt.Printf("Attempt %d: Database connected but health check failed: %v\n", attempt, healthErr)
			} else {
				fmt.Printf("Attempt %d: Waiting for database connection: %v\n", attempt, err)
			}
		}
	}
}