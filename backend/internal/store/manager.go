package store

import (
	"context"
	"fmt"
	"time"

	"github.com/wcygan/todo/backend/internal/config"
)

// Manager handles database connections and provides store instances
type Manager struct {
	taskStore TaskRepository
}

// NewManager creates a new store manager with the appropriate backend
func NewManager(cfg *config.Config) (*Manager, error) {
	var taskStore TaskRepository
	var err error

	if cfg.IsDevelopment() {
		// In development, try to connect to MySQL first, fallback to in-memory
		taskStore, err = NewMySQLTaskStore(&cfg.Database)
		if err != nil {
			fmt.Printf("Failed to connect to MySQL in development, falling back to in-memory store: %v\n", err)
			taskStore = New() // fallback to in-memory store
		} else {
			fmt.Println("Connected to MySQL database successfully")
		}
	} else {
		// In production, MySQL is required
		taskStore, err = NewMySQLTaskStore(&cfg.Database)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MySQL database: %w", err)
		}
		fmt.Println("Connected to MySQL database in production mode")
	}

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

// WaitForDatabase waits for the database to become available
func WaitForDatabase(cfg *config.DatabaseConfig, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for database to become available")
		case <-ticker.C:
			store, err := NewMySQLTaskStore(cfg)
			if err == nil {
				store.Close()
				return nil
			}
			fmt.Printf("Waiting for database to become available... %v\n", err)
		}
	}
}