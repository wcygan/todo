package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"

	"github.com/wcygan/todo/backend/internal/config"
)

func TestMySQLTaskStore_Integration(t *testing.T) {
	// Skip integration tests in short mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := context.Background()

	// Start MariaDB container
	mariadbContainer, err := mariadb.Run(ctx,
		"mariadb:11.5",
		mariadb.WithDatabase("testdb"),
		mariadb.WithUsername("testuser"),
		mariadb.WithPassword("testpass"),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, mariadbContainer.Terminate(ctx))
	}()

	// Get connection details
	host, err := mariadbContainer.Host(ctx)
	require.NoError(t, err)

	port, err := mariadbContainer.MappedPort(ctx, "3306")
	require.NoError(t, err)

	// Create database config
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

	// Create store
	store, err := NewMySQLTaskStore(dbConfig)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, store.Close())
	}()

	// Run comprehensive tests
	t.Run("CreateTask", func(t *testing.T) {
		testCreateTask(t, store)
	})

	t.Run("GetTask", func(t *testing.T) {
		testGetTask(t, store)
	})

	t.Run("ListTasks", func(t *testing.T) {
		testListTasks(t, store)
	})

	t.Run("UpdateTask", func(t *testing.T) {
		testUpdateTask(t, store)
	})

	t.Run("DeleteTask", func(t *testing.T) {
		testDeleteTask(t, store)
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		testConcurrentOperations(t, store)
	})
}

func testCreateTask(t *testing.T, store TaskRepository) {
	ctx := context.Background()

	// Test successful task creation
	task, err := store.CreateTask(ctx, "Test task")
	require.NoError(t, err)
	assert.NotEmpty(t, task.Id)
	assert.Equal(t, "Test task", task.Description)
	assert.False(t, task.Completed)
	assert.NotNil(t, task.CreatedAt)
	assert.NotNil(t, task.UpdatedAt)

	// Test with empty description
	_, err = store.CreateTask(ctx, "")
	assert.Error(t, err)

	// Test context cancellation
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	_, err = store.CreateTask(cancelCtx, "Should fail")
	assert.Error(t, err)
}

func testGetTask(t *testing.T, store TaskRepository) {
	ctx := context.Background()

	// Create a task first
	createdTask, err := store.CreateTask(ctx, "Get test task")
	require.NoError(t, err)

	// Test successful retrieval
	retrievedTask, err := store.GetTask(ctx, createdTask.Id)
	require.NoError(t, err)
	assert.Equal(t, createdTask.Id, retrievedTask.Id)
	assert.Equal(t, createdTask.Description, retrievedTask.Description)
	assert.Equal(t, createdTask.Completed, retrievedTask.Completed)

	// Test non-existent task
	_, err = store.GetTask(ctx, "99999")
	assert.Error(t, err)

	// Test invalid ID format
	_, err = store.GetTask(ctx, "invalid")
	assert.Error(t, err)
}

func testListTasks(t *testing.T, store TaskRepository) {
	ctx := context.Background()

	// Get initial count
	initialTasks, err := store.ListTasks(ctx)
	require.NoError(t, err)
	initialCount := len(initialTasks)

	// Create multiple tasks
	descriptions := []string{"Task 1", "Task 2", "Task 3"}
	for _, desc := range descriptions {
		_, err := store.CreateTask(ctx, desc)
		require.NoError(t, err)
	}

	// List all tasks
	tasks, err := store.ListTasks(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tasks), initialCount+3)

	// Verify tasks are ordered by created_at DESC (newest first)
	if len(tasks) >= 2 {
		assert.True(t, tasks[0].CreatedAt.AsTime().After(tasks[1].CreatedAt.AsTime()) ||
			tasks[0].CreatedAt.AsTime().Equal(tasks[1].CreatedAt.AsTime()))
	}
}

func testUpdateTask(t *testing.T, store TaskRepository) {
	ctx := context.Background()

	// Create a task first
	task, err := store.CreateTask(ctx, "Original description")
	require.NoError(t, err)

	// Add small delay to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Test updating description and completion status
	updatedTask, err := store.UpdateTask(ctx, task.Id, "Updated description", true)
	require.NoError(t, err)
	assert.Equal(t, task.Id, updatedTask.Id)
	assert.Equal(t, "Updated description", updatedTask.Description)
	assert.True(t, updatedTask.Completed)
	assert.True(t, updatedTask.UpdatedAt.AsTime().After(task.UpdatedAt.AsTime()))

	// Test updating only completion status
	updatedTask2, err := store.UpdateTask(ctx, task.Id, "", false)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", updatedTask2.Description) // Should remain unchanged
	assert.False(t, updatedTask2.Completed)

	// Test non-existent task
	_, err = store.UpdateTask(ctx, "99999", "Should fail", false)
	assert.Error(t, err)

	// Test invalid ID format
	_, err = store.UpdateTask(ctx, "invalid", "Should fail", false)
	assert.Error(t, err)
}

