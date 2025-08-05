package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"google.golang.org/protobuf/types/known/timestamppb"

	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"github.com/wcygan/todo/backend/internal/config"
	"github.com/wcygan/todo/backend/internal/errors"
)

// MySQLTaskStore provides MySQL-backed storage for tasks
type MySQLTaskStore struct {
	db *sql.DB
}

// NewMySQLTaskStore creates a new MySQLTaskStore instance
func NewMySQLTaskStore(cfg *config.DatabaseConfig) (*MySQLTaskStore, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &MySQLTaskStore{db: db}

	// Run migrations
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// findMigrationsPath returns the path to the migrations directory
func findMigrationsPath() (string, error) {
	// Try the standard path first: internal/store/migrations
	// This works for:
	// - Development: Run from backend/ directory  
	// - Container: Dockerfile copies to ./internal/store/migrations with workdir /app
	migrationsPath := "internal/store/migrations"
	
	if absPath, err := filepath.Abs(migrationsPath); err == nil {
		if _, err := os.Stat(absPath); err == nil {
			return "file://" + absPath, nil
		}
	}
	
	// Fallback: search upward for internal/store/migrations
	// This handles cases where tests run from subdirectories
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	
	dir := wd
	for i := 0; i < 5; i++ { // Search up to 5 levels up
		migrationsPath := filepath.Join(dir, "internal", "store", "migrations")
		if _, err := os.Stat(migrationsPath); err == nil {
			return "file://" + migrationsPath, nil
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached filesystem root
		}
		dir = parent
	}
	
	return "", fmt.Errorf("migrations directory 'internal/store/migrations' not found from working directory: %s", wd)
}

// migrate runs database migrations
func (s *MySQLTaskStore) migrate() error {
	driver, err := mysql.WithInstance(s.db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	migrationsPath, err := findMigrationsPath()
	if err != nil {
		return fmt.Errorf("failed to find migrations path: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance with path %s: %w", migrationsPath, err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *MySQLTaskStore) Close() error {
	return s.db.Close()
}

// GetDB returns the underlying database connection
func (s *MySQLTaskStore) GetDB() *sql.DB {
	return s.db
}

// HealthCheck performs a basic health check on the database connection
func (s *MySQLTaskStore) HealthCheck(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// CreateTask creates a new task with the given description
func (s *MySQLTaskStore) CreateTask(ctx context.Context, description string) (*taskv1.Task, error) {
	if description == "" {
		return nil, fmt.Errorf("task description cannot be empty")
	}

	query := `INSERT INTO tasks (description, completed) VALUES (?, ?)`
	result, err := s.db.ExecContext(ctx, query, description, false)
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to create task")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to get last insert ID")
	}

	// Retrieve the created task to get timestamps
	return s.GetTask(ctx, strconv.FormatInt(id, 10))
}

// GetTask retrieves a task by ID
func (s *MySQLTaskStore) GetTask(ctx context.Context, id string) (*taskv1.Task, error) {
	taskID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid task ID format: %s", id)
	}

	query := `SELECT id, description, completed, created_at, updated_at FROM tasks WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, taskID)

	var task taskv1.Task
	var createdAt, updatedAt time.Time

	err = row.Scan(
		&taskID,
		&task.Description,
		&task.Completed,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound("task", id)
		}
		return nil, errors.InternalWrap(err, "failed to scan task")
	}

	task.Id = strconv.FormatInt(taskID, 10)
	task.CreatedAt = timestamppb.New(createdAt)
	task.UpdatedAt = timestamppb.New(updatedAt)

	return &task, nil
}

// ListTasks returns all tasks in the store
func (s *MySQLTaskStore) ListTasks(ctx context.Context) ([]*taskv1.Task, error) {
	query := `SELECT id, description, completed, created_at, updated_at FROM tasks ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to query tasks")
	}
	defer rows.Close()

	var tasks []*taskv1.Task
	for rows.Next() {
		var task taskv1.Task
		var taskID int64
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&taskID,
			&task.Description,
			&task.Completed,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, errors.InternalWrap(err, "failed to scan task")
		}

		task.Id = strconv.FormatInt(taskID, 10)
		task.CreatedAt = timestamppb.New(createdAt)
		task.UpdatedAt = timestamppb.New(updatedAt)

		tasks = append(tasks, &task)

		// Check for context cancellation during iteration
		select {
		case <-ctx.Done():
			return nil, errors.InternalWrap(ctx.Err(), "context cancelled during task listing")
		default:
		}
	}

	if err := rows.Err(); err != nil {
		return nil, errors.InternalWrap(err, "error iterating over task rows")
	}

	return tasks, nil
}

// UpdateTask updates an existing task
func (s *MySQLTaskStore) UpdateTask(ctx context.Context, id, description string, completed bool) (*taskv1.Task, error) {
	taskID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid task ID format: %s", id)
	}

	// Build dynamic query based on what needs to be updated
	var query string
	var args []interface{}

	if description != "" {
		query = `UPDATE tasks SET description = ?, completed = ?, updated_at = NOW(6) WHERE id = ?`
		args = []interface{}{description, completed, taskID}
	} else {
		query = `UPDATE tasks SET completed = ?, updated_at = NOW(6) WHERE id = ?`
		args = []interface{}{completed, taskID}
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to update task")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return nil, errors.NotFound("task", id)
	}

	// Retrieve the updated task
	return s.GetTask(ctx, id)
}

// DeleteTask removes a task by ID
func (s *MySQLTaskStore) DeleteTask(ctx context.Context, id string) error {
	taskID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid task ID format: %s", id)
	}

	query := `DELETE FROM tasks WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, taskID)
	if err != nil {
		return errors.InternalWrap(err, "failed to delete task")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.InternalWrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.NotFound("task", id)
	}

	return nil
}

// Verify that MySQLTaskStore implements the TaskRepository interface
var _ TaskRepository = (*MySQLTaskStore)(nil)
