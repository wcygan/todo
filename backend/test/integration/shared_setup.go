package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	taskconnect "buf.build/gen/go/wcygan/todo/connectrpc/go/task/v1/taskv1connect"
	"connectrpc.com/grpcreflect"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/wcygan/todo/backend/internal/config"
	"github.com/wcygan/todo/backend/internal/handler"
	"github.com/wcygan/todo/backend/internal/service"
	"github.com/wcygan/todo/backend/internal/store"
)

var (
	sharedSuite *SharedIntegrationSuite
	suiteMu     sync.Mutex
	suiteOnce   sync.Once
)

// SharedIntegrationSuite provides a shared TestContainer for all integration tests
type SharedIntegrationSuite struct {
	Container *mariadb.MariaDBContainer
	Manager   *store.Manager
	Server    *httptest.Server
	Client    taskconnect.TaskServiceClient
	DB        *sql.DB
	Config    *config.Config
	ctx       context.Context
}

// GetSharedIntegrationSuite returns the singleton shared integration test suite
func GetSharedIntegrationSuite(t *testing.T) *SharedIntegrationSuite {
	t.Helper()
	
	suiteOnce.Do(func() {
		suite, err := setupSharedIntegrationSuite()
		require.NoError(t, err, "Failed to setup shared integration suite")
		sharedSuite = suite
		
		// Note: We don't register cleanup here because t.Cleanup() runs after EACH test
		// The testcontainers framework will handle cleanup automatically when the process exits
	})
	
	// Clean database for this specific test
	require.NoError(t, sharedSuite.CleanDatabase(t), "Failed to clean database")
	
	return sharedSuite
}

// setupSharedIntegrationSuite creates the shared test infrastructure
func setupSharedIntegrationSuite() (*SharedIntegrationSuite, error) {
	ctx := context.Background()
	
	// Create MariaDB container
	container, err := mariadb.Run(ctx,
		"mariadb:11.5",
		mariadb.WithDatabase("shared_integration_test"),
		mariadb.WithUsername("testuser"),
		mariadb.WithPassword("testpass"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create MariaDB container: %w", err)
	}

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "3306")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	// Create configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            host,
			Port:            port.Int(),
			User:            "testuser",
			Password:        "testpass",
			Database:        "shared_integration_test",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 5 * time.Minute,
			SSLMode:         "false",
		},
		Server: config.ServerConfig{
			Port:        0, // Let httptest.Server choose
			ReadTimeout: 30 * time.Second,
		},
	}

	// Create store manager
	manager, err := store.NewManager(cfg)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to create store manager: %w", err)
	}

	// Get direct database connection for cleanup operations
	db, err := manager.GetDB()
	if err != nil {
		manager.Close()
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Create service and handler
	taskStore := manager.TaskStore()
	taskService := service.NewTaskService(taskStore)
	taskHandler := handler.NewTaskHandler(taskService)

	// Create HTTP server
	mux := http.NewServeMux()

	// Register TaskService
	path, serviceHandler := taskconnect.NewTaskServiceHandler(taskHandler)
	mux.Handle(path, serviceHandler)

	// Add reflection support
	reflector := grpcreflect.NewStaticReflector(taskconnect.TaskServiceName)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Add health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		healthCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := manager.HealthCheck(healthCtx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","error":"database_unavailable"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","database":"mysql"}`))
	})

	// Create test server with HTTP/2 support
	server := httptest.NewUnstartedServer(
		h2c.NewHandler(mux, &http2.Server{}),
	)
	server.EnableHTTP2 = true
	server.Start()

	// Create client
	client := taskconnect.NewTaskServiceClient(
		http.DefaultClient,
		server.URL,
	)

	return &SharedIntegrationSuite{
		Container: container,
		Manager:   manager,
		Server:    server,
		Client:    client,
		DB:        db,
		Config:    cfg,
		ctx:       ctx,
	}, nil
}

// CleanDatabase truncates all tables and resets auto-increment
func (s *SharedIntegrationSuite) CleanDatabase(t *testing.T) error {
	t.Helper()
	
	suiteMu.Lock()
	defer suiteMu.Unlock()

	// Check if database connection is still alive
	if err := s.DB.Ping(); err != nil {
		t.Logf("Database connection not available, skipping cleanup: %v", err)
		return nil // Don't fail the test if cleanup can't happen
	}

	// Disable foreign key checks temporarily
	_, err := s.DB.Exec("SET FOREIGN_KEY_CHECKS = 0")
	if err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Truncate tasks table
	_, err = s.DB.Exec("TRUNCATE TABLE tasks")
	if err != nil {
		return fmt.Errorf("failed to truncate tasks table: %w", err)
	}

	// Reset auto-increment
	_, err = s.DB.Exec("ALTER TABLE tasks AUTO_INCREMENT = 1")
	if err != nil {
		return fmt.Errorf("failed to reset auto-increment: %w", err)
	}

	// Re-enable foreign key checks
	_, err = s.DB.Exec("SET FOREIGN_KEY_CHECKS = 1")
	if err != nil {
		return fmt.Errorf("failed to re-enable foreign key checks: %w", err)
	}

	return nil
}

// TearDown cleans up the shared test infrastructure
func (s *SharedIntegrationSuite) TearDown() {
	if s.Server != nil {
		s.Server.Close()
	}
	if s.Manager != nil {
		s.Manager.Close()
	}
	if s.Container != nil {
		s.Container.Terminate(s.ctx)
	}
}

// HealthCheck verifies the shared infrastructure is working
func (s *SharedIntegrationSuite) HealthCheck(t *testing.T) error {
	t.Helper()
	return s.Manager.HealthCheck(s.ctx)
}

// CreateSharedDBConfig creates a database config for tests that need direct MySQL access
func (s *SharedIntegrationSuite) CreateSharedDBConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Host:            s.Config.Database.Host,
		Port:            s.Config.Database.Port,
		User:            s.Config.Database.User,
		Password:        s.Config.Database.Password,
		Database:        s.Config.Database.Database,
		MaxOpenConns:    s.Config.Database.MaxOpenConns,
		MaxIdleConns:    s.Config.Database.MaxIdleConns,
		ConnMaxLifetime: s.Config.Database.ConnMaxLifetime,
		ConnMaxIdleTime: s.Config.Database.ConnMaxIdleTime,
		SSLMode:         s.Config.Database.SSLMode,
	}
}

// CreateIsolatedStoreManager creates a new store manager using the shared database
// for tests that need their own manager instance
func (s *SharedIntegrationSuite) CreateIsolatedStoreManager(t *testing.T) (*store.Manager, error) {
	t.Helper()
	
	cfg := &config.Config{
		Database: *s.CreateSharedDBConfig(),
	}
	
	manager, err := store.NewManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create isolated store manager: %w", err)
	}
	
	// Register cleanup for the isolated manager
	t.Cleanup(func() {
		manager.Close()
	})
	
	return manager, nil
}