func testDeleteTask(t *testing.T, store TaskRepository) {
	ctx := context.Background()

	// Create a task first
	task, err := store.CreateTask(ctx, "Task to delete")
	require.NoError(t, err)

	// Test successful deletion
	err = store.DeleteTask(ctx, task.Id)
	require.NoError(t, err)

	// Verify task is deleted
	_, err = store.GetTask(ctx, task.Id)
	assert.Error(t, err)

	// Test deleting non-existent task
	err = store.DeleteTask(ctx, "99999")
	assert.Error(t, err)

	// Test invalid ID format
	err = store.DeleteTask(ctx, "invalid")
	assert.Error(t, err)
}

func testConcurrentOperations(t *testing.T, store TaskRepository) {
	ctx := context.Background()
	const numGoroutines = 10
	const tasksPerGoroutine = 5

	// Channel to collect errors
	errChan := make(chan error, numGoroutines*tasksPerGoroutine)
	done := make(chan bool, numGoroutines)

	// Start multiple goroutines creating tasks concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()
			
			for j := 0; j < tasksPerGoroutine; j++ {
				desc := fmt.Sprintf("Concurrent task G%d-T%d", goroutineID, j)
				task, err := store.CreateTask(ctx, desc)
				if err != nil {
					errChan <- err
					return
				}

				// Try to update the task
				_, err = store.UpdateTask(ctx, task.Id, desc+" UPDATED", j%2 == 0)
				if err != nil {
					errChan <- err
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check for errors
	close(errChan)
	for err := range errChan {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	// Verify we can still list tasks
	tasks, err := store.ListTasks(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tasks), numGoroutines*tasksPerGoroutine)
}

func TestMySQLTaskStore_Manager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := context.Background()

	// Start MariaDB container
	mariadbContainer, err := mariadb.Run(ctx,
		"mariadb:11.5",
		mariadb.WithDatabase("testdb"),
		mariadb.WithUsername("testuser"),
		mariadb.WithPassword("testpass"),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, mariadbContainer.Terminate(ctx))
	}()

	// Get connection details
	host, err := mariadbContainer.Host(ctx)
	require.NoError(t, err)

	port, err := mariadbContainer.MappedPort(ctx, "3306")
	require.NoError(t, err)

	// Create config
	cfg := &config.Config{
		Database: config.DatabaseConfig{
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
		},
	}

	// Test manager creation
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, manager.Close())
	}()

	// Test health check
	err = manager.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test task operations through manager
	taskStore := manager.TaskStore()
	task, err := taskStore.CreateTask(ctx, "Manager test task")
	require.NoError(t, err)
	assert.NotEmpty(t, task.Id)

	retrievedTask, err := taskStore.GetTask(ctx, task.Id)
	require.NoError(t, err)
	assert.Equal(t, task.Id, retrievedTask.Id)
}

// Benchmark tests
func BenchmarkMySQLTaskStore_CreateTask(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark tests in short mode")
	}

	ctx := context.Background()

	// Start MariaDB container
	mariadbContainer, err := mariadb.Run(ctx,
		"mariadb:11.5",
		mariadb.WithDatabase("benchdb"),
		mariadb.WithUsername("benchuser"),
		mariadb.WithPassword("benchpass"),
	)
	require.NoError(b, err)
	defer func() {
		assert.NoError(b, mariadbContainer.Terminate(ctx))
	}()

	// Get connection details
	host, err := mariadbContainer.Host(ctx)
	require.NoError(b, err)

	port, err := mariadbContainer.MappedPort(ctx, "3306")
	require.NoError(b, err)

	// Create database config
	dbConfig := &config.DatabaseConfig{
		Host:            host,
		Port:            port.Int(),
		User:            "benchuser",
		Password:        "benchpass",
		Database:        "benchdb",
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		SSLMode:         "false",
	}

	// Create store
	store, err := NewMySQLTaskStore(dbConfig)
	require.NoError(b, err)
	defer func() {
		assert.NoError(b, store.Close())
	}()

	b.ResetTimer()

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			desc := fmt.Sprintf("Benchmark task %d", i)
			_, err := store.CreateTask(ctx, desc)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				desc := fmt.Sprintf("Parallel benchmark task %d", i)
				_, err := store.CreateTask(ctx, desc)
				if err != nil {
					b.Fatal(err)
				}
				i++
			}
		})
	})
